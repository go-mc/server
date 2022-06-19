package player

import (
	"github.com/go-mc/server/world"
	"github.com/google/uuid"
)

type Player struct {
	name     string
	uuid     uuid.UUID
	entityID int32

	w *world.World

	pos          [3]float64
	viewDistance int32
	gamemode     int32
}

func (p *Player) ChunkPos() [2]int32 {
	return [2]int32{int32(p.pos[0]) >> 5, int32(p.pos[2]) >> 5}
}

func (p *Player) ID() int32 {
	return p.entityID
}

func (p *Player) GameMode() byte {
	return byte(p.gamemode)
}

func (p *Player) World() *world.World {
	return p.w
}

func (p *Player) SetWorld(w *world.World) {
	p.w = w
}

func (p *Player) ChunkRadius() int32 {
	return p.viewDistance
}

func (p *Player) ClientInformation() {

}

func New(name string, id uuid.UUID) *Player {
	return &Player{
		name: name,
		uuid: id,
	}
}
