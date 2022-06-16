package player

import (
	"github.com/go-mc/server/world"
	"github.com/google/uuid"
)

type Player struct {
	name     string
	uuid     uuid.UUID
	entityID int32
}

func (p *Player) ID() int32 {
	return p.entityID
}

func (p *Player) GameMode() byte {
	//TODO implement me
	panic("implement me")
}

func (p *Player) World() *world.World {
	//TODO implement me
	panic("implement me")
}

func (p *Player) ChunkRadius() int32 {
	//TODO implement me
	panic("implement me")
}

func New(name string, id uuid.UUID) *Player {
	return &Player{
		name: name,
		uuid: id,
	}
}
