package world

import (
	"io"
	"sync"
	"time"

	pk "github.com/Tnze/go-mc/net/packet"

	"github.com/google/uuid"

	"github.com/Tnze/go-mc/server/auth"
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
	PubKey     *auth.PublicKey
	Properties []auth.Property
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
