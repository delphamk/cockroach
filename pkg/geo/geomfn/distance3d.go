package geomfn

import (
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geodist"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/xyz"
)

func CustomMinDistance3D(a geo.Geometry, b geo.Geometry) (float64, error) {
	if a.SRID() != b.SRID() {
		return 0, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	}
	_ = geos.HasZ(a.EWKB())

	// a.AsGeomT()
	// aGeomRepr, err := a.AsGeomT()
	// if err != nil {
	// 	return 0, err
	// }

	// switch geomRepr.Layout() {
	// case geom.XYZ, geom.XYZM:
	// 	return length3DFromGeomT(geomRepr)
	// }

	return min3DDistanceInternal(a, b, 0, geo.EmptyBehaviorOmit, geo.FnInclusive)
}

func min3DDistanceInternal(
	a geo.Geometry,
	b geo.Geometry,
	stopAfter float64,
	emptyBehavior geo.EmptyBehavior,
	exclusivity geo.FnExclusivity,
) (float64, error) {

	u := newGeomMin3DDistanceUpdater(stopAfter, exclusivity)

	bbb := a.CartesianBoundingBox().Intersects(b.CartesianBoundingBox())
	bbb = false
	c := &geom3DDistanceCalculator{updater: u, boundingBoxIntersects: bbb}
	return distanceInternal3D(a, b, c, emptyBehavior)
}

func newGeomMin3DDistanceUpdater(
	stopAfter float64, exclusivity geo.FnExclusivity,
) *geomMin3DDistanceUpdater {
	return &geomMin3DDistanceUpdater{
		currentValue:        math.MaxFloat64,
		stopAfter:           stopAfter,
		exclusivity:         exclusivity,
		coordA:              nil,
		coordB:              nil,
		geometricalObjOrder: geometricalObjectsNotFlipped,
	}
}

type geom3DDistanceCalculator struct {
	updater               geodist.DistanceUpdater
	boundingBoxIntersects bool
}

var _ geodist.DistanceCalculator = (*geom3DDistanceCalculator)(nil)

// BoundingBoxIntersects implements geodist.DistanceCalculator.
func (g *geom3DDistanceCalculator) BoundingBoxIntersects() bool {
	return g.boundingBoxIntersects
}

func (g *geom3DDistanceCalculator) ClosestPointToPolygon(point geodist.Point, poly geodist.Polygon) (geodist.Point, bool) {

	pop, pv := DefinePlane(poly)

	projectedPoint := ProjectPointOnPlan(point, pop, pv)

	// Checking if the point projected on the plane of the polygon actually is inside that polygon.

	return projectedPoint, CheckPointOnPoly(point, pop, pv, poly)
}

// ClosestPointToEdge implements geodist.DistanceCalculator.
func (c *geom3DDistanceCalculator) ClosestPointToEdge(e geodist.Edge, p geodist.Point) (geodist.Point, bool) {

	// fmt.Printf("!!! p %v \n", p.GeomPoint)
	// fmt.Printf("!!! e.V0 %v \n", e.V0.GeomPoint)
	// fmt.Printf("!!! e.V1 %v \n", e.V1.GeomPoint)

	if coordEqual3D(e.V0.GeomPoint, e.V1.GeomPoint) {
		return e.V0, coordEqual3D(e.V0.GeomPoint, p.GeomPoint)
	}

	if coordEqual3D(p.GeomPoint, e.V0.GeomPoint) {
		return p, true
	}
	if coordEqual3D(p.GeomPoint, e.V1.GeomPoint) {
		return p, true
	}

	ac := coordSub3D(p.GeomPoint, e.V0.GeomPoint)
	ab := coordSub3D(e.V1.GeomPoint, e.V0.GeomPoint)

	r := coordDot3D(ac, ab) / coordNorm2_3D(ab)

	if r < 0 || r > 1 {
		return p, false
	}

	mul := coordMul3D(ab, r)
	res := coordAdd3D(e.V0.GeomPoint, mul)

	return geodist.Point{GeomPoint: res}, c.DistanceUpdater().Update(p, geodist.Point{GeomPoint: res})
}

// DistanceUpdater implements geodist.DistanceCalculator.
func (g *geom3DDistanceCalculator) DistanceUpdater() geodist.DistanceUpdater {
	return g.updater
}

