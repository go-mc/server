package client

import "github.com/go-mc/server/world"

type Player interface {
	world.Entity
	GameMode() byte
	World() *world.World
	ChunkRadius() int32
}
