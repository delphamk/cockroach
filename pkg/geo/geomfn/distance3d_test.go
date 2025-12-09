// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package geomfn

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geodist"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-geom"
)

var point0 = "POINT(0 0 0)"
var point1 = "POINT(0 0 1)"
var linestring0 = "LINESTRING(-10 0 0, 10 0 0)"
var linestring1 = "LINESTRING(-10 0 1, 10 0 1)"
var polygon0 = "POLYGON((-10 -10 0, 10 -10 0, 10 10 0, -10 10 0, -10 -10 0))"
var polygon1 = "POLYGON((-10 -10 1, 10 -10 1, 10 10 1, -10 10 1, -10 -10 1))"
var _ = point0
var _ = point1
var _ = linestring0
var _ = linestring1
var _ = polygon0
var _ = polygon1

var tolerance = 0.001

func polygonFromRect(x1, y1, x2, y2, z float64) string {
	// if x1 > x2 {
	// 	x1, x2 = x2, x1
	// }
	// if y1 > y2 {
	// 	y1, y2 = y2, y1
	// }
	points := []string{
		fmt.Sprintf("%.6g %.6g %g", x1, y1, z),
		fmt.Sprintf("%.6g %.6g %g", x2, y1, z),
		fmt.Sprintf("%.6g %.6g %g", x2, y2, z),
		fmt.Sprintf("%.6g %.6g %g", x1, y2, z),
		fmt.Sprintf("%.6g %.6g %g", x1, y1, z), // closing point
	}

	return "POLYGON((" + fmt.Sprintf("%s", strings.Join(points, ", ")) + "))"
}

func Test3DCustomMinDistance3d(t *testing.T) {

	// default_accepted_error := 0.00001
	zero_accepted_error := 0.0

	// TODO failing. line is on poly
	customDist3dTest(t, "LINESTRING(1 0 0, 0 0 0)", "POLYGON((-10 0 -10, 10 0 -10, 10 0 10, -10 0 10, -10 0 -10))", 0, zero_accepted_error)

	if true {
		return
	}

	// point
	customDist3dTest(t, point0, point1, 1, zero_accepted_error)
	customDist3dTest(t, point0, linestring1, 1, zero_accepted_error)
	customDist3dTest(t, point0, polygon1, 1.0, zero_accepted_error)
	customDist3dTest(t, point0, "POLYGON((-10 -10 1, 10 -10 1, 10 10 1, -10 10 1, -15 5 1, -10 -10 1))", 1.0, zero_accepted_error)

	// some random far distance
	customDist3dTest(t, "POINT(0 0 0)", "POLYGON((-10 -10 10, 10 -10 10, 10 10 20, -10 10 20, -10 -10 10))", 13.416407864998737, zero_accepted_error)

	// point is on the plane but projection is not on poly. It should not be 0
	customDist3dTest(t, "POINT(0 100 1)", polygon1, 90, zero_accepted_error)

	// linetype
	customDist3dTest(t, linestring0, point1, 1, zero_accepted_error)
	customDist3dTest(t, linestring0, linestring1, 1, zero_accepted_error)
	customDist3dTest(t, linestring0, polygon1, 1, zero_accepted_error)

	// line cross not on vertex
	customDist3dTest(t, "LINESTRING(-10 10 0, 10 -10 0)", "LINESTRING(-10 -10 0, 10 10 0)", 0, zero_accepted_error)
	customDist3dTest(t, "LINESTRING(0 -10 10, 0 10 -10)", "LINESTRING(0 -10 -10, 0 10 10)", 0, zero_accepted_error)

	// line closest to poly edge
	customDist3dTest(t, "LINESTRING(-100 0 0, -101 0 0)", polygon1, 90.00555538409837, zero_accepted_error)
	customDist3dTest(t, "LINESTRING(-100 0 1, -101 0 1)", polygon1, 90, zero_accepted_error)
	customDist3dTest(t, "LINESTRING(-1000 -1000 1, 1000 -1000 1)", polygon1, 990.0, zero_accepted_error)

	// line intersect poly
	customDist3dTest(t, "LINESTRING(0 0 -15, 0 0 15)", "POLYGON((-10 0 -10, 10 0 -10, 10 0 10, -10 0 10, -10 0 -10))", 0, zero_accepted_error)

	customDist3dTest(t, linestring0, polygon0, 0, zero_accepted_error)

	//polygon
	customDist3dTest(t, polygon0, point1, 1, zero_accepted_error)
	customDist3dTest(t, polygon0, linestring1, 1, zero_accepted_error)
	customDist3dTest(t, polygon0, polygon1, 1, zero_accepted_error)

	customDist3dTest(t,
		polygonFromRect(1, 1, 2, 2, 0),
		polygonFromRect(3, 3, 4, 4, 0),
		1.4142135623730951, zero_accepted_error)

	customDist3dTest(t,
		polygonFromRect(1, 1, 2, 2, 0),
		polygonFromRect(3, 3, 4, 4, 5),
		5.196152422706632, zero_accepted_error)

	customDist3dTest(t,
		polygonFromRect(1, 1, 10, 10, 0),
		polygonFromRect(10, 10, 1, 1, 5),
		5, zero_accepted_error)

	customDist3dTest(t,
		polygonFromRect(1, 1, 10, 10, 0),
		polygonFromRect(10, 10, 1, 1, 0),
		0, zero_accepted_error)

	// customDist3dTest(t,
	// 	"POLYGON((-10 0 -10, 10 0 -10, 10 0 10, -10 0 10, -10 0 -10))",
	// 	"POLYGON((-10 -10 0, 10 -10 0, 10 10 0, -10 10 0, -10 -10 0))",
	// 	0, zero_accepted_error)

}

