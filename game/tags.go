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

package game

import (
	"io"

	pk "github.com/Tnze/go-mc/net/packet"
)

type Tag[T ~int32 | ~int] struct {
	Name   string
	Values map[string][]T
}

func (t Tag[T]) WriteTo(w io.Writer) (n int64, err error) {
	n1, err := pk.Identifier(t.Name).WriteTo(w)
	if err != nil {
		return n1, err
	}
	n2, err := pk.VarInt(len(t.Values)).WriteTo(w)
	if err != nil {
		return n1 + n2, err
	}
	for k, v := range t.Values {
		n3, err := pk.Identifier(k).WriteTo(w)
		n += n3
		if err != nil {
			return n + n1 + n2, err
		}
		n4, err := pk.VarInt(len(v)).WriteTo(w)
		n += n4
		if err != nil {
			return n + n1 + n2, err
		}
		for _, v := range v {
			n5, err := pk.VarInt(v).WriteTo(w)
			n += n5
			if err != nil {
				return n + n1 + n2, err
			}
		}
	}
	return n + n1 + n2, err
}

var defaultTags = []pk.FieldEncoder{
	Tag[int32]{
		Name: "minecraft:fluid",
		Values: map[string][]int32{
			"minecraft:water": {1, 2},
			"minecraft:lava":  {3, 4},
		},
	},
}
