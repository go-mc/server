package entity

import (
	"io"

	pk "github.com/Tnze/go-mc/net/packet"
)

type MetadataSet []MetadataField

type MetadataField struct {
	Index byte
	MetadataValue
}

func (m MetadataSet) WriteTo(w io.Writer) (n int64, err error) {
	var tmpN int64
	for _, v := range m {
		tmpN, err = pk.UnsignedByte(v.Index).WriteTo(w)
		n += tmpN
		if err != nil {
			return
		}
		tmpN, err = v.WriteTo(w)
		if err != nil {
			return
		}
	}
	tmpN, err = pk.UnsignedByte(0xFF).WriteTo(w)
	return n + tmpN, err
}

func (m *MetadataField) WriteTo(w io.Writer) (n int64, err error) {
	n1, err := pk.VarInt(m.MetadataValue.TypeID()).WriteTo(w)
	if err != nil {
		return n1, err
	}
	n2, err := m.MetadataValue.WriteTo(w)
	return n1 + n2, err
}

type MetadataValue interface {
	TypeID() int32
	pk.Field
}

type (
	Byte struct{ pk.Byte }
	// VarInt       struct{ pk.VarInt }
	// Float        struct{ pk.Float }
	// String       struct{ pk.String }
	// Chat         struct{ chat.Message }
	// OptionalChat struct {
	// 	Exist   bool
	// 	Message chat.Message
	// }
	// Slot     struct{}
	// Boolean  struct{ pk.Boolean }
	// Rotation [3]pk.Float
	// Position struct{ pk.Position }

	Pose int32
)

func (b *Byte) TypeID() int32 { return 0 }
func (p *Pose) TypeID() int32 { return 18 }

const (
	Standing Pose = iota
	FallFlying
	Sleeping
	Swimming
	SpinAttack
	Crouching
	LongJumping
	Dying
	Croaking
	UsingTongue
	Roaring
	Sniffing
	Emerging
	Digging
)

func (p Pose) WriteTo(w io.Writer) (n int64, err error)   { return pk.VarInt(p).WriteTo(w) }
func (p *Pose) ReadFrom(r io.Reader) (n int64, err error) { return (*pk.VarInt)(p).ReadFrom(r) }
