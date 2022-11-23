package world

import (
	"bytes"
	_ "embed"

	"github.com/Tnze/go-mc/nbt"
)

//go:embed RegistryCodec.nbt
var dimensionCodecBytes []byte
var RegistryCodec nbt.RawMessage

func init() {
	r := bytes.NewReader(dimensionCodecBytes)
	d := nbt.NewDecoder(r)
	_, err := d.Decode(&RegistryCodec)
	if err != nil {
		panic(err)
	}
}

func (w *World) DimensionCodec() nbt.RawMessage {
	return RegistryCodec
}
