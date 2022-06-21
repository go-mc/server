package world

import (
	"time"
)

func (w *World) tickLoop() {
	ticker := time.NewTicker(time.Microsecond * 20)
	for {
		select {
		case <-ticker.C:
			w.tick()
		}
	}
}

func (w *World) tick() {
	w.tickLock.Lock()
	defer w.tickLock.Unlock()

	for loader, viewer := range w.loaders {
		loadChunkLimit := 4
		loader.calcLoadingQueue()
		for _, pos := range loader.loadQueue {
			if loadChunkLimit > 0 {
				loadChunkLimit--
			} else {
				break
			}
			loader.loaded[pos] = struct{}{}
			if _, ok := w.chunks[pos]; !ok {
				w.loadChunk(pos)
			}
			w.chunksRC[pos]++
			lc := w.chunks[pos]
			lc.Lock()
			lc.viewers = append(w.chunks[pos].viewers, viewer)
			if viewer != nil {
				viewer.ViewChunkLoad(pos, lc.Chunk)
			}
			lc.Unlock()
		}
		loader.loadQueue = loader.loadQueue[:0]

		loader.calcUnusedChunks()
		for _, pos := range loader.unloadQueue {
			delete(loader.loaded, pos)
			w.chunksRC[pos]--
			lc := w.chunks[pos]
			lc.Lock()
			for i, v := range lc.viewers {
				if v == viewer {
					last := len(lc.viewers) - 1
					lc.viewers[i] = lc.viewers[last]
					lc.viewers = lc.viewers[:last]
					break
				}
			}
			if viewer != nil {
				viewer.ViewChunkUnload(pos)
			}
			lc.Unlock()
		}
		loader.unloadQueue = loader.unloadQueue[:0]
	}
	for pos, count := range w.chunksRC {
		if count == 0 {
			w.unloadChunk(pos)
			delete(w.chunksRC, pos)
		}
	}
}
