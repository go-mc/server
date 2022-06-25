package client

import (
	"bytes"
	pk "github.com/Tnze/go-mc/net/packet"
)

type clientAcceptTeleportation struct{}
type clientMovePlayerPos struct{}
type clientMovePlayerPosRot struct{}
type clientMovePlayerRot struct{}
type clientMovePlayerStatusOnly struct{}
type clientMoveVehicle struct{}

func (clientAcceptTeleportation) Handle(p pk.Packet, c *Client) error {
	var TeleportID pk.VarInt
	_, err := TeleportID.ReadFrom(bytes.NewReader(p.Data))
	if err != nil {
		return err
	}
	c.player.AcceptTeleport(int32(TeleportID))
	return nil
}

func (clientMovePlayerPos) Handle(p pk.Packet, c *Client) error {
	var X, FeetY, Z pk.Double
	var OnGround pk.Boolean
	if err := p.Scan(&X, &FeetY, &Z, &OnGround); err != nil {
		return err
	}
	c.player.SetNextPosition([3]float64{float64(X), float64(FeetY), float64(Z)})
	return nil
}
func (clientMovePlayerPosRot) Handle(p pk.Packet, c *Client) error {
	var X, FeetY, Z pk.Double
	var Yaw, Pitch pk.Float
	var OnGround pk.Boolean
	if err := p.Scan(&X, &FeetY, &Z, &Yaw, &Pitch, &OnGround); err != nil {
		return err
	}
	c.player.SetNextPosition([3]float64{float64(X), float64(FeetY), float64(Z)})
	c.player.SetNextRotation([2]float32{float32(Yaw), float32(Pitch)})
	return nil
}
func (clientMovePlayerRot) Handle(p pk.Packet, c *Client) error {
	var Yaw, Pitch pk.Float
	var OnGround pk.Boolean
	if err := p.Scan(&Yaw, &Pitch, &OnGround); err != nil {
		return err
	}
	c.player.SetNextRotation([2]float32{float32(Yaw), float32(Pitch)})
	return nil
}
func (clientMovePlayerStatusOnly) Handle(p pk.Packet, c *Client) error {
	return nil
}
func (clientMoveVehicle) Handle(p pk.Packet, c *Client) error {
	return nil
}
