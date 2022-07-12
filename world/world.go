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

	// playerViews is a BVH tree，storing the visual range collision boxes of each player.
	// the data structure is used to determine quickly which players to send notify when entity moves.
	playerViews playerViewTree
	players     map[Client]*Player

	chunkProvider ChunkProvider

	tickLock sync.Mutex
}

type playerView struct {
	EntityViewer
	*Player
}
type vec3d = bvh.Vec3[float64]
type aabb3d = bvh.AABB[float64, vec3d]
type playerViewNode = bvh.Node[float64, aabb3d, playerView]
type playerViewTree = bvh.Tree[float64, aabb3d, playerView]

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
	// delete the player from all chunks which load the player.
	for pos := range w.loaders[c].loaded {
		if !w.chunks[pos].RemoveViewer(c) {
			w.log.Panic("viewer is not found in the loaded chunk")
		}
	}
	delete(w.loaders, c)
	delete(w.players, c)
	// delete the player from entity system.
	w.playerViews.Delete(p.view)
	w.playerViews.Find(
		bvh.TouchPoint[vec3d, aabb3d](bvh.Vec3[float64](p.Position)),
		func(n *playerViewNode) bool {
			n.Value.ViewRemoveEntities([]int32{p.EntityID})
			delete(n.Value.EntitiesInView, p.EntityID)
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
