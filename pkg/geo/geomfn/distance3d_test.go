// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package geomfn

import (
	"fmt"
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

func TestCustomMinDistance3d(t *testing.T) {

	// default_accepted_error := 0.00001
	zero_accepted_error := 0.0

	// point
	// customDist3dTest(t, point0, point1, 1, zero_accepted_error)
	// customDist3dTest(t, point0, linestring1, 1, zero_accepted_error)
	customDist3dTest(t, point0, polygon1, 1.0, zero_accepted_error)
	// customDist3dTest(t, point0, "POLYGON((-10 -10 1, 10 -10 1, 10 10 1, -10 10 1, -15 5 1, -10 -10 1))", 1.0, zero_accepted_error)

	// some random far distance
	// customDist3dTest(t, "POINT(0 0 0)", "POLYGON((-10 -10 10, 10 -10 10, 10 10 20, -10 10 20, -10 -10 10))", 13.416407864998737, zero_accepted_error)

	// point is on the plane but projection is not on poly. It should not be 0
	// customDist3dTest(t, "POINT(0 100 1)", polygon1, 90, zero_accepted_error)

	//linetype
	// customDist3dTest(t, linestring0, point1, 1, zero_accepted_error)
	// customDist3dTest(t, linestring0, linestring1, 1, zero_accepted_error)
	// customDist3dTest(t, linestring0, polygon1, 1.0, zero_accepted_error)

	//polygon
	// customDist3dTest(t, polygon0, point1, 1, zero_accepted_error)
	// customDist3dTest(t, polygon0, linestring1, 1, zero_accepted_error)
	// customDist3dTest(t, polygon0, polygon1, 1, zero_accepted_error)

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
		err = fmt.Errorf("failed dist3dTest. \n\n\t\t [ \"%v\" != \"%v\"] (%v)", dist, ans, tolerance)
		fmt.Printf(">>> 1: %s \n", geo1)
		fmt.Printf(">>> 2: %s \n", geo2)
		fmt.Printf(">>> ERROR %s \n", err)
		t.Error(err)
	}
}

func coordEqual3ToleranceD(a geom.Coord, b geom.Coord, tolerance float64) bool {
	return inTolerance(a.X(), b.X(), tolerance) && inTolerance(a.Y(), b.Y(), tolerance) && inTolerance(a[2], b[2], tolerance)
}

func TestDefinePlane(t *testing.T) {

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

func TestProjectPointOnPlan(t *testing.T) {

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

func TestCheckPointOnPoly(t *testing.T) {

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
