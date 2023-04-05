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

package world

import (
	"math"
	"sync/atomic"
)

var entityCounter atomic.Int32

func NewEntityID() int32 {
	return entityCounter.Add(1)
}

type Entity struct {
	EntityID int32
	Position
	Rotation
	OnGround
	pos0 Position
	rot0 Rotation
}

type (
	Position [3]float64
	Rotation [2]float32
	OnGround bool
)

func (e *Entity) getPoint() [2]float64 {
	return [2]float64{e.Position[0], e.Position[2]}
}

func (p *Position) IsValid() bool {
	return !math.IsNaN((*p)[0]) && !math.IsNaN((*p)[1]) && !math.IsNaN((*p)[2]) &&
		!math.IsInf((*p)[0], 0) && !math.IsInf((*p)[1], 0) && !math.IsInf((*p)[2], 0)
}
