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

package ecsnbt

import (
	"fmt"
	"io"

	"github.com/Tnze/go-ecs"
	ecsreflect "github.com/Tnze/go-ecs/reflect"
	"github.com/Tnze/go-mc/nbt"
)

type EntityMarshaler struct {
	TagName ecs.Component
	World   *ecs.World
	Entity  ecs.Entity
}

func (e EntityMarshaler) TagType() byte {
	return nbt.TagCompound
}

func (e EntityMarshaler) MarshalNBT(w io.Writer) error {
	enc := nbt.NewEncoder(w)
	v := ecsreflect.ValueOf(e.World, e.Entity)
	n := v.NumComps()
	for i := 0; i < n; i++ {
		c, val := v.IndexComps(i)
		if val == nil {
			continue
		}

		name := ecs.GetComp[string](e.World, c.Entity, e.TagName)
		if name == nil {
			return fmt.Errorf("%#v doesn't have Name component", c)
		}

		if err := enc.Encode(val, *name); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte{nbt.TagEnd})
	return err
}