func customDist3dTest(t *testing.T, geo1 string, geo2 string, ans float64, tolerance float64) {
	a, err := geo.ParseGeometry(geo1)
	if err != nil {
		fmt.Printf(">>> err: %s\n", err)
		return
	}
	require.NoError(t, err)
	b, err := geo.ParseGeometry(geo2)
	if err != nil {
		fmt.Printf(">>> err: %s\n", err)
		return
	}
	require.NoError(t, err)

	dist, err := CustomMinDistance3D(a, b)
	require.NoError(t, err)

	valid := inTolerance(dist, ans, tolerance)
	if !valid {
		err = fmt.Errorf("failed dist3dTest. \n\n\t\tGOT: %v \n\t\tEXPECTED:%v", dist, ans)
		fmt.Printf(">>> 1: %s \n", geo1)
		fmt.Printf(">>> 2: %s \n", geo2)
		fmt.Printf(">>> ERROR %s \n", err)
		t.Error(err)
	}
}

func coordEqual3ToleranceD(a geom.Coord, b geom.Coord, tolerance float64) bool {
	return inTolerance(a.X(), b.X(), tolerance) && inTolerance(a.Y(), b.Y(), tolerance) && inTolerance(a[2], b[2], tolerance)
}

func Test3DDefinePlane(t *testing.T) {

	testCases := []struct {
		poly string
		ePop geom.Coord
		ePv  geom.Coord
	}{
		{polygon0, geom.Coord{0, 0, 0}, geom.Coord{0, 0, float64(0.015)}},
		{polygon1, geom.Coord{0, 0, 1}, geom.Coord{0, 0, float64(0.015)}},
	}

	for _, tc := range testCases {
		poly, err := geo.ParseGeometry(tc.poly)
		require.NoError(t, err)

		polyGeomT, err := poly.AsGeomT()
		require.NoError(t, err)

		poly_Geodist, err := geomToGeodist(polyGeomT)
		require.NoError(t, err)

		var pop geodist.Point
		var pv geodist.Point
		switch a := poly_Geodist.(type) {
		case geodist.Polygon:
			pop, pv = DefinePlane(a)
		default:
			t.Error("not a polygon")
		}

		fmt.Printf(">>> pop %v \n", pop)
		fmt.Printf(">>> pv %v \n", pv)

		require.Truef(t, coordEqual3D(tc.ePop, pop.GeomPoint), "pop not equal [%v](%v)", tc.ePop, pop)
		require.Truef(t, coordEqual3ToleranceD(tc.ePv, pv.GeomPoint, tolerance), "pv not equal [%v](%v)", tc.ePv, pv)
	}

}

func Test3DProjectPointOnPlan(t *testing.T) {

	testCases := []struct {
		point  geom.Coord
		pop    geom.Coord
		pv     geom.Coord
		ePoint geom.Coord
	}{
		{geom.Coord{0, 0, 0}, geom.Coord{0, 0, 0}, geom.Coord{0, 0, 0.015}, geom.Coord{0, 0, 0}},
		{geom.Coord{1, 2, 5}, geom.Coord{0, 0, 0}, geom.Coord{0, 0, 0.015}, geom.Coord{1, 2, 0}},
		{geom.Coord{-3, 4, -2}, geom.Coord{0, 0, 0}, geom.Coord{0, 0, 0.015}, geom.Coord{-3, 4, 0}},
		{geom.Coord{2, 3, 4}, geom.Coord{0, 0, 1}, geom.Coord{0, 0, 0.015}, geom.Coord{2, 3, 1}},
		{geom.Coord{10, -10, 100}, geom.Coord{0, 0, 0}, geom.Coord{0, 0, 1}, geom.Coord{10, -10, 0}},
	}

	for i, tc := range testCases {
		point := geodist.Point{GeomPoint: tc.point}
		pop := geodist.Point{GeomPoint: tc.pop}
		pv := geodist.Point{GeomPoint: tc.pv}
		projectedPoint := ProjectPointOnPlan(point, pop, pv)

		fmt.Printf(">>> projectedPoint %v \n", projectedPoint)
		require.Truef(t, coordEqual3ToleranceD(tc.ePoint, projectedPoint.GeomPoint, tolerance), "[%v] projectedPoint not equal [%v](%v)", i, tc.ePoint, projectedPoint)
	}

}

