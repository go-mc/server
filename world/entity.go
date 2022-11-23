package world

import (
	"math"

	"github.com/df-mc/atomic"
)

var entityCounter atomic.Int32

func NewEntityID() int32 {
	return entityCounter.Inc()
}

type Entity struct {
	EntityID int32
	Position
	Rotation
	OnGround
	pos0 Position
	rot0 Rotation
}

type (
	Position [3]float64
	Rotation [2]float32
	OnGround bool
)

func (e *Entity) getPoint() [2]float64 {
	return [2]float64{e.Position[0], e.Position[2]}
}

func (p *Position) IsValid() bool {
	return !math.IsNaN((*p)[0]) && !math.IsNaN((*p)[1]) && !math.IsNaN((*p)[2]) &&
		!math.IsInf((*p)[0], 0) && !math.IsInf((*p)[1], 0) && !math.IsInf((*p)[2], 0)
}