// NewEdgeCrosser implements geodist.DistanceCalculator.
func (g *geom3DDistanceCalculator) NewEdgeCrosser(edge geodist.Edge, startPoint geodist.Point) geodist.EdgeCrosser {
	// BoundingBoxIntersects: if the bounding box of the two shapes do not intersect,
	// then we don't need to check whether edges intersect either.
	panic("unimplemented")
}

// PointIntersectsLinearRing implements geodist.DistanceCalculator.
func (g *geom3DDistanceCalculator) PointIntersectsLinearRing(point geodist.Point, linearRing geodist.LinearRing) bool {
	x := findPointSideOfLinearRing3D(point, linearRing)

	switch x {
	case insideLinearRing, onLinearRing:
		return true
	default:
		return false
	}
}

type geomMin3DDistanceUpdater struct {
	currentValue float64
	stopAfter    float64

	exclusivity geo.FnExclusivity
	// coordA represents the first vertex of the edge that holds the maximum distance.
	coordA geom.Coord
	// coordB represents the second vertex of the edge that holds the maximum distance.
	coordB geom.Coord

	geometricalObjOrder geometricalObjectsOrder
}

var _ geodist.DistanceUpdater = (*geomMin3DDistanceUpdater)(nil)

// Distance implements geodist.DistanceUpdater.
func (u *geomMin3DDistanceUpdater) Distance() float64 {
	return u.currentValue
}

// FlipGeometries implements geodist.DistanceUpdater.
func (u *geomMin3DDistanceUpdater) FlipGeometries() {
	u.geometricalObjOrder = -u.geometricalObjOrder
}

// IsMaxDistance implements geodist.DistanceUpdater.
func (u *geomMin3DDistanceUpdater) IsMaxDistance() bool {
	return false
}

// OnIntersects implements geodist.DistanceUpdater.
func (u *geomMin3DDistanceUpdater) OnIntersects(p geodist.Point) bool {
	u.coordA = p.GeomPoint
	u.coordB = p.GeomPoint
	u.currentValue = 0
	return true
}

// Update implements geodist.DistanceUpdater.
func (u *geomMin3DDistanceUpdater) Update(aPoint geodist.Point, bPoint geodist.Point) bool { // @D
	a := aPoint.GeomPoint
	b := bPoint.GeomPoint

	var dist float64
	if len(a) >= 3 && len(b) >= 3 {
		dist = xyz.Distance(a, b)
	} else {
		panic("3d distance update in 2d")
	}
	// fmt.Printf(">>> a %v \n", a)
	// fmt.Printf(">>> b %v \n", b)
	// fmt.Printf(">>> dist %v\n", dist)

	if dist < u.currentValue || u.coordA == nil {
		u.currentValue = dist
		if u.geometricalObjOrder == geometricalObjectsFlipped {
			u.coordA = b
			u.coordB = a
		} else {
			u.coordA = a
			u.coordB = b
		}
		if u.exclusivity == geo.FnExclusive {
			return dist < u.stopAfter
		}
		return dist <= u.stopAfter
	}
	return false
}

// findPointSideOfLinearRing returns whether a point is outside, on, or inside a
// linear ring.
func findPointSideOfLinearRing3D(point geodist.Point, linearRing geodist.LinearRing) linearRingSide {


	// panic("this is used only on to Polygon")
	windingNumber := 0
	p := point.GeomPoint
	for edgeIdx, numEdges := 0, linearRing.NumEdges(); edgeIdx < numEdges; edgeIdx++ {
		e := linearRing.Edge(edgeIdx)
		eV0 := e.V0.GeomPoint
		eV1 := e.V1.GeomPoint
		// Same vertex; none of these checks will pass.
		if coordEqual3D(eV0, eV1) {
			continue
		}
		yMin := math.Min(eV0.Y(), eV1.Y())
		yMax := math.Max(eV0.Y(), eV1.Y())
		// If the edge isn't on the same level as Y, this edge isn't worth considering.
		if p.Y() > yMax || p.Y() < yMin {
			continue
		}
		side := findPointSide(p, eV0, eV1)
		// If the point is on the line if the edge was infinite, and the point is within the bounds
		// of the line segment denoted by the edge, there is a covering.
		if side == pointSideOn && (eV0.X() <= p.X() && p.X() <= eV1.X()) {
			return onLinearRing
		}
		// If the point is left of the segment and the line is rising
		// we have a circle going CCW, so increment.
		// Note we only compare [start, end) as we do not want to double count points
		// which are on the same X / Y axis as an edge vertex.
		if side == pointSideLeft && eV0.Y() <= p.Y() && p.Y() < eV1.Y() {
			windingNumber++
		}
		// If the line is to the right of the segment and the
		// line is falling, we a have a circle going CW so decrement.
		// Note we only compare [start, end) as we do not want to double count points
		// which are on the same X / Y axis as an edge vertex.
		if side == pointSideRight && eV1.Y() <= p.Y() && p.Y() < eV0.Y() {
			windingNumber--
		}
	}
	if windingNumber != 0 {
		return insideLinearRing
	}
	return outsideLinearRing
}

