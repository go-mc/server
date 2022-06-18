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
	for pos, count := range w.chunksRC {
		if count == 0 {
			w.unloadChunk(pos)
			delete(w.chunksRC, pos)
		}
	}
}
