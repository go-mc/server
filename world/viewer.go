package world

import "github.com/Tnze/go-mc/level"

type Viewer interface {
	ChunkViewer
	//ViewEntityMove()
}

type ChunkViewer interface {
	ViewChunkLoad(pos level.ChunkPos, c *level.Chunk)
	ViewChunkUnload(pos level.ChunkPos)
}
