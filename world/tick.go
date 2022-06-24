package world

import (
	"github.com/go-mc/server/world/internal/bvh"
	"time"
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
	for _, p := range w.players {
		p.ChunkPos = [2]int32{
			int32(p.Position[0]) >> 5,
			int32(p.Position[2]) >> 5,
		}
	}
	// 由于w.loaders的遍历顺序是随机的，所以每个loader每次都有相同的机会，相对比较公平
	for viewer, loader := range w.loaders {
		loader.calcLoadingQueue()
		for _, pos := range loader.loadQueue {
			if _, ok := w.chunks[pos]; !ok {
				if !w.loadChunk(pos) {
					continue
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
	for _, p := range w.players {
		p.pos0 = p.nextPos.Load()
		p.rot0 = p.nextRot.Load()
	}
}

func (w *World) subtickUpdateEntities() {
	for _, e := range w.players {
		// 当实体移动时，向每个能看到它的玩家发送实体移动数据包
		var sendMove func(viewer EntityViewer)
		rot := [2]int8{
			int8(e.rot0[0] * 256 / 360),
			int8(e.rot0[1] * 256 / 360),
		}
		if e.Position != e.pos0 {
			delta := [3]int16{
				int16((e.pos0[0] - e.Position[0]) * 32 * 128),
				int16((e.pos0[1] - e.Position[1]) * 32 * 128),
				int16((e.pos0[2] - e.Position[2]) * 32 * 128),
			}
			e.Position = e.pos0
			// TODO: 当实体移动距离大于8，改为发送实体传送数据包
			if e.Rotation != e.rot0 {
				sendMove = func(viewer EntityViewer) {
					viewer.ViewMoveEntityPosAndRot(e.EntityID, delta, rot, false)
				}
				e.Rotation = e.rot0
			} else {
				sendMove = func(viewer EntityViewer) {
					viewer.ViewMoveEntityPos(e.EntityID, delta, false)
				}
			}
		} else if e.Rotation != e.rot0 {
			sendMove = func(viewer EntityViewer) {
				viewer.ViewMoveEntityRot(e.EntityID, rot, false)
			}
			e.Rotation = e.rot0
		}
		if sendMove != nil {
			w.playerViews.Find(
				bvh.TouchPoint[bvh.Vec2[float64], playerViewBound](e.getPoint()),
				func(n *playerViewNode) bool {
					if n.Value.Player == e {
						return true // 不向玩家自己发送自己的移动
					}
					// 检查当前实体是否在玩家的显示列表内，如果存在则转发移动数据
					if _, ok := n.Value.EntitiesInView[e.EntityID]; ok {
						sendMove(n.Value)
					} else {
						n.Value.ViewAddPlayer(e)
						n.Value.EntitiesInView[e.EntityID] = &e.Entity
					}
					return true
				},
			)
		}
	}
	for _, p := range w.players {
		p.view = w.playerViews.Insert(p.getView(), w.playerViews.Delete(p.view))
		for id, e := range p.EntitiesInView {
			if !p.view.Box.WithIn(e.getPoint()) {
				delete(p.EntitiesInView, id) // 从正在遍历的map中删除元素应该是安全的
				p.view.Value.ViewRemoveEntities([]int32{id})
			}
		}
	}
}
