package world

import (
	"github.com/df-mc/atomic"
)

var entityCounter atomic.Int32

func NewEntityID() int32 {
	return entityCounter.Inc()
}

type Entity struct {
	EntityID int32
	Position
	pos0 Position
	Rotation
	rot0 Rotation
}

type Position [3]float64

type Rotation [2]float32

func (e *Entity) getPoint() [2]float64 {
	return [2]float64{e.Position[0], e.Position[2]}
}