func distanceInternal3D(
	aGeo geo.Geometry, bGeo geo.Geometry, distCalc geodist.DistanceCalculator, emptyBehavior geo.EmptyBehavior,
) (float64, error) {
	// If either side has no geoms, then we error out regardless of emptyBehavior.
	if aGeo.Empty() || bGeo.Empty() {
		return 0, geo.NewEmptyGeometryError()
	}

	aGeomT, err := aGeo.AsGeomT()
	if err != nil {
		return 0, err
	}
	bGeomT, err := bGeo.AsGeomT()
	if err != nil {
		return 0, err
	}
	// If we early exit, we have to check empty behavior upfront to return
	// the appropriate error message.
	// This matches PostGIS's behavior for DWithin, which is always false
	// if at least one element is empty.
	if emptyBehavior == geo.EmptyBehaviorError &&
		(geo.GeomTContainsEmpty(aGeomT) || geo.GeomTContainsEmpty(bGeomT)) {
		return 0, geo.NewEmptyGeometryError()
	}

	a_GeoIt := geo.NewGeomTIterator(aGeomT, emptyBehavior)
	aGeom, aNext, aErr := a_GeoIt.Next()
	if aErr != nil {
		return 0, aErr
	}
	for aNext {
		a_Geodist, err := geomToGeodist(aGeom)

		if err != nil {
			return 0, err
		}

		b_GeoIt := geo.NewGeomTIterator(bGeomT, emptyBehavior)
		bGeom, bNext, bErr := b_GeoIt.Next()
		if bErr != nil {
			return 0, bErr
		}
		for bNext {
			b_Geodist, err := geomToGeodist(bGeom)
			if err != nil {
				return 0, err
			}
			// fmt.Printf(">>> a_Geodist %v \n", a_Geodist)
			// fmt.Printf(">>> b_Geodist %v\n", b_Geodist)
			earlyExit, err := geodist.ShapeDistance3D(distCalc, a_Geodist, b_Geodist)
			if err != nil {
				return 0, err
			}
			// earlyExit = false
			if earlyExit {
				// fmt.Printf(">>> earlyExit %v dist %v\n", earlyExit, distCalc.DistanceUpdater().Distance())
				return distCalc.DistanceUpdater().Distance(), nil
			}

			bGeom, bNext, bErr = b_GeoIt.Next()
			if bErr != nil {
				return 0, bErr
			}
		}

		aGeom, aNext, aErr = a_GeoIt.Next()
		if aErr != nil {
			return 0, aErr
		}
	}
	// fmt.Printf(">>> FINAL %v\n", distCalc.DistanceUpdater().Distance())

	return distCalc.DistanceUpdater().Distance(), nil
}

func DefinePlane(poly geodist.Polygon) (geodist.Point, geodist.Point) {
	lineRing := poly.LinearRing(0)

	unique_points := lineRing.NumVertexes() - 1

	if unique_points < 3 {
		panic("less than 3 points")
	}

	pop := geom.Coord{0, 0, 0}
	pv := geom.Coord{0, 0, 0}

	for i := 0; i < unique_points; i++ {
		vertex := lineRing.Vertex(i).GeomPoint
		pop = coordAdd3D(pop, vertex)
	}

	pop = coordMul3D(pop, float64(1)/float64(unique_points))

	POL_BREAKS := 3
	for i := 0; i < POL_BREAKS; i++ {

		// this could be buggy.
		index1 := i * unique_points / POL_BREAKS
		index2 := index1 + unique_points/POL_BREAKS

		if index1 == index2 {
			panic("check this")
			// continue
		}

		p1 := lineRing.Vertex(index1).GeomPoint
		p2 := lineRing.Vertex(index2).GeomPoint

		v1 := coordSub3D(p1, pop)
		v2 := coordSub3D(p2, pop)

		cross := coordCross3D(v1, v2)

		norm := coordNorm2_3D(cross)

		mul := coordMul3D(cross, 1/norm)

		pv = coordAdd3D(pv, mul)

	}

	return geodist.Point{GeomPoint: pop}, geodist.Point{GeomPoint: pv}

}

