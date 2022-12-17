package client

import (
	"bytes"

	pk "github.com/Tnze/go-mc/net/packet"
)

type (
	clientAcceptTeleportation  struct{}
	clientMovePlayerPos        struct{}
	clientMovePlayerPosRot     struct{}
	clientMovePlayerRot        struct{}
	clientMovePlayerStatusOnly struct{}
	clientMoveVehicle          struct{}
)

func (clientAcceptTeleportation) Handle(p pk.Packet, c *Client) error {
	var TeleportID pk.VarInt
	_, err := TeleportID.ReadFrom(bytes.NewReader(p.Data))
	if err != nil {
		return err
	}
	c.Inputs.Lock()
	c.Inputs.TeleportID = int32(TeleportID)
	c.Inputs.Unlock()
	return nil
}

func (clientMovePlayerPos) Handle(p pk.Packet, c *Client) error {
	var X, FeetY, Z pk.Double
	var OnGround pk.Boolean
	if err := p.Scan(&X, &FeetY, &Z, &OnGround); err != nil {
		return err
	}
	c.Inputs.Lock()
	c.Inputs.Position = [3]float64{float64(X), float64(FeetY), float64(Z)}
	c.Inputs.Unlock()
	return nil
}

func (clientMovePlayerPosRot) Handle(p pk.Packet, c *Client) error {
	var X, FeetY, Z pk.Double
	var Yaw, Pitch pk.Float
	var OnGround pk.Boolean
	if err := p.Scan(&X, &FeetY, &Z, &Yaw, &Pitch, &OnGround); err != nil {
		return err
	}
	c.Inputs.Lock()
	c.Inputs.Position = [3]float64{float64(X), float64(FeetY), float64(Z)}
	c.Inputs.Rotation = [2]float32{float32(Yaw), float32(Pitch)}
	c.Inputs.Unlock()
	return nil
}

func (clientMovePlayerRot) Handle(p pk.Packet, c *Client) error {
	var Yaw, Pitch pk.Float
	var OnGround pk.Boolean
	if err := p.Scan(&Yaw, &Pitch, &OnGround); err != nil {
		return err
	}
	c.Inputs.Lock()
	c.Inputs.Rotation = [2]float32{float32(Yaw), float32(Pitch)}
	c.Inputs.Unlock()
	return nil
}

func (clientMovePlayerStatusOnly) Handle(p pk.Packet, c *Client) error {
	var OnGround pk.UnsignedByte
	if err := p.Scan(&OnGround); err != nil {
		return err
	}
	c.Inputs.Lock()
	c.Inputs.OnGround = OnGround != 0
	c.Inputs.Unlock()
	return nil
}

func (clientMoveVehicle) Handle(_ pk.Packet, _ *Client) error {
	return nil
}
