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
	"io"
	"sync"
	"time"

	"github.com/google/uuid"

	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/yggdrasil/user"
)

func (i *ClientInfo) ReadFrom(r io.Reader) (n int64, err error) {
	return pk.Tuple{
		(*pk.String)(&i.Locale),
		(*pk.Byte)(&i.ViewDistance),
		(*pk.VarInt)(&i.ChatMode),
		(*pk.Boolean)(&i.ChatColors),
		(*pk.UnsignedByte)(&i.DisplayedSkinParts),
		(*pk.VarInt)(&i.MainHand),
		(*pk.Boolean)(&i.EnableTextFiltering),
		(*pk.Boolean)(&i.AllowServerListings),
	}.ReadFrom(r)
}

type Player struct {
	Entity
	Name       string
	UUID       uuid.UUID
	PubKey     *user.PublicKey
	Properties []user.Property
	Latency    time.Duration

	lastChatTimestamp time.Time
	lastChatSignature []byte

	ChunkPos     [3]int32
	ViewDistance int32

	Gamemode       int32
	EntitiesInView map[int32]*Entity
	view           *playerViewNode
	teleport       *TeleportRequest

	Inputs Inputs
}

func (p *Player) chunkPosition() [2]int32 { return [2]int32{p.ChunkPos[0], p.ChunkPos[2]} }
func (p *Player) chunkRadius() int32      { return p.ViewDistance }

// getView calculate the visual range enclosure with Position and ViewDistance of a player.
func (p *Player) getView() aabb3d {
	viewDistance := float64(p.ViewDistance) * 16 // the unit of ViewDistance is 1 Chunk（16 Block）
	return aabb3d{
		Upper: vec3d{p.Position[0] + viewDistance, p.Position[1] + viewDistance, p.Position[2] + viewDistance},
		Lower: vec3d{p.Position[0] - viewDistance, p.Position[1] - viewDistance, p.Position[2] - viewDistance},
	}
}

type TeleportRequest struct {
	ID int32
	Position
	Rotation
}

type Inputs struct {
	sync.Mutex
	ClientInfo
	Position
	Rotation
	OnGround
	Latency    time.Duration
	TeleportID int32
}

type ClientInfo struct {
	Locale              string
	ViewDistance        int8
	ChatMode            int32
	ChatColors          bool
	DisplayedSkinParts  byte
	MainHand            int32
	EnableTextFiltering bool
	AllowServerListings bool
}
