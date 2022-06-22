package world

import (
	"errors"
	"github.com/Tnze/go-mc/level"
	"github.com/go-mc/server/player"
	"go.uber.org/zap"
	"sync"
)

type World struct {
	log *zap.Logger

	chunks  map[[2]int32]*LoadedChunk
	loaders map[Viewer]*loader

	chunkProvider Provider

	tickLock sync.Mutex
}

func New(logger *zap.Logger, provider Provider) (w *World) {
	w = &World{
		log:           logger,
		chunks:        make(map[[2]int32]*LoadedChunk),
		loaders:       make(map[Viewer]*loader),
		chunkProvider: provider,
	}
	//spawnLoader := NewLoader(w, spawnPoint{[2]int32{0, 0}, 20})
	//w.loaders[spawnLoader] = nil
	go w.tickLoop()
	return
}

type spawnPoint struct {
	pos [2]int32
	r   int32
}

func (s spawnPoint) ChunkPos() [2]int32 { return s.pos }
func (s spawnPoint) ChunkRadius() int32 { return s.r }

func (w *World) Name() string {
	return "minecraft:overworld"
}

func (w *World) HashedSeed() [8]byte {
	return [8]byte{}
}

func (w *World) AddPlayer(v Viewer, p *player.Player) {
	w.tickLock.Lock()
	defer w.tickLock.Unlock()
	w.loaders[v] = NewLoader(p)
}

func (w *World) RemovePlayer(v Viewer) {
	w.tickLock.Lock()
	defer w.tickLock.Unlock()
	w.log.Debug("Remove Player",
		zap.Int("loader count", len(w.loaders[v].loaded)),
		zap.Int("world count", len(w.chunks)),
	)
	for pos := range w.loaders[v].loaded {
		if !w.chunks[pos].RemoveViewer(v) {
			w.log.Panic("viewer is not found in the loaded chunk")
		}
	}
	delete(w.loaders, v)
}

func (w *World) loadChunk(pos [2]int32) bool {
	logger := w.log.With(zap.Int32("x", pos[0]), zap.Int32("z", pos[1]))
	//logger.Debug("Load chunk")
	c, err := w.chunkProvider.GetChunk(pos)
	if errors.Is(err, errChunkNotExist) {
		logger.Debug("Generate chunk")
		// TODO: because there is no chunk generator，generate an empty chunk and mark it as generated
		c = level.EmptyChunk(24)
		c.Status = level.StatusFull
	} else if err != nil {
		return false
	}
	w.chunks[pos] = &LoadedChunk{Chunk: c}
	return true
}

func (w *World) unloadChunk(pos [2]int32) {
	logger := w.log.With(zap.Int32("x", pos[0]), zap.Int32("z", pos[1]))
	//logger.Debug("Unload chunk")
	c, ok := w.chunks[pos]
	if !ok {
		logger.Panic("Unloading an non-exist chunk")
	}
	// notify all viewers who are watching the chunk to unload the chunk
	for _, viewer := range c.viewers {
		viewer.ViewChunkUnload(pos)
	}
	// move the chunk to provider and save
	err := w.chunkProvider.PutChunk(pos, c.Chunk)
	if err != nil {
		logger.Error("Store chunk data error", zap.Error(err))
	}
	delete(w.chunks, pos)
}

type LoadedChunk struct {
	sync.Mutex
	viewers []Viewer
	*level.Chunk
}

func (lc *LoadedChunk) AddViewer(v Viewer) {
	lc.Lock()
	defer lc.Unlock()
	//for _, v2 := range lc.viewers {
	//	if v2 == v {
	//		panic("append an exist viewer")
	//	}
	//}
	lc.viewers = append(lc.viewers, v)
}

func (lc *LoadedChunk) RemoveViewer(v Viewer) bool {
	lc.Lock()
	defer lc.Unlock()
	for i, v2 := range lc.viewers {
		if v2 == v {
			last := len(lc.viewers) - 1
			lc.viewers[i] = lc.viewers[last]
			lc.viewers = lc.viewers[:last]
			return true
		}
	}
	return false
}

type Entity interface {
	ID() int32
}
