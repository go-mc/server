package world

import "time"

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
	for loader, viewer := range w.loaders {
		for _, pos := range loader.loadQueue {
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
	}
	for pos, count := range w.chunksRC {
		if count == 0 {
			w.unloadChunk(pos)
			delete(w.chunksRC, pos)
		}
	}
}
