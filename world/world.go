package world

import (
	"errors"
	"github.com/Tnze/go-mc/level"
	"github.com/go-mc/server/world/internal/bvh"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"sync"
)

type World struct {
	log *zap.Logger

	chunks  map[[2]int32]*LoadedChunk
	loaders map[ChunkViewer]*loader

	// playerViews 是一颗BVH树，储存了世界中每个玩家的可视距离碰撞箱，
	// 该数据结构用于快速判定每个Entity移动时应该向哪些Player发送通知。
	playerViews playerViewTree
	players     map[Client]*Player

	chunkProvider ChunkProvider

	tickLock sync.Mutex
}

type playerView struct {
	EntityViewer
	*Player
}
type playerViewBound = bvh.AABB[float64, bvh.Vec2[float64]]
type playerViewNode = bvh.Node[float64, playerViewBound, playerView]
type playerViewTree = bvh.Tree[float64, bvh.AABB[float64, bvh.Vec2[float64]], playerView]

func New(logger *zap.Logger, provider ChunkProvider) (w *World) {
	w = &World{
		log:           logger,
		chunks:        make(map[[2]int32]*LoadedChunk),
		loaders:       make(map[ChunkViewer]*loader),
		players:       make(map[Client]*Player),
		chunkProvider: provider,
	}
	go w.tickLoop()
	return
}

func (w *World) Name() string {
	return "minecraft:overworld"
}

func (w *World) HashedSeed() [8]byte {
	return [8]byte{}
}

func (w *World) AddPlayer(c Client, p *Player, limiter *rate.Limiter) {
	w.tickLock.Lock()
	defer w.tickLock.Unlock()
	w.loaders[c] = NewLoader(p, limiter)
	w.players[c] = p
	p.view = w.playerViews.Insert(p.getView(), playerView{c, p})
}

func (w *World) RemovePlayer(c Client, p *Player) {
	w.tickLock.Lock()
	defer w.tickLock.Unlock()
	w.log.Debug("Remove Player",
		zap.Int("loader count", len(w.loaders[c].loaded)),
		zap.Int("world count", len(w.chunks)),
	)
	// 从该玩家加载的所有区块中删除该玩家
	for pos := range w.loaders[c].loaded {
		if !w.chunks[pos].RemoveViewer(c) {
			w.log.Panic("viewer is not found in the loaded chunk")
		}
	}
	delete(w.loaders, c)
	delete(w.players, c)
	// 从实体系统中删除该玩家
	w.playerViews.Delete(p.view)
	w.playerViews.Find(
		bvh.TouchPoint[bvh.Vec2[float64], playerViewBound](p.getPoint()),
		func(n *playerViewNode) bool {
			n.Value.ViewRemoveEntities([]int32{p.EntityID})
			return true
		},
	)
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
	viewers []ChunkViewer
	*level.Chunk
}

func (lc *LoadedChunk) AddViewer(v ChunkViewer) {
	lc.Lock()
	defer lc.Unlock()
	//for _, v2 := range lc.viewers {
	//	if v2 == v {
	//		panic("append an exist viewer")
	//	}
	//}
	lc.viewers = append(lc.viewers, v)
}

func (lc *LoadedChunk) RemoveViewer(v ChunkViewer) bool {
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

func sliceDeleteElem[E comparable](slice *[]E, v E) bool {
	for i, v2 := range *slice {
		if v2 == v {
			last := len(*slice) - 1
			(*slice)[i] = (*slice)[last]
			*slice = (*slice)[:last]
			return true
		}
	}
	return false
}
