package world

import (
	"bytes"
	_ "embed"

	"github.com/Tnze/go-mc/nbt"
)

//go:embed DimensionCodec.nbt
var dimensionCodecBytes []byte
var DimensionCodec nbt.RawMessage

func init() {
	r := bytes.NewReader(dimensionCodecBytes)
	d := nbt.NewDecoder(r)
	_, err := d.Decode(&DimensionCodec)
	if err != nil {
		panic(err)
	}
}

func (w *World) DimensionCodec() nbt.RawMessage {
	return DimensionCodec
}
