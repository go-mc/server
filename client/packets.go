package client

import (
	"bytes"
	"encoding/binary"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/level"
	pk "github.com/Tnze/go-mc/net/packet"
	"sync"
)

var bufferPool = sync.Pool{New: func() any {
	return new(bytes.Buffer)
}}

func (c *Client) sendPacket(id int32, fields ...pk.FieldEncoder) error {
	// Get buffers from the pool
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)
	buffer.Reset()

	// Write the packet fields
	for i := range fields {
		if _, err := fields[i].WriteTo(buffer); err != nil {
			return err
		}
	}

	// Send the packet data
	return c.conn.WritePacket(pk.Packet{
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

func (c *Client) SendLogin(p *Player) error {
	w := p.World()
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
