// This file is part of go-mc/server project.
// Copyright (C) 2023.  Tnze
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

package client

import (
	"bytes"
	"encoding/binary"
	"sync/atomic"
	"unsafe"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/chat/sign"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/level"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/go-mc/server/world"
)

func (c *Client) SendPacket(id packetid.ClientboundPacketID, fields ...pk.FieldEncoder) {
	var buffer bytes.Buffer

	// Write the packet fields
	for i := range fields {
		if _, err := fields[i].WriteTo(&buffer); err != nil {
			c.log.Panic("Marshal packet error", zap.Error(err))
		}
	}

	// Send the packet data
	c.queue.Push(pk.Packet{
		ID:   int32(id),
		Data: buffer.Bytes(),
	})
}

func (c *Client) SendKeepAlive(id int64) {
	c.SendPacket(packetid.ClientboundKeepAlive, pk.Long(id))
}

// SendDisconnect send ClientboundDisconnect packet to client.
// Once the packet is sent, the connection will be closed.
func (c *Client) SendDisconnect(reason chat.Message) {
	c.log.Debug("Disconnect player", zap.String("reason", reason.ClearString()))
	c.SendPacket(packetid.ClientboundDisconnect, reason)
}

func (c *Client) SendLogin(w *world.World, p *world.Player) {
	hashedSeed := w.HashedSeed()
	c.SendPacket(
		packetid.ClientboundLogin,
		pk.Int(p.EntityID),
		pk.Boolean(false), // Is Hardcore
		pk.Byte(p.Gamemode),
		pk.Byte(-1),
		pk.Array([]pk.Identifier{
			pk.Identifier(w.Name()),
		}),
		pk.NBT(world.NetworkCodec),
		pk.Identifier("minecraft:overworld"),
		pk.Identifier(w.Name()),
		pk.Long(binary.BigEndian.Uint64(hashedSeed[:8])),
		pk.VarInt(0),              // Max players (ignored by client)
		pk.VarInt(p.ViewDistance), // View Distance
		pk.VarInt(p.ViewDistance), // Simulation Distance
		pk.Boolean(false),         // Reduced Debug Info
		pk.Boolean(false),         // Enable respawn screen
		pk.Boolean(false),         // Is Debug
		pk.Boolean(false),         // Is Flat
		pk.Boolean(false),         // Has Last Death Location
	)
}

func (c *Client) SendServerData(motd *chat.Message, favIcon string, enforceSecureProfile bool) {
	c.SendPacket(
		packetid.ClientboundServerData,
		pk.OptionEncoder[*chat.Message]{
			Has: motd != nil,
			Val: motd,
		},
		pk.Option[pk.String, *pk.String]{
			Has: favIcon != "",
			Val: pk.String(favIcon),
		},
		pk.Boolean(enforceSecureProfile),
	)
}

// Actions of [SendPlayerInfoUpdate]
const (
	PlayerInfoAddPlayer = iota
	PlayerInfoInitializeChat
	PlayerInfoUpdateGameMode
	PlayerInfoUpdateListed
	PlayerInfoUpdateLatency
	PlayerInfoUpdateDisplayName
	// PlayerInfoEnumGuard is the number of the enums
	PlayerInfoEnumGuard
)

func NewPlayerInfoAction(actions ...int) pk.FixedBitSet {
	enumSet := pk.NewFixedBitSet(PlayerInfoEnumGuard)
	for _, action := range actions {
		enumSet.Set(action, true)
	}
	return enumSet
}

