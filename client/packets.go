package client

import (
	"bytes"
	"encoding/binary"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/level"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/go-mc/server/player"
	"github.com/go-mc/server/world"
	"go.uber.org/zap"
	"time"
)

func (c *Client) sendPacket(id int32, fields ...pk.FieldEncoder) {
	var buffer bytes.Buffer

	// Write the packet fields
	for i := range fields {
		if _, err := fields[i].WriteTo(&buffer); err != nil {
			c.log.Panic("Marshal packet error", zap.Error(err))
		}
	}

	// Send the packet data
	c.queue.Push(pk.Packet{
		ID:   id,
		Data: buffer.Bytes(),
	})
}

func (c *Client) SendKeepAlive(id int64) {
	c.sendPacket(packetid.ClientboundKeepAlive, pk.Long(id))
}

func (c *Client) SendDisconnect(reason chat.Message) {
	c.sendPacket(packetid.ClientboundDisconnect, reason)
}

func (c *Client) SendLogin(w *world.World, p *player.Player) {
	hashedSeed := w.HashedSeed()
	c.sendPacket(
		packetid.ClientboundLogin,
		pk.Int(p.ID()),
		pk.Boolean(false), // Is Hardcore
		pk.Byte(p.Gamemode()),
		pk.Byte(-1),
		pk.Array([]pk.String{
			pk.String(w.Name()),
		}),
		pk.NBT(w.DimensionCodec()),
		pk.Identifier("minecraft:overworld"),
		pk.String(w.Name()),
		pk.Long(binary.BigEndian.Uint64(hashedSeed[:8])),
		pk.VarInt(0),               // Max players (ignored by client)
		pk.VarInt(p.ChunkRadius()), // View Distance
		pk.VarInt(p.ChunkRadius()), // Simulation Distance
		pk.Boolean(false),          // Reduced Debug Info
		pk.Boolean(false),          // Enable respawn screen
		pk.Boolean(false),          // Is Debug
		pk.Boolean(false),          // Is Flat
		pk.Boolean(false),          // Has Last Death Location
	)
}

func (c *Client) SendPlayerInfoAdd(players []*player.Player) {
	var buffer bytes.Buffer
	_, err := pk.Tuple{
		pk.VarInt(0),            // Action
		pk.VarInt(len(players)), // Number of players
	}.WriteTo(&buffer)
	if err != nil {
		c.log.Panic("Marshal packet error", zap.Error(err))
	}

	// Player
	for _, p := range players {
		pubKey := p.PublicKey()
		_, err := pk.Tuple{
			pk.UUID(p.UUID()),
			pk.String(p.Name()),
			pk.Array(p.Properties()),
			pk.VarInt(p.Gamemode()),
			pk.VarInt(p.Latency().Milliseconds()),
			pk.Boolean(false), // Has Display Name
			pk.Boolean(pubKey != nil),
			pk.Opt{
				Has:   pubKey != nil,
				Field: pubKey,
			},
		}.WriteTo(&buffer)
		if err != nil {
			c.log.Panic("Marshal packet error", zap.Error(err))
		}
	}
	c.queue.Push(pk.Packet{
		ID:   packetid.ClientboundPlayerInfo,
		Data: buffer.Bytes(),
	})
}

func (c *Client) SendPlayerInfoUpdateLatency(player *player.Player, latency time.Duration) {
	c.sendPacket(
		packetid.ClientboundPlayerInfo,
		pk.VarInt(2),
		pk.VarInt(1),
		pk.UUID(player.UUID()),
		pk.VarInt(latency.Milliseconds()),
	)
}

func (c *Client) SendPlayerInfoRemove(player *player.Player) {
	c.sendPacket(
		packetid.ClientboundPlayerInfo,
		pk.VarInt(4),
		pk.VarInt(1),
		pk.UUID(player.UUID()),
	)
}

func (c *Client) SendLevelChunkWithLight(pos level.ChunkPos, chunk *level.Chunk) {
	c.sendPacket(packetid.ClientboundLevelChunkWithLight, pos, chunk)
}

func (c *Client) SendForgetLevelChunk(pos level.ChunkPos) {
	c.sendPacket(packetid.ClientboundForgetLevelChunk, pos)
}

func (c *Client) ViewChunkLoad(pos level.ChunkPos, chunk *level.Chunk) {
	c.SendLevelChunkWithLight(pos, chunk)
}
func (c *Client) ViewChunkUnload(pos level.ChunkPos) { c.SendForgetLevelChunk(pos) }
