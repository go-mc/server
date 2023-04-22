// This file is part of go-mc/server project.
// Copyright (C) 2023.  Tnze
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package world

import (
	"errors"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/level/block"
	"github.com/go-mc/server/world/internal/bvh"
)

type World struct {
	log           *zap.Logger
	config        Config
	chunkProvider ChunkProvider

	chunks   map[[2]int32]*LoadedChunk
	loaders  map[ChunkViewer]*loader
	tickLock sync.Mutex

	// playerViews is a BVH tree，storing the visual range collision boxes of each player.
	// the data structure is used to determine quickly which players to send notify when entity moves.
	playerViews playerViewTree
	players     map[Client]*Player
}

type Config struct {
	ViewDistance  int32
	SpawnAngle    float32
	SpawnPosition [3]int32
}

type playerView struct {
	EntityViewer
	*Player
}

type (
	vec3d          = bvh.Vec3[float64]
	aabb3d         = bvh.AABB[float64, vec3d]
	playerViewNode = bvh.Node[float64, aabb3d, playerView]
	playerViewTree = bvh.Tree[float64, aabb3d, playerView]
)

func New(logger *zap.Logger, provider ChunkProvider, config Config) (w *World) {
	w = &World{
		log:           logger,
		config:        config,
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

func (w *World) SpawnPositionAndAngle() ([3]int32, float32) {
	return w.config.SpawnPosition, w.config.SpawnAngle
}

func (w *World) HashedSeed() [8]byte {
	return [8]byte{}
}

func (w *World) AddPlayer(c Client, p *Player, limiter *rate.Limiter) {
	w.tickLock.Lock()
	defer w.tickLock.Unlock()
	w.loaders[c] = newLoader(p, limiter)
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
	logger.Debug("Loading chunk")
	c, err := w.chunkProvider.GetChunk(pos)
	if err != nil {
		if errors.Is(err, errChunkNotExist) {
			logger.Debug("Generate chunk")
			// TODO: because there is no chunk generator，generate an empty chunk and mark it as generated
			c = level.EmptyChunk(24)
			stone := block.ToStateID[block.Stone{}]
			for s := range c.Sections {
				for i := 0; i < 16*16*16; i++ {
					c.Sections[s].SetBlock(i, stone)
				}
			}
			c.Status = level.StatusFull
		} else if !errors.Is(err, ErrReachRateLimit) {
			logger.Error("GetChunk error", zap.Error(err))
			return false
		}
	}
	w.chunks[pos] = &LoadedChunk{Chunk: c}
	return true
}

func (w *World) unloadChunk(pos [2]int32) {
	logger := w.log.With(zap.Int32("x", pos[0]), zap.Int32("z", pos[1]))
	logger.Debug("Unloading chunk")
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
	for _, v2 := range lc.viewers {
		if v2 == v {
			panic("append an exist viewer")
		}
	}
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
