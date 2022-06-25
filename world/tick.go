package world

import (
	"github.com/Tnze/go-mc/chat"
	"github.com/go-mc/server/world/internal/bvh"
	"go.uber.org/zap"
	"math"
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
	// TODO: 这里本来应该遍历实体列表，但是目前只有玩家是实体
	for _, e := range w.players {
		// 当实体移动时，向每个能看到它的玩家发送实体移动数据包
		var delta [3]int16
		var rot [2]int8
		if e.Position != e.pos0 { // TODO: 当实体移动距离大于8，改为发送实体传送数据包
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
		cond := bvh.TouchPoint[bvh.Vec2[float64], playerViewBound](e.getPoint())
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
					return true // 不向玩家自己发送自己的移动
				}
				// 检查当前实体是否在玩家的显示列表内，如果存在则转发移动数据
				if _, ok := n.Value.EntitiesInView[e.EntityID]; ok {
					sendMove(n.Value.EntityViewer)
				} else {
					// 否则，将该实体添加到玩家的实体列表
					// TODO: 处理实体不是玩家的情况
					n.Value.ViewAddPlayer(e)
					n.Value.EntitiesInView[e.EntityID] = &e.Entity
				}
				return true
			},
		)
	}
	// 从每个玩家的实体列表中删除不再在范围内的实体
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
