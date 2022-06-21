package player

import (
	"github.com/google/uuid"
	"sync/atomic"
)

type Player struct {
	name     string
	uuid     uuid.UUID
	entityID int32

	pos          atomic.Value //[3]float64
	rot          atomic.Value // [2]float32
	viewDistance int32
	gamemode     int32
}

func (p *Player) ChunkPos() [2]int32 {
	pos, ok := p.pos.Load().([3]float64)
	if !ok {
		pos = [3]float64{}
	}
	return [2]int32{int32(pos[0]) >> 5, int32(pos[2]) >> 5}
}

func (p *Player) ID() int32 {
	return p.entityID
}

func (p *Player) GameMode() byte {
	return byte(p.gamemode)
}

func (p *Player) ChunkRadius() int32 {
	return atomic.LoadInt32(&p.viewDistance)
}

func (p *Player) ClientInformation() {

}

func (p *Player) SetPos(pos [3]float64) {
	p.pos.Store(pos)
}
func (p *Player) SetRot(rot [2]float32) {
	p.rot.Store(rot)
}

func New(name string, id uuid.UUID) *Player {
	return &Player{
		name: name,
		uuid: id,
	}
}
