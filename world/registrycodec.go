package world

import (
	"bytes"
	_ "embed"

	"github.com/Tnze/go-mc/registry"

	"github.com/Tnze/go-mc/nbt"
)

//go:embed RegistryCodec.nbt
var networkCodecBytes []byte
var NetworkCodec registry.NetworkCodec

func init() {
	r := bytes.NewReader(networkCodecBytes)
	d := nbt.NewDecoder(r)
	_, err := d.Decode(&NetworkCodec)
	if err != nil {
		panic(err)
	}
}

func (w *World) NetworkCodec() registry.NetworkCodec {
	return NetworkCodec
}