func ProjectPointOnPlan(point geodist.Point, pop geodist.Point, pv geodist.Point) geodist.Point {
	v1 := coordSub3D(point.GeomPoint, pop.GeomPoint)
	f := coordDot3D(pv.GeomPoint, v1)
	if f == 0 {
		return point
	}

	f = -f / coordDot3D(pv.GeomPoint, pv.GeomPoint)

	mul := coordMul3D(pv.GeomPoint, f)

	projectedPoint := coordAdd3D(point.GeomPoint, mul)
	fmt.Printf(">>>  projectedPoint %v\n", projectedPoint)

	return geodist.Point{GeomPoint: projectedPoint}
}

func CheckPointOnPoly(point geodist.Point, pop geodist.Point, pvector geodist.Point, poly geodist.Polygon) bool {

	linering := poly.LinearRing(0)

	cn := 0

	p := point.GeomPoint
	pv := pvector.GeomPoint
	v1 := linering.Vertex(0).GeomPoint

	if (pv[2] >= pv.X()) && pv[2] >= pv.Y() {
		// If the z vector of the normal vector to the plane is larger than x and y vector we project the ring to the xy-plane

		for i := 1; i < linering.NumVertexes(); i++ {

			v2 := linering.Vertex(i).GeomPoint

			if (v1.Y() <= p.Y() && v2.Y() > p.Y()) /* an upward crossing */ ||
				(v1.Y() > p.Y() && v2.Y() <= p.Y() /* a downward crossing */) {


				vt := (p.Y() - v1.Y()) / (v2.Y() - v1.Y())

				/* P.x <intersect */
				val := v1.X() + vt*(v2.X()-v1.X())
				if p.X() < val {

					/* a valid crossing of y=p.y right of p.x */
					cn++
				}
			}
			v1 = v2
		}
	} else if (pv.Y() >= pv.X()) && pv.Y() >= pv[2] {

		for i := 1; i < linering.NumVertexes(); i++ {
			v2 := linering.Vertex(i).GeomPoint

			if (v1[2] <= p[2] && v2[2] > p[2]) /* an upward crossing */ ||
				(v1[2] > p[2] && v2[2] <= p[2] /* a downward crossing */) {

				vt := (p[2] - v1[2]) / (v2[2] - v1[2])
				/* P.x <intersect */
				if p.X() < v1.X()+vt*(v2.X()-v1.X()) {
					/* a valid crossing of y=p.y right of p.x */
					cn++
				}
			}
			v1 = v2
		}
	} else {

		for i := 1; i < linering.NumVertexes(); i++ {
			v2 := linering.Vertex(i).GeomPoint

			if (v1[2] <= p[2] && v2[2] > p[2]) /* an upward crossing */ ||
				(v1[2] > p[2] && v2[2] <= p[2] /* a downward crossing */) {

				vt := (p[2] - v1[2]) / (v2[2] - v1[2])
				/* P.x <intersect */
				if p[2] < v1[2]+vt*(v2[2]-v1[2]) {
					/* a valid crossing of y=p.y right of p.x */
					cn++
				}
			}
			v1 = v2
		}
	}


	return cn%2 == 1
}

