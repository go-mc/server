package world

import (
	"math"
	"sort"
)

// loader 用于实现世界区块的加载，每个 loader 包含 位置 pos 和一个半径 r
// 位置和半径指示的范围内的区块将被加载。
type loader struct {
	loaderSource
	w *World

	loaded      map[[2]int32]struct{}
	loadQueue   [][2]int32
	unloadQueue [][2]int32
}

type loaderSource interface {
	ChunkPos() [2]int32
	ChunkRadius() int32
}

func NewLoader(w *World, source loaderSource) (l *loader) {
	l = &loader{
		loaderSource: source,
		w:            w,
		loaded:       map[[2]int32]struct{}{},
		loadQueue:    nil,
	}
	l.calcLoadingQueue()
	return
}

func (l *loader) calcLoadingQueue() {
	for _, v := range loadList[:radiusIdx[l.ChunkRadius()]] {
		pos := l.ChunkPos()
		pos[0], pos[1] = pos[0]+v[0], pos[1]+v[1]
		if _, ok := l.loaded[pos]; !ok {
			l.loadQueue = append(l.loadQueue, pos)
		}
	}
}

func (l *loader) calcUnusedChunks() {
	for chunkPos := range l.loaded {
		currentPos := l.ChunkPos()
		diff := [2]int32{chunkPos[0] - currentPos[0], chunkPos[1] - currentPos[1]}
		r := l.ChunkRadius()
		if distance(diff) > float64(r) {
			l.unloadQueue = append(l.unloadQueue, chunkPos)
		}
	}
}

// loadList 是(0, 0)周围一定范围内的区块，按欧几里得距离排序的列表
// 越靠前的区块距离(0, 0)越近，越靠近末尾的区块距离(0, 0)越远
var loadList [][2]int32

// radiusIdx 中下标为i的数n，代表loadList中前n个区块到(0, 0)的距离在i以内
var radiusIdx []int

func init() {
	const maxR int32 = 32

	// 计算 loadList
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

	// 计算 radiusIdx
	radiusIdx = make([]int, maxR+1)
	for i, v := range loadList {
		r := int32(math.Ceil(distance(v)))
		if r > maxR {
			break
		}
		radiusIdx[r] = i
	}
}

// distance 计算一个坐标到原点的欧几里得距离
func distance(pos [2]int32) float64 {
	return math.Sqrt(float64(pos[0]*pos[0]) + float64(pos[1]*pos[1]))
}
