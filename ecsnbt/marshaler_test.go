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
	"encoding/hex"
	"testing"

	"github.com/google/uuid"

	"github.com/Tnze/go-ecs"
	"github.com/Tnze/go-mc/nbt"
)

func TestEntityMarshaler(t *testing.T) {
	w := ecs.NewWorld()
	tagNameComp := ecs.NewComponent(w)
	newComp := func(tagName string) (comp ecs.Component) {
		comp = ecs.NewComponent(w)
		ecs.SetComp(w, comp.Entity, tagNameComp, tagName)
		return
	}

	id := newComp("id")
	invulnerable := newComp("Invulnerable")
	onGround := newComp("OnGround")
	UUID := newComp("UUID")
	fire := newComp("Fire")

	chest := ecs.NewEntity(w)
	ecs.SetComp(w, chest, id, "minecraft:chest_minecart")
	ecs.SetComp(w, chest, invulnerable, byte(0))
	ecs.SetComp(w, chest, onGround, false)
	eUUID := uuid.UUID{
		0xF0, 0x4C, 0xAA, 0x0C, 0xE7, 0x41, 0x19, 0x6A,
		0xD3, 0xFE, 0x46, 0x88, 0x71, 0xEB, 0x10, 0xE3,
	}
	ecs.SetComp(w, chest, UUID, eUUID)
	ecs.SetComp(w, chest, fire, int16(-1))

	em := EntityMarshaler{
		TagName: tagNameComp,
		World:   w,
		Entity:  chest,
	}

	binary, err := nbt.Marshal(em)
	if err != nil {
		t.Error(err)
	}

	t.Log("Data:\n" + hex.Dump(binary))

	var data struct {
		ID           string `nbt:"id"`
		Invulnerable byte
		OnGround     bool
		UUID         uuid.UUID
		Fire         int16
	}
	if err := nbt.Unmarshal(binary, &data); err != nil {
		t.Error(err)
	}
	if data.ID != "minecraft:chest_minecart" ||
		data.Invulnerable != 0 ||
		data.OnGround != false ||
		data.UUID != eUUID ||
		data.Fire != -1 {
		t.Error("data not match")
	}
}
