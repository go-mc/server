package world

import (
	"github.com/Tnze/go-mc/server/auth"
	"github.com/df-mc/atomic"
	"github.com/google/uuid"
	"time"
)

type Player struct {
	Entity
	Name       string
	UUID       uuid.UUID
	PubKey     *auth.PublicKey
	Properties []auth.Property

	ChunkPos     [2]int32
	ViewDistance int32

	Gamemode       int32
	EntitiesInView map[int32]*Entity
	view           *playerViewNode
	teleport       *TeleportRequest

	nextPos          atomic.Value[Position]
	nextRot          atomic.Value[Rotation]
	nextOnGround     atomic.Bool
	latency          atomic.Duration
	acceptTeleportID atomic.Int32
}

func (p *Player) ChunkPosition() [2]int32 { return p.ChunkPos }
func (p *Player) ChunkRadius() int32      { return p.ViewDistance }

func (p *Player) NextPosition() [3]float64         { return p.nextPos.Load() }
func (p *Player) SetNextPosition(pos [3]float64)   { p.nextPos.Store(pos) }
func (p *Player) NextRotation() [2]float32         { return p.nextRot.Load() }
func (p *Player) SetNextRotation(rot [2]float32)   { p.nextRot.Store(rot) }
func (p *Player) NextOnGround() [2]float32         { return p.nextRot.Load() }
func (p *Player) SetNextOnGround(onGround bool)    { p.nextOnGround.Store(onGround) }
func (p *Player) Latency() time.Duration           { return p.latency.Load() }
func (p *Player) SetLatency(latency time.Duration) { p.latency.Store(latency) }
func (p *Player) AcceptTeleport(id int32)          { p.acceptTeleportID.Store(id) }

// getView 根据玩家Position和ViewDistance计算玩家可视距离包围盒
func (p *Player) getView() aabb3d {
	viewDistance := float64(p.ViewDistance) * 16 // ViewDistance单位是1 Chunk（16 Block）
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
