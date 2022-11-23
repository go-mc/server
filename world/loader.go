package world

import (
	"math"
	"sort"

	"golang.org/x/time/rate"
)

// loader take part in chunk loading，each loader contains a position 'pos' and a radius 'r'
// chunks pointed by the position and the radius of loader will be load。
type loader struct {
	loaderSource
	loaded      map[[2]int32]struct{}
	loadQueue   [][2]int32
	unloadQueue [][2]int32
	limiter     *rate.Limiter
}

type loaderSource interface {
	ChunkPosition() [2]int32
	ChunkRadius() int32
}

func NewLoader(source loaderSource, limiter *rate.Limiter) (l *loader) {
	l = &loader{
		loaderSource: source,
		loaded:       make(map[[2]int32]struct{}),
		limiter:      limiter,
	}
	l.calcLoadingQueue()
	return
}

// calcLoadingQueue calculate the chunks which loader point.
// The result is stored in l.loadQueue and the previous will be removed.
func (l *loader) calcLoadingQueue() {
	l.loadQueue = l.loadQueue[:0]
	for _, v := range loadList[:radiusIdx[l.ChunkRadius()]] {
		pos := l.ChunkPosition()
		pos[0], pos[1] = pos[0]+v[0], pos[1]+v[1]
		if _, ok := l.loaded[pos]; !ok {
			l.loadQueue = append(l.loadQueue, pos)
		}
	}
}

// calcUnusedChunks calculate the chunks the loader want to remove.
// Behaviour is same with calcLoadingQueue.
func (l *loader) calcUnusedChunks() {
	l.unloadQueue = l.unloadQueue[:0]
	for chunk := range l.loaded {
		player := l.ChunkPosition()
		r := l.ChunkRadius()
		if distance([2]int32{chunk[0] - player[0], chunk[1] - player[1]}) > float64(r) {
			l.unloadQueue = append(l.unloadQueue, chunk)
		}
	}
}

// loadList is chunks in a certain distance of (0, 0), order by euclidean distance
// the more forward the chunk is, the closer it to (0, 0)
var loadList [][2]int32

// radiusIdx[i] is the count of chunks in loadList and the distance of i
var radiusIdx []int

func init() {
	const maxR int32 = 32

	// calculate loadList
	for x := -maxR; x <= maxR; x++ {
		for z := -maxR; z <= maxR; z++ {
			pos := [2]int32{x, z}
			if distance(pos) < float64(maxR) {
				loadList = append(loadList, pos)
			}
		}
	}
	sort.Slice(loadList, func(i, j int) bool {
		return distance(loadList[i]) < distance(loadList[j])
	})

	// calculate radiusIdx
	radiusIdx = make([]int, maxR+1)
	for i, v := range loadList {
		r := int32(math.Ceil(distance(v)))
		if r > maxR {
			break
		}
		radiusIdx[r] = i
	}
}

// distance calculate the euclidean distance that a position to the origin point
func distance(pos [2]int32) float64 {
	return math.Sqrt(float64(pos[0]*pos[0]) + float64(pos[1]*pos[1]))
}
