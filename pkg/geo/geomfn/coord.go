// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package geomfn

import (
	"math"

	"github.com/twpayne/go-geom"
)

// coordAdd adds two coordinates and returns a new result.
func coordAdd(a geom.Coord, b geom.Coord) geom.Coord {
	return geom.Coord{a.X() + b.X(), a.Y() + b.Y()}
}

// coordSub subtracts two coordinates and returns a new result.
func coordSub(a geom.Coord, b geom.Coord) geom.Coord {
	return geom.Coord{a.X() - b.X(), a.Y() - b.Y()}
}

// coordMul multiplies a coord by a scalar and returns the new result.
func coordMul(a geom.Coord, s float64) geom.Coord {
	return geom.Coord{a.X() * s, a.Y() * s}
}

// coordDet returns the determinant of the 2x2 matrix formed by the vectors a and b.
func coordDet(a geom.Coord, b geom.Coord) float64 {
	return a.X()*b.Y() - b.X()*a.Y()
}

// coordDot returns the dot product of two coords if the coord was a vector.
func coordDot(a geom.Coord, b geom.Coord) float64 {
	return a.X()*b.X() + a.Y()*b.Y()
}

// coordCross returns the cross product of two coords if the coord was a vector.
func coordCross(a geom.Coord, b geom.Coord) float64 {
	return a.X()*b.Y() - a.Y()*b.X()
}

// coordNorm2 returns the normalization^2 of a coordinate if the coord was a vector.
func coordNorm2(c geom.Coord) float64 {
	return coordDot(c, c)
}

// coordNorm returns the normalization of a coordinate if the coord was a vector.
func coordNorm(c geom.Coord) float64 {
	return math.Sqrt(coordNorm2(c))
}

// coordEqual returns whether two coordinates are equal.
func coordEqual(a geom.Coord, b geom.Coord) bool {
	return a.X() == b.X() && a.Y() == b.Y()
}

// coordMag2 returns the magnitude^2 of a coordinate if the coord was a vector.
func coordMag2(c geom.Coord) float64 {
	return coordDot(c, c)
}

// coordAdd3D adds two 3D coordinates and returns a new result.
func coordAdd3D(a geom.Coord, b geom.Coord) geom.Coord {
	return geom.Coord{a.X() + b.X(), a.Y() + b.Y(), a[2] + b[2]}
}

// coordSub3D subtracts two 3D coordinates and returns a new result.
func coordSub3D(a geom.Coord, b geom.Coord) geom.Coord {
	return geom.Coord{a.X() - b.X(), a.Y() - b.Y(), a[2] - b[2]}
}

// coordMul3D multiplies a 3D coord by a scalar and returns the new result.
func coordMul3D(a geom.Coord, s float64) geom.Coord {
	return geom.Coord{a.X() * s, a.Y() * s, a[2] * s}
}

// coordDet3D returns the determinant of the 3x3 matrix formed by the vectors a and b and a unit vector along z.
func coordDet3D(a geom.Coord, b geom.Coord) float64 {
	return a.X()*b.Y()*1 + a.Y()*1*a[2] + 1*b.X()*b[2] - 1*b.Y()*a[2] - a.X()*1*b[2] - a.Y()*b.X()*1
}

// coordDot3D returns the dot product of two 3D coords if the coord was a vector.
func coordDot3D(a geom.Coord, b geom.Coord) float64 {
	return a.X()*b.X() + a.Y()*b.Y() + a[2]*b[2]
}

// coordCross3D returns the cross product of two 3D coords if the coord was a vector.
func coordCross3D(a geom.Coord, b geom.Coord) geom.Coord {
	return geom.Coord{
		a.Y()*b[2] - a[2]*b.Y(),
		a[2]*b.X() - a.X()*b[2],
		a.X()*b.Y() - a.Y()*b.X(),
	}
}

// coordNorm2_3D returns the normalization^2 of a 3D coordinate if the coord was a vector.
func coordNorm2_3D(c geom.Coord) float64 {
	return coordDot3D(c, c)
}

// coordNorm3D returns the normalization of a 3D coordinate if the coord was a vector.
func coordNorm3D(c geom.Coord) float64 {
	return math.Sqrt(coordNorm2_3D(c))
}

// coordEqual3D returns whether two 3D coordinates are equal.
func coordEqual3D(a geom.Coord, b geom.Coord) bool {
	return a.X() == b.X() && a.Y() == b.Y() && a[2] == b[2]
}

// coordMag2_3D returns the magnitude^2 of a 3D coordinate if the coord was a vector.
func coordMag2_3D(c geom.Coord) float64 {
	return coordDot3D(c, c)
}
