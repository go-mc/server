package world

import (
	"errors"
	"github.com/Tnze/go-mc/level"
	"go.uber.org/zap"
	"sync"
)

type World struct {
	log *zap.Logger

	chunks    map[[2]int32]*LoadedChunk
	chunksRC  map[[2]int32]int
	loaders   map[*Loader]Viewer
	loadersMu sync.Mutex

	provider Provider
}

func New(logger *zap.Logger, provider Provider) (w *World) {
	w = &World{
		log:      logger,
		chunks:   make(map[[2]int32]*LoadedChunk),
		chunksRC: make(map[[2]int32]int),
		loaders:  make(map[*Loader]Viewer),
		provider: provider,
	}
	spawnLoader := NewLoader(w, [2]int32{0, 0}, 20)
	w.loaders[spawnLoader] = nil
	go w.tickLoop()
	return
}

func (w *World) Name() string {
	return "world"
}

func (w *World) HashedSeed() [8]byte {
	return [8]byte{}
}

func (w *World) AddLoader(l *Loader, viewer Viewer) {
	w.loadersMu.Lock()
	defer w.loadersMu.Unlock()
	w.loaders[l] = viewer
}

func (w *World) loadChunk(pos [2]int32) {
	logger := w.log.With(zap.Int32("x", pos[0]), zap.Int32("z", pos[1]))
	c, err := w.provider.GetChunk(pos)
	if errors.Is(err, errChunkNotExist) {
		logger.Debug("Generate chunk")
		// TODO: 目前还没有区块生成器，这里仅生成了一个空区块,然后将区块标记为已生成
		c = level.EmptyChunk(24)
		//bedrock := block.ToStateID[block.Bedrock{}]
		//for i := 0; i < 16*16; i++ {
		//	c.Sections[0].SetBlock(i, bedrock)
		//}
		c.Status = level.StatusFull
	}
	logger.Debug("Load chunk")
	w.chunks[pos] = &LoadedChunk{Chunk: c}
}

func (w *World) unloadChunk(pos [2]int32) {
	logger := w.log.With(zap.Int32("x", pos[0]), zap.Int32("z", pos[1]))
	logger.Debug("Unload chunk")
	c, ok := w.chunks[pos]
	if !ok {
		logger.Panic("Unloading an non-exist chunk")
	}
	// 通知所有监视该区块的viewer卸载该区块
	for _, viewer := range c.viewers {
		viewer.ViewChunkUnload(pos)
	}
	// 将该区块交给provider保存
	err := w.provider.PutChunk(pos, c.Chunk)
	if err != nil {
		logger.Error("Store chunk data error", zap.Error(err))
	}
	delete(w.chunks, pos)
}

type LoadedChunk struct {
	viewers []Viewer
	*level.Chunk
}

type Entity interface {
	ID() int32
}