func (c *Client) SendPlayerInfoUpdate(actions pk.FixedBitSet, players []*world.Player) {
	var buf bytes.Buffer
	_, _ = actions.WriteTo(&buf)
	_, _ = pk.VarInt(len(players)).WriteTo(&buf)
	for _, player := range players {
		_, _ = pk.UUID(player.UUID).WriteTo(&buf)
		if actions.Get(PlayerInfoAddPlayer) {
			_, _ = pk.String(player.Name).WriteTo(&buf)
			_, _ = pk.Array(player.Properties).WriteTo(&buf)
		}
		if actions.Get(PlayerInfoInitializeChat) {
			panic("not yet support InitializeChat")
		}
		if actions.Get(PlayerInfoUpdateGameMode) {
			_, _ = pk.VarInt(player.Gamemode).WriteTo(&buf)
		}
		if actions.Get(PlayerInfoUpdateListed) {
			_, _ = pk.Boolean(true).WriteTo(&buf)
		}
		if actions.Get(PlayerInfoUpdateLatency) {
			_, _ = pk.VarInt(player.Latency.Milliseconds()).WriteTo(&buf)
		}
		if actions.Get(PlayerInfoUpdateDisplayName) {
			panic("not yet support DisplayName")
		}
	}
	c.queue.Push(pk.Packet{
		ID:   int32(packetid.ClientboundPlayerInfoUpdate),
		Data: buf.Bytes(),
	})
}

func (c *Client) SendPlayerInfoRemove(players []*world.Player) {
	var buff bytes.Buffer

	if _, err := pk.VarInt(len(players)).WriteTo(&buff); err != nil {
		c.log.Panic("Marshal packet error", zap.Error(err))
	}
	for _, p := range players {
		if _, err := pk.UUID(p.UUID).WriteTo(&buff); err != nil {
			c.log.Panic("Marshal packet error", zap.Error(err))
		}
	}

	c.queue.Push(pk.Packet{
		ID:   int32(packetid.ClientboundPlayerInfoRemove),
		Data: buff.Bytes(),
	})
}

func (c *Client) SendLevelChunkWithLight(pos level.ChunkPos, chunk *level.Chunk) {
	c.SendPacket(packetid.ClientboundLevelChunkWithLight, pos, chunk)
}

func (c *Client) SendForgetLevelChunk(pos level.ChunkPos) {
	c.SendPacket(packetid.ClientboundForgetLevelChunk, pos)
}

func (c *Client) SendAddPlayer(p *world.Player) {
	c.SendPacket(
		packetid.ClientboundAddPlayer,
		pk.VarInt(p.EntityID),
		pk.UUID(p.UUID),
		pk.Double(p.Position[0]),
		pk.Double(p.Position[1]),
		pk.Double(p.Position[2]),
		pk.Angle(p.Rotation[0]),
		pk.Angle(p.Rotation[1]),
	)
}

func (c *Client) SendMoveEntitiesPos(eid int32, delta [3]int16, onGround bool) {
	c.SendPacket(
		packetid.ClientboundMoveEntityPos,
		pk.VarInt(eid),
		pk.Short(delta[0]),
		pk.Short(delta[1]),
		pk.Short(delta[2]),
		pk.Boolean(onGround),
	)
}

func (c *Client) SendMoveEntitiesPosAndRot(eid int32, delta [3]int16, rot [2]int8, onGround bool) {
	c.SendPacket(
		packetid.ClientboundMoveEntityPosRot,
		pk.VarInt(eid),
		pk.Short(delta[0]),
		pk.Short(delta[1]),
		pk.Short(delta[2]),
		pk.Angle(rot[0]),
		pk.Angle(rot[1]),
		pk.Boolean(onGround),
	)
}

func (c *Client) SendMoveEntitiesRot(eid int32, rot [2]int8, onGround bool) {
	c.SendPacket(
		packetid.ClientboundMoveEntityRot,
		pk.VarInt(eid),
		pk.Angle(rot[0]),
		pk.Angle(rot[1]),
		pk.Boolean(onGround),
	)
}

func (c *Client) SendRotateHead(eid int32, yaw int8) {
	c.SendPacket(
		packetid.ClientboundRotateHead,
		pk.VarInt(eid),
		pk.Angle(yaw),
	)
}

