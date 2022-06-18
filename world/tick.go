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
			w.chunks[pos].viewers = append(w.chunks[pos].viewers, viewer)
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
