// This file is part of go-mc/server project.
// Copyright (C) 2022.  Tnze
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package bvh

import (
	"math"

	"golang.org/x/exp/constraints"
)

type Vec2[I constraints.Signed | constraints.Float] [2]I

func (v Vec2[I]) Add(other Vec2[I]) Vec2[I] { return Vec2[I]{v[0] + other[0], v[1] + other[1]} }
func (v Vec2[I]) Sub(other Vec2[I]) Vec2[I] { return Vec2[I]{v[0] - other[0], v[1] - other[1]} }
func (v Vec2[I]) Mul(i I) Vec2[I]           { return Vec2[I]{v[0] * i, v[1] * i} }
func (v Vec2[I]) Max(other Vec2[I]) Vec2[I] { return Vec2[I]{max(v[0], other[0]), max(v[1], other[1])} }
func (v Vec2[I]) Min(other Vec2[I]) Vec2[I] { return Vec2[I]{min(v[0], other[0]), min(v[1], other[1])} }
func (v Vec2[I]) Less(other Vec2[I]) bool   { return v[0] < other[0] && v[1] < other[1] }
func (v Vec2[I]) More(other Vec2[I]) bool   { return v[0] > other[0] && v[1] > other[1] }
func (v Vec2[I]) Norm() float64             { return sqrt(v[0]*v[0] + v[1]*v[1]) }
func (v Vec2[I]) Sum() I                    { return v[0] + v[1] }

type Vec3[I constraints.Signed | constraints.Float] [3]I

func (v Vec3[I]) Add(other Vec3[I]) Vec3[I] {
	return Vec3[I]{v[0] + other[0], v[1] + other[1], v[2] + other[2]}
}

func (v Vec3[I]) Sub(other Vec3[I]) Vec3[I] {
	return Vec3[I]{v[0] - other[0], v[1] - other[1], v[2] - other[2]}
}
func (v Vec3[I]) Mul(i I) Vec3[I] { return Vec3[I]{v[0] * i, v[1] * i, v[2] * i} }
func (v Vec3[I]) Max(other Vec3[I]) Vec3[I] {
	return Vec3[I]{max(v[0], other[0]), max(v[1], other[1]), max(v[2], other[2])}
}

func (v Vec3[I]) Min(other Vec3[I]) Vec3[I] {
	return Vec3[I]{min(v[0], other[0]), min(v[1], other[1]), min(v[2], other[2])}
}
func (v Vec3[I]) Less(other Vec3[I]) bool { return v[0] < other[0] && v[1] < other[1] }
func (v Vec3[I]) More(other Vec3[I]) bool { return v[0] > other[0] && v[1] > other[1] }
func (v Vec3[I]) Norm() float64           { return sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2]) }
func (v Vec3[I]) Sum() I                  { return v[0] + v[1] }

func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func sqrt[T constraints.Signed | constraints.Float](v T) float64 {
	return math.Sqrt(float64(v))
}
