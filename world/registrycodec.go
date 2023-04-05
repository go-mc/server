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
	"bytes"
	_ "embed"

	"github.com/Tnze/go-mc/registry"

	"github.com/Tnze/go-mc/nbt"
)

//go:embed RegistryCodec.nbt
var networkCodecBytes []byte
var NetworkCodec registry.NetworkCodec

func init() {
	r := bytes.NewReader(networkCodecBytes)
	d := nbt.NewDecoder(r)
	_, err := d.Decode(&NetworkCodec)
	if err != nil {
		panic(err)
	}
}
