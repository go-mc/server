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

	chunks   map[[2]int32]*LoadedChunk
	chunksRC map[[2]int32]int
	loaders  map[*loader]Viewer

	chunkProvider Provider

	tickLock sync.Mutex
}

func New(logger *zap.Logger, provider Provider) (w *World) {
	w = &World{
		log:           logger,
		chunks:        make(map[[2]int32]*LoadedChunk),
		chunksRC:      make(map[[2]int32]int),
		loaders:       make(map[*loader]Viewer),
		chunkProvider: provider,
	}
	spawnLoader := NewLoader(w, spawnPoint{[2]int32{0, 0}, 20})
	w.loaders[spawnLoader] = nil
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
	w.loaders[NewLoader(w, p)] = v
}

func (w *World) loadChunk(pos [2]int32) {
	logger := w.log.With(zap.Int32("x", pos[0]), zap.Int32("z", pos[1]))
	c, err := w.chunkProvider.GetChunk(pos)
	if errors.Is(err, errChunkNotExist) {
		logger.Debug("Generate chunk")
		// TODO: 目前还没有区块生成器，生成一个空区块,然后将区块标记为已生成
		c = level.EmptyChunk(24)
		//bedrock := block.ToStateID[block.Bedrock{}]
		//for i := 0; i < 16*16; i++ {
		//	c.Sections[0].SetBlock(i, bedrock)
		//}
		c.Status = level.StatusFull
	} else if err != nil {
		logger.Panic("Load chunk error", zap.Error(err))
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

type Entity interface {
	ID() int32
}