func (c *Client) SendTeleportEntity(eid int32, pos [3]float64, rot [2]int8, onGround bool) {
	c.SendPacket(
		packetid.ClientboundTeleportEntity,
		pk.VarInt(eid),
		pk.Double(pos[0]),
		pk.Double(pos[1]),
		pk.Double(pos[2]),
		pk.Angle(rot[0]),
		pk.Angle(rot[1]),
		pk.Boolean(onGround),
	)
}

var teleportCounter atomic.Int32

func (c *Client) SendPlayerPosition(pos [3]float64, rot [2]float32, dismountVehicle bool) (teleportID int32) {
	teleportID = teleportCounter.Add(1)
	c.SendPacket(
		packetid.ClientboundPlayerPosition,
		pk.Double(pos[0]),
		pk.Double(pos[1]),
		pk.Double(pos[2]),
		pk.Float(rot[0]),
		pk.Float(rot[1]),
		pk.Byte(0), // Absolute
		pk.VarInt(teleportID),
		pk.Boolean(dismountVehicle),
	)
	return
}

func (c *Client) SendSetDefaultSpawnPosition(xyz [3]int32, angle float32) {
	c.SendPacket(
		packetid.ClientboundSetDefaultSpawnPosition,
		pk.Position{X: int(xyz[0]), Y: int(xyz[1]), Z: int(xyz[2])},
		pk.Float(angle),
	)
	return
}

func (c *Client) SendRemoveEntities(entityIDs []int32) {
	c.SendPacket(
		packetid.ClientboundRemoveEntities,
		pk.Array(*(*[]pk.VarInt)(unsafe.Pointer(&entityIDs))),
	)
}

func (c *Client) SendSystemChat(msg chat.Message, overlay bool) {
	c.SendPacket(packetid.ClientboundSystemChat, msg, pk.Boolean(overlay))
}

func (c *Client) SendPlayerChat(
	sender uuid.UUID,
	index int32,
	signature pk.Option[sign.Signature, *sign.Signature],
	body *sign.PackedMessageBody,
	unsignedContent *chat.Message,
	filter *sign.FilterMask,
	chatType *chat.Type,
) {
	c.SendPacket(
		packetid.ClientboundPlayerChat,
		pk.UUID(sender),
		pk.VarInt(index),
		signature,
		body,
		pk.OptionEncoder[*chat.Message]{
			Has: unsignedContent != nil,
			Val: unsignedContent,
		},
		filter,
		chatType,
	)
}

func (c *Client) SendSetChunkCacheCenter(chunkPos [2]int32) {
	c.SendPacket(
		packetid.ClientboundSetChunkCacheCenter,
		pk.VarInt(chunkPos[0]),
		pk.VarInt(chunkPos[1]),
	)
}

func (c *Client) ViewChunkLoad(pos level.ChunkPos, chunk *level.Chunk) {
	c.SendLevelChunkWithLight(pos, chunk)
}
func (c *Client) ViewChunkUnload(pos level.ChunkPos)   { c.SendForgetLevelChunk(pos) }
func (c *Client) ViewAddPlayer(p *world.Player)        { c.SendAddPlayer(p) }
func (c *Client) ViewRemoveEntities(entityIDs []int32) { c.SendRemoveEntities(entityIDs) }
func (c *Client) ViewMoveEntityPos(id int32, delta [3]int16, onGround bool) {
	c.SendMoveEntitiesPos(id, delta, onGround)
}

func (c *Client) ViewMoveEntityPosAndRot(id int32, delta [3]int16, rot [2]int8, onGround bool) {
	c.SendMoveEntitiesPosAndRot(id, delta, rot, onGround)
}

func (c *Client) ViewMoveEntityRot(id int32, rot [2]int8, onGround bool) {
	c.SendMoveEntitiesRot(id, rot, onGround)
}

func (c *Client) ViewRotateHead(id int32, yaw int8) {
	c.SendRotateHead(id, yaw)
}

func (c *Client) ViewTeleportEntity(id int32, pos [3]float64, rot [2]int8, onGround bool) {
	c.SendTeleportEntity(id, pos, rot, onGround)
}