func Test3DCheckPointOnPoly(t *testing.T) {

	// The point must exist on the plane.

	testCases := []struct {
		point  geom.Coord //  geom.Coord{0, 0, 1}
		pop    geom.Coord //  geom.Coord{0, 0, 1}
		pv     geom.Coord // geom.Coord{0, 0, float64(0.015)}
		poly   string     //polygon0
		inside bool       // true
	}{
		{geom.Coord{0, 0, 1}, geom.Coord{0, 0, 1}, geom.Coord{0, 0, float64(0.015)}, polygon1, true},
		{geom.Coord{10.1, 0, 1}, geom.Coord{0, 0, 1}, geom.Coord{0, 0, float64(0.015)}, polygon1, false},
		{geom.Coord{0, 10.1, 1}, geom.Coord{0, 0, 1}, geom.Coord{0, 0, float64(0.015)}, polygon1, false},
		// {geom.Coord{10.1, 0, 1}, geom.Coord{0, 0, 1}, geom.Coord{0, 0, float64(0.015)}, polygon1, false},
	}

	for i, tc := range testCases {
		fmt.Printf(">>> %v %v \n", i, tc)

		polyS, err := geo.ParseGeometry(tc.poly)
		require.NoError(t, err)
		polyGeomT, err := polyS.AsGeomT()
		require.NoError(t, err)
		poly_Geodist, err := geomToGeodist(polyGeomT)
		require.NoError(t, err)

		var poly geodist.Polygon
		switch a := poly_Geodist.(type) {
		case geodist.Polygon:
			poly = a
		default:
			t.Error("not a polygon")
		}

		point := geodist.Point{GeomPoint: tc.point}
		pop := geodist.Point{GeomPoint: tc.pop}
		pv := geodist.Point{GeomPoint: tc.pv}

		valid := CheckPointOnPoly(point, pop, pv, poly)

		fmt.Printf(">>> pointOnPoly: %v \n", valid)
		require.Truef(t, valid == tc.inside, " result is not correct. got: %v expected: %v", valid, tc.inside)

	}

}

func Test3DBoundingBox3D(t *testing.T) {

	a, err := geo.ParseGeometry("LINESTRING(-10 10 10, 10 -10 10)")
	if err != nil {
		fmt.Printf(">>> err: %s\n", err)
		return
	}
	require.NoError(t, err)

	require.NoError(t, err)
	a.CartesianBoundingBox()
	fmt.Printf(">>> a BB: %v \n", a.SpatialObject().BoundingBox)

	// bbb := a.CartesianBoundingBox().Intersects(b.CartesianBoundingBox())
	// c := &geom3DDistanceCalculator{updater: u, boundingBoxIntersects: bbb}

}

func Test3DClosestEdgeToEdge(t *testing.T) {

	testCases := []struct {
		start1 geom.Coord
		end1   geom.Coord
		start2 geom.Coord
		end2   geom.Coord
		dist   float64
		valid  bool
	}{
		{geom.Coord{-10, 0, 1}, geom.Coord{10, 0, 1}, geom.Coord{-10, 0, 0}, geom.Coord{10, 0, 0}, 1, true},
		{geom.Coord{-10, 10, 0}, geom.Coord{10, -10, 0}, geom.Coord{-10, -10, 0}, geom.Coord{10, 10, 0}, 0, true},
		{geom.Coord{-10, 10, 0}, geom.Coord{10, -10, 0}, geom.Coord{-10, -10, 1}, geom.Coord{10, 10, 1}, 1, true},
	}

	for i, tc := range testCases {
		fmt.Printf(">>> %v %v \n", i, tc)

		start1 := geodist.Point{GeomPoint: tc.start1}
		end1 := geodist.Point{GeomPoint: tc.end1}
		start2 := geodist.Point{GeomPoint: tc.start2}
		end2 := geodist.Point{GeomPoint: tc.end2}

		toString := func(p geodist.Point) string {
			return fmt.Sprintf("%v %v %v", p.GeomPoint.X(), p.GeomPoint.Y(), p.GeomPoint[2])
		}

		x := fmt.Sprintf("select ST_3DDistance( 'LINESTRING(%v, %v)'::geometry, 'LINESTRING(%v, %v)'::geometry );", toString(start1), toString(end1), toString(start2), toString(end2))
		fmt.Printf(">>> \n\n%s\n\n \n", x)

		u := newGeomMin3DDistanceUpdater(0, geo.FnInclusive)
		c := &geom3DDistanceCalculator{updater: u}
		e1 := geodist.Edge{
			V0: start1,
			V1: end1,
		}
		e2 := geodist.Edge{
			V0: start2,
			V1: end2,
		}

		closest1, closest2, valid := c.ClosestEdgeToEdge(e1, e2)

		if valid {
			c.updater.Update(closest1, closest2)
		}

		fmt.Printf(">>> c1 %v \n", closest1.GeomPoint)
		fmt.Printf(">>> c1 %v \n", closest2.GeomPoint)
		fmt.Printf(">>> valid %v \n", valid)

		fmt.Printf(">>> dist %v \n", c.updater.Distance())

		// require.Equalf(t, tc.valid, valid, "valid not equal")
		require.Equalf(t, tc.dist, c.updater.Distance(), "dist not equal")
	}

}
