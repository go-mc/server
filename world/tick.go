package world

import (
	"math"
	"time"

	"github.com/Tnze/go-mc/chat"
	"github.com/go-mc/server/world/internal/bvh"
	"go.uber.org/zap"
)

func (w *World) tickLoop() {
	var n uint
	for range time.Tick(time.Microsecond * 20) {
		w.tick(n)
		n++
	}
}

func (w *World) tick(n uint) {
	w.tickLock.Lock()
	defer w.tickLock.Unlock()

	if n%8 == 0 {
		w.subtickChunkLoad()
	}
	w.subtickUpdatePlayers()
	w.subtickUpdateEntities()
}

func (w *World) subtickChunkLoad() {
	for c, p := range w.players {
		x := int32(p.Position[0]) >> 4
		y := int32(p.Position[1]) >> 4
		z := int32(p.Position[2]) >> 4
		if newChunkPos := [3]int32{x, y, z}; newChunkPos != p.ChunkPos {
			p.ChunkPos = newChunkPos
			w.log.Debug("SendSetChunkCacheCenter", zap.Int32("x", x), zap.Int32("z", z))
			c.SendSetChunkCacheCenter([2]int32{x, z})
		}
	}
	// because of the random traversal order of w.loaders,every loader has the same opportunity,so it's relatively fair.
LoadChunk:
	for viewer, loader := range w.loaders {
		loader.calcLoadingQueue()
		for _, pos := range loader.loadQueue {
			if !loader.limiter.Allow() { // We reach the player limit. skip
				break
			}
			if _, ok := w.chunks[pos]; !ok {
				if !w.loadChunk(pos) {
					break LoadChunk // We reach the global limit. skip
				}
			}
			loader.loaded[pos] = struct{}{}
			lc := w.chunks[pos]
			lc.AddViewer(viewer)
			lc.Lock()
			viewer.ViewChunkLoad(pos, lc.Chunk)
			lc.Unlock()
		}
	}
	for viewer, loader := range w.loaders {
		loader.calcUnusedChunks()
		for _, pos := range loader.unloadQueue {
			delete(loader.loaded, pos)
			if !w.chunks[pos].RemoveViewer(viewer) {
				w.log.Panic("viewer is not found in the loaded chunk")
			}
			viewer.ViewChunkUnload(pos)
		}
	}
	var unloadQueue [][2]int32
	for pos, chunk := range w.chunks {
		if len(chunk.viewers) == 0 {
			unloadQueue = append(unloadQueue, pos)
		}
	}
	for i := range unloadQueue {
		w.unloadChunk(unloadQueue[i])
	}
}

func (w *World) subtickUpdatePlayers() {
	for c, p := range w.players {
		// update the range of visual.
		p.view = w.playerViews.Insert(p.getView(), w.playerViews.Delete(p.view))
		// delete entities that not in range from entities lists of each player.
		for id, e := range p.EntitiesInView {
			if !p.view.Box.WithIn(vec3d(e.Position)) {
				delete(p.EntitiesInView, id) // it should be safe to delete element from a map being traversed.
				p.view.Value.ViewRemoveEntities([]int32{id})
			}
		}
		if p.teleport != nil {
			if p.acceptTeleportID.Load() == p.teleport.ID {
				p.pos0 = p.teleport.Position
				p.rot0 = p.teleport.Rotation
				p.teleport = nil
			}
		} else {
			pos := p.nextPos.Load()
			delta := [3]float64{
				pos[0] - p.Position[0],
				pos[1] - p.Position[1],
				pos[2] - p.Position[2],
			}
			distance := math.Sqrt(delta[0]*delta[0] + delta[1]*delta[1] + delta[2]*delta[2])
			if distance > 100 {
				w.log.Info("Player move too quickly", zap.Float64("delta", distance))
				// You moved too quickly :( (Hacking?)
				teleportID := c.SendPlayerPosition(p.Position, p.Rotation, true)
				p.teleport = &TeleportRequest{
					ID:       teleportID,
					Position: p.Position,
					Rotation: p.Rotation,
				}
			} else if pos.IsValid() {
				p.pos0 = pos
				p.rot0 = p.nextRot.Load()
				p.OnGround = OnGround(p.nextOnGround.Load())
			} else {
				w.log.Info("Player move invalid", zap.Float64("x", pos[0]), zap.Float64("y", pos[1]), zap.Float64("z", pos[2]))
				c.SendDisconnect(chat.TranslateMsg("multiplayer.disconnect.invalid_player_movement"))
			}
		}
	}
}

func (w *World) subtickUpdateEntities() {
	// TODO: entity list should be traversed here, but players are the only entities now.
	for _, e := range w.players {
		// sending Update Entity Position pack to every player who can see it, when it moves.
		var delta [3]int16
		var rot [2]int8
		if e.Position != e.pos0 { // TODO: send Teleport Entity pack instead when moving distance is greater than 8.
			delta = [3]int16{
				int16((e.pos0[0] - e.Position[0]) * 32 * 128),
				int16((e.pos0[1] - e.Position[1]) * 32 * 128),
				int16((e.pos0[2] - e.Position[2]) * 32 * 128),
			}
		}
		if e.Rotation != e.rot0 {
			rot = [2]int8{
				int8(e.rot0[0] * 256 / 360),
				int8(e.rot0[1] * 256 / 360),
			}
		}
		cond := bvh.TouchPoint[vec3d, aabb3d](vec3d(e.Position))
		w.playerViews.Find(cond,
			func(n *playerViewNode) bool {
				if n.Value.Player == e {
					return true // don't send the player self to the player
				}
				// check if current entity is in range of player visual. if so, moving data will be forwarded.
				if _, ok := n.Value.EntitiesInView[e.EntityID]; !ok {
					// add the entity to the entities list of the player
					n.Value.ViewAddPlayer(e)
					n.Value.EntitiesInView[e.EntityID] = &e.Entity
				}
				return true
			},
		)
		var sendMove func(v EntityViewer)
		switch {
		case e.Position != e.pos0 && e.Rotation != e.rot0:
			sendMove = func(v EntityViewer) {
				v.ViewMoveEntityPosAndRot(e.EntityID, delta, rot, bool(e.OnGround))
				v.ViewRotateHead(e.EntityID, rot[0])
			}
		case e.Position != e.pos0:
			sendMove = func(v EntityViewer) {
				v.ViewMoveEntityPos(e.EntityID, delta, bool(e.OnGround))
			}
		case e.Rotation != e.rot0:
			sendMove = func(v EntityViewer) {
				v.ViewMoveEntityRot(e.EntityID, rot, bool(e.OnGround))
				v.ViewRotateHead(e.EntityID, rot[0])
			}
		default:
			continue
		}
		e.Position = e.pos0
		e.Rotation = e.rot0
		w.playerViews.Find(cond,
			func(n *playerViewNode) bool {
				if n.Value.Player == e {
					return true // not sending self movements to player self.
				}
				// check if current entity is in the player visual entities list. if so, moving data will be forwarded.
				if _, ok := n.Value.EntitiesInView[e.EntityID]; ok {
					sendMove(n.Value.EntityViewer)
				} else {
					// or the entity will be add to the entities list of the player
					// TODO: deal with the situation that the entity is not a player
					n.Value.ViewAddPlayer(e)
					n.Value.EntitiesInView[e.EntityID] = &e.Entity
				}
				return true
			},
		)
	}
}
