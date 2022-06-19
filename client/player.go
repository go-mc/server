package client

import (
	"github.com/go-mc/server/world"
)

type Player interface {
	world.Entity
	GameMode() byte
	World() *world.World
	SetWorld(w *world.World)
	ChunkRadius() int32
	// ChunkPos 返回玩家当前在哪个区块
	ChunkPos() [2]int32
}
