package world

type World struct {
}

func (w *World) Name() string {
	return "world"
}

func (w *World) HashedSeed() [8]byte {
	return [8]byte{}
}

type Entity interface {
	ID() int32
}
