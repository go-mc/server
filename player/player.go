package player

import (
	"github.com/Tnze/go-mc/server/auth"
	"github.com/df-mc/atomic"
	"github.com/google/uuid"
	"time"
)

type Player struct {
	name       string
	uuid       uuid.UUID
	pubKey     *auth.PublicKey
	properties []auth.Property
	entityID   int32

	pos          atomic.Value[[3]float64]
	rot          atomic.Value[[2]float32]
	viewDistance atomic.Int32
	gamemode     atomic.Int32
	latency      atomic.Duration
}

func New(name string, id uuid.UUID, pubKey *auth.PublicKey, properties []auth.Property) *Player {
	return &Player{name: name, uuid: id, pubKey: pubKey, properties: properties}
}

func (p *Player) Name() string                { return p.name }
func (p *Player) UUID() uuid.UUID             { return p.uuid }
func (p *Player) PublicKey() *auth.PublicKey  { return p.pubKey }
func (p *Player) Properties() []auth.Property { return p.properties }
func (p *Player) ID() int32                   { return p.entityID }
func (p *Player) Gamemode() int32             { return p.gamemode.Load() }
func (p *Player) SetGamemode(mode int32)      { p.gamemode.Store(mode) }
func (p *Player) ChunkPos() [2]int32 {
	pos := p.pos.Load()
	return [2]int32{int32(pos[0]) >> 5, int32(pos[2]) >> 5}
}
func (p *Player) ChunkRadius() int32               { return p.viewDistance.Load() }
func (p *Player) SetViewDistance(d int32)          { p.viewDistance.Store(d) }
func (p *Player) Latency() time.Duration           { return p.latency.Load() }
func (p *Player) SetLatency(latency time.Duration) { p.latency.Store(latency) }
func (p *Player) ClientInformation()               {}
func (p *Player) SetPos(pos [3]float64)            { p.pos.Store(pos) }
func (p *Player) SetRot(rot [2]float32)            { p.rot.Store(rot) }
