// This file is part of go-mc/server project.
// Copyright (C) 2023.  Tnze
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package bvh

import "testing"

func TestAABB_WithIn(t *testing.T) {
	aabb := AABB[float64, Vec2[float64]]{
		Upper: Vec2[float64]{2, 2},
		Lower: Vec2[float64]{-1, -1},
	}
	if !aabb.WithIn(Vec2[float64]{0, 0}) {
		panic("(0, 0) should included")
	}
	if aabb.WithIn(Vec2[float64]{-2, -2}) {
		panic("(-2, -2) shouldn't included")
	}

	aabb2 := AABB[int, Vec3[int]]{
		Upper: Vec3[int]{1, 1, 1},
		Lower: Vec3[int]{-1, -1, -1},
	}
	if !aabb2.WithIn(Vec3[int]{0, 0, 0}) {
		panic("(0, 0, 0) should included")
	}
	if aabb2.WithIn(Vec3[int]{-2, -2, 0}) {
		panic("(-2, -2, 0) shouldn't included")
	}

	sphere := Sphere[float64, Vec2[float64]]{
		Center: Vec2[float64]{0, 0},
		R:      1.0,
	}
	if !sphere.WithIn(Vec2[float64]{0, 0}) {
		t.Errorf("(0,0) is in")
	}
	if sphere.WithIn(Vec2[float64]{1, 1}) {
		t.Errorf("(1,1) isn't in")
	}
}
