package client

import pk "github.com/Tnze/go-mc/net/packet"

type clientMovePlayerPos struct{}
type clientMovePlayerPosRot struct{}
type clientMovePlayerRot struct{}
type clientMovePlayerStatusOnly struct{}
type clientMoveVehicle struct{}

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
