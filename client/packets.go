package client

import (
	"bytes"
	"encoding/binary"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/level"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/go-mc/server/world"
	"go.uber.org/zap"
)

func (c *Client) sendPacket(id int32, fields ...pk.FieldEncoder) (err error) {
	var buffer bytes.Buffer

	// Write the packet fields
	for i := range fields {
		_, err = fields[i].WriteTo(&buffer)
		if err != nil {
			return
		}
	}

	// Send the packet data
	c.queue.Push(pk.Packet{
		ID:   id,
		Data: buffer.Bytes(),
	})
	return
}

func (c *Client) SendKeepAlive(id int64) {
	_ = c.sendPacket(packetid.ClientboundKeepAlive, pk.Long(id))
}

func (c *Client) SendDisconnect(reason chat.Message) {
	_ = c.sendPacket(packetid.ClientboundDisconnect, reason)

}

func (c *Client) SendLogin(w *world.World, p *Player) error {
	hashedSeed := w.HashedSeed()
	return c.sendPacket(
		packetid.ClientboundLogin,
		pk.Int(p.ID()),
		pk.Boolean(false), // Is Hardcore
		pk.Byte(p.GameMode()),
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

func (c *Client) SendLevelChunkWithLight(pos level.ChunkPos, chunk *level.Chunk) error {
	return c.sendPacket(
		packetid.ClientboundLevelChunkWithLight,
		pos, chunk,
	)
}

func (c *Client) ViewChunkLoad(pos level.ChunkPos, chunk *level.Chunk) {
	c.log.Info("Send chunk load", zap.Int32("x", pos[0]), zap.Int32("z", pos[1]))
	c.SendLevelChunkWithLight(pos, chunk)
}

func (c *Client) ViewChunkUnload(pos level.ChunkPos) {

}