// ClosestEdgeToEdge implements geodist.DistanceCalculator.
func (g *geom3DDistanceCalculator) ClosestEdgeToEdge(edge1 geodist.Edge, edge2 geodist.Edge) (geodist.Point, geodist.Point, bool) {
	start1 := edge1.V0.GeomPoint
	end1 := edge1.V1.GeomPoint

	start2 := edge2.V0.GeomPoint
	end2 := edge2.V1.GeomPoint

	pointToEdge := func(p geodist.Point, e geodist.Edge) (geodist.Point, geodist.Point, bool) {
		ret, res := g.ClosestPointToEdge(e, p)
		return p, ret, res
	}

	if coordEqual3D(start1, end1) {
		panic("need testcase")
		// return pointToEdge(edge1.V0, edge2)

	}

	if coordEqual3D(start2, end2) {
		panic("need testcase")
		// return pointToEdge(edge2.V0, edge1)
		// ret, res := c.ClosestPointToEdge(edge1, edge2.V0)
		// return edge2.V0, ret, res
	}

	if coordEqual3D(start1, start2) {
		panic("need testcase")
		// return edge1.V0, edge2.V0, true
		// return pointToEdge( edge2.V0,edge1)
		// distance should be 0
	}

	v1 := coordSub3D(end1, start1)
	v2 := coordSub3D(end2, start2)
	vl := coordSub3D(start1, start2)

	a := coordDot3D(v1, v1)
	b := coordDot3D(v1, v2)
	c := coordDot3D(v2, v2)
	d := coordDot3D(v1, vl)
	e := coordDot3D(v2, vl)
	D := a*c - b*b

	var s1k, s2k float64

	if D < 0.000000001 { /* the lines are almost parallel*/
		s1k = 0.0
		/*If the lines are parallel we try by using the startpoint of first segment. If that gives a
		  projected point on the second line outside segment 2 it will be found that s2k is >1 or <0.*/

		if b > c { /* use the largest denominator*/
			s2k = d / b
		} else {
			s2k = e / c
		}

	} else {
		s1k = (b*e - c*d) / D
		s2k = (a*e - b*d) / D
	}

	// fmt.Printf(" \n")
	// fmt.Printf(">>> v1 %v \n", v1)
	// fmt.Printf(">>> v2 %v \n", v2)
	// fmt.Printf(">>> vl %v \n", vl)
	// fmt.Printf(">>> a %v \n", a)
	// fmt.Printf(">>> b %v \n", b)
	// fmt.Printf(">>> c %v \n", c)
	// fmt.Printf(">>> d %v \n", d)
	// fmt.Printf(">>> e %v \n", e)
	// fmt.Printf(">>> s1k %v \n", s1k)
	// fmt.Printf(">>> s2k %v \n", s2k)
	// fmt.Printf(" \n")

	/* Now we check if the projected closest point on the infinite lines is outside our segments. If so the
	 * combinations with start and end points will be tested*/

	if (s1k <= 0.0 || s1k >= 1.0 || s2k <= 0.0 || s2k >= 1.0) && true {

		// this will check a vertex against edge of points..
		if s1k <= 0.0 {
			if ret1, ret2, ok := pointToEdge(edge1.V0, edge2); ok {
				return ret1, ret2, ok
			}
		}
		if s1k >= 1.0 {
			if ret1, ret2, ok := pointToEdge(edge1.V1, edge2); ok {
				return ret1, ret2, ok
			}
		}
		if s2k <= 0.0 {
			// dl->twisted = ((dl->twisted) * (-1));

			if ret1, ret2, ok := pointToEdge(edge2.V0, edge1); ok {
				return ret1, ret2, ok
			}
		}
		if s2k >= 1.0 {
			// dl->twisted = ((dl->twisted) * (-1));
			if ret1, ret2, ok := pointToEdge(edge2.V1, edge1); ok {
				return ret1, ret2, ok
			}
		}
	} else { /*Find the closest point on the edges of both segments*/
		// (s1p2->x - s1p1->x)
		p1 := v1
		p1 = coordMul3D(p1, s1k)
		p1 = coordAdd3D(start1, p1)
		// start1 + (v1 * s1k )
		// p1.x = s1p1->x + s1k * (s1p2->x - s1p1->x);
		// p1.y = s1p1->y + s1k * (s1p2->y - s1p1->y);
		// p1.z = s1p1->z + s1k * (s1p2->z - s1p1->z);

		p2 := v2
		p2 = coordMul3D(p2, s2k)
		p2 = coordAdd3D(start2, p2)
		// p2.x = s2p1->x + s2k * (s2p2->x - s2p1->x);
		// p2.y = s2p1->y + s2k * (s2p2->y - s2p1->y);
		// p2.z = s2p1->z + s2k * (s2p2->z - s2p1->z);

		// if !lw_dist3d_pt_pt(&p1, &p2, dl) {
		// 	/* Send the closest points to point-point calculation*/
		// 	return LW_FALSE
		// }
		res1 := geodist.Point{GeomPoint: p1}
		res2 := geodist.Point{GeomPoint: p2}
		return res1, res2, g.DistanceUpdater().Update(res1, res2)
	}
	// return false: we should not calc the distance between the points.
	return geodist.Point{}, geodist.Point{}, false

}
