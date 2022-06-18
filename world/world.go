package world

import (
	"errors"
	"github.com/Tnze/go-mc/level"
	"go.uber.org/zap"
)

type World struct {
	log *zap.Logger

	chunks   map[[2]int32]*level.Chunk
	chunksRC map[[2]int32]int
	viewers  map[*Loader]Viewer

	provider Provider
}

func New(logger *zap.Logger) (w *World) {
	w = &World{
		log:      logger,
		chunks:   make(map[[2]int32]*level.Chunk),
		provider: Provider{},
	}
	go w.tickLoop()
	return
}

func (w *World) Name() string {
	return "world"
}

func (w *World) HashedSeed() [8]byte {
	return [8]byte{}
}

func (w *World) loadChunk(pos [2]int32) {
	c, err := w.provider.GetChunk(pos)
	if errors.Is(err, errChunkNotExist) {
		w.log.Info("Chunk not found, generating", zap.Int32("x", pos[0]), zap.Int32("z", pos[1]))
		panic("generator not implement")
	}
	w.chunks[pos] = c
}

func (w *World) unloadChunk(pos [2]int32) {
	c, ok := w.chunks[pos]
	if !ok {
		w.log.Panic("Unloading an non-exist chunk", zap.Int32("x", pos[0]), zap.Int32("z", pos[1]))
	}
	err := w.provider.PutChunk(pos, c)
	if err != nil {
		w.log.Error("Store chunk data error", zap.Error(err), zap.Int32("x", pos[0]), zap.Int32("z", pos[1]))
	}
	delete(w.chunks, pos)
}

type Entity interface {
	ID() int32
}
