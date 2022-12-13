package client

import (
	"bytes"
	"encoding/binary"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/chat/sign"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/level"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/go-mc/server/world"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (c *Client) sendPacket(id packetid.ClientboundPacketID, fields ...pk.FieldEncoder) {
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
	c.sendPacket(packetid.ClientboundKeepAlive, pk.Long(id))
}

// SendDisconnect send ClientboundDisconnect packet to client.
// Once the packet is sent, the connection will be closed.
func (c *Client) SendDisconnect(reason chat.Message) {
	c.log.Debug("Disconnect player", zap.String("reason", reason.ClearString()))
	c.sendPacket(packetid.ClientboundDisconnect, reason)
}

func (c *Client) SendLogin(w *world.World, p *world.Player) {
	hashedSeed := w.HashedSeed()
	c.sendPacket(
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
	c.sendPacket(
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

// playerInfoUpdate Enums
const (
	playerInfoAddPlayer = iota
	playerInfoInitializeChat
	playerInfoUpdateGameMode
	playerInfoUpdateListed
	playerInfoUpdateLatency
	playerInfoUpdateDisplayName
	playerInfoEnumGuard
)

func (c *Client) SendPlayerInfoAdd(players []*world.Player) {
	enumSet := pk.NewFixedBitSet(playerInfoEnumGuard)
	enumSet.Set(playerInfoAddPlayer, true)

	var buffer bytes.Buffer
	_, err := pk.Tuple{
		enumSet,                 // Actions
		pk.VarInt(len(players)), // Number of players
	}.WriteTo(&buffer)
	if err != nil {
		c.log.Panic("Marshal packet error", zap.Error(err))
	}

	// Player
	for _, p := range players {
		_, err := pk.Tuple{
			pk.UUID(p.UUID),
			pk.String(p.Name),
			pk.Array(p.Properties),
		}.WriteTo(&buffer)
		if err != nil {
			c.log.Panic("Marshal packet error", zap.Error(err))
		}
	}
	c.queue.Push(pk.Packet{
		ID:   int32(packetid.ClientboundPlayerInfoUpdate),
		Data: buffer.Bytes(),
	})
}

func (c *Client) SendPlayerInfoUpdateLatency(player *world.Player, latency time.Duration) {
	enumSet := pk.NewFixedBitSet(playerInfoEnumGuard)
	enumSet.Set(playerInfoUpdateLatency, true)
	c.sendPacket(
		packetid.ClientboundPlayerInfoUpdate,
		enumSet,
		pk.VarInt(1),
		pk.UUID(player.UUID),
		pk.VarInt(latency.Milliseconds()),
	)
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
	c.sendPacket(packetid.ClientboundLevelChunkWithLight, pos, chunk)
}

func (c *Client) SendForgetLevelChunk(pos level.ChunkPos) {
	c.sendPacket(packetid.ClientboundForgetLevelChunk, pos)
}

func (c *Client) SendAddPlayer(p *world.Player) {
	c.sendPacket(
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
	c.sendPacket(
		packetid.ClientboundMoveEntityPos,
		pk.VarInt(eid),
		pk.Short(delta[0]),
		pk.Short(delta[1]),
		pk.Short(delta[2]),
		pk.Boolean(onGround),
	)
}

func (c *Client) SendMoveEntitiesPosAndRot(eid int32, delta [3]int16, rot [2]int8, onGround bool) {
	c.sendPacket(
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
	c.sendPacket(
		packetid.ClientboundMoveEntityRot,
		pk.VarInt(eid),
		pk.Angle(rot[0]),
		pk.Angle(rot[1]),
		pk.Boolean(onGround),
	)
}

func (c *Client) SendRotateHead(eid int32, yaw int8) {
	c.sendPacket(
		packetid.ClientboundRotateHead,
		pk.VarInt(eid),
		pk.Angle(yaw),
	)
}

func (c *Client) SendTeleportEntity(eid int32, pos [3]float64, rot [2]int8, onGround bool) {
	c.sendPacket(
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
	c.sendPacket(
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

func (c *Client) SendRemoveEntities(entityIDs []int32) {
	c.sendPacket(
		packetid.ClientboundRemoveEntities,
		pk.Array(*(*[]pk.VarInt)(unsafe.Pointer(&entityIDs))),
	)
}

func (c *Client) SendSystemChat(msg chat.Message, overlay bool) {
	c.sendPacket(packetid.ClientboundSystemChat, msg, pk.Boolean(overlay))
}

func (c *Client) SendPlayerChat(
	sender uuid.UUID,
	index int32,
	signature pk.Option[sign.Signature, *sign.Signature],
	body *sign.MessageBody,
	unsignedContent *chat.Message,
	filter *sign.FilterMask,
	chatType *chat.Type,
) {
	c.sendPacket(
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
	c.sendPacket(
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
