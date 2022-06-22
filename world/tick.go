package world

import (
	"time"
)

func (w *World) tickLoop() {
	ticker := time.NewTicker(time.Microsecond * 20)
	var n uint
	for {
		select {
		case <-ticker.C:
			w.tick(n)
		}
		n++
	}
}

func (w *World) tick(n uint) {
	w.tickLock.Lock()
	defer w.tickLock.Unlock()

	if n%8 == 0 {
		w.subtickChunkLoad()
	}
}

func (w *World) subtickChunkLoad() {
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
