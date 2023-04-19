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
	"errors"
	"fmt"
	"io"

	"github.com/Tnze/go-ecs"
	"github.com/Tnze/go-mc/nbt"
)

type EntityUnmarshaler struct {
	TagName ecs.Component
	World   *ecs.World
	Entity  ecs.Entity

	Components map[string]ecs.Component
}

func (e EntityUnmarshaler) UnmarshalNBT(tagType byte, r nbt.DecoderReader) error {
	if tagType != nbt.TagCompound {
		return errors.New("unsupported unmarshal from non-Compound Tag")
	}
	for {
		fieldType, tagName, err := readTag(r)
		if err != nil {
			return err
		}
		_, ok := e.Components[tagName]
		if !ok {
			return fmt.Errorf("unknwon tagName %q", tagName)
		}
		switch fieldType {
		case nbt.TagEnd:
			return nil
		case nbt.TagByte:
		case nbt.TagShort:
		case nbt.TagInt:
		case nbt.TagLong:
		case nbt.TagFloat:
		case nbt.TagDouble:
		case nbt.TagByteArray:
		case nbt.TagString:
		case nbt.TagList:
		case nbt.TagCompound:
		case nbt.TagIntArray:
		case nbt.TagLongArray:
		default:
			return fmt.Errorf("unknown tag type %#x", fieldType)
		}
	}
}

func readTag(r nbt.DecoderReader) (tagType byte, tagName string, err error) {
	tagType, err = r.ReadByte()
	if err != nil {
		return
	}

	switch tagType {
	case nbt.TagEnd:
	default: // Read Tag
		tagName, err = readString(r)
	}
	return
}

func readString(r nbt.DecoderReader) (string, error) {
	length, err := readShort(r)
	if err != nil {
		return "", err
	} else if length < 0 {
		return "", errors.New("string length less than 0")
	}

	var str string
	if length > 0 {
		buf := make([]byte, length)
		_, err = io.ReadFull(r, buf)
		str = string(buf)
	}
	return str, err
}

func readShort(r nbt.DecoderReader) (int16, error) {
	var data [2]byte
	_, err := io.ReadFull(r, data[:])
	return int16(data[0])<<8 | int16(data[1]), err
}
