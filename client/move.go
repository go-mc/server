// This file is part of go-mc/server project.
// Copyright (C) 2023.  Tnze
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package client

import (
	"bytes"

	pk "github.com/Tnze/go-mc/net/packet"
)

func clientAcceptTeleportation(p pk.Packet, c *Client) error {
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

func clientMovePlayerPos(p pk.Packet, c *Client) error {
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

func clientMovePlayerPosRot(p pk.Packet, c *Client) error {
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

func clientMovePlayerRot(p pk.Packet, c *Client) error {
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

func clientMovePlayerStatusOnly(p pk.Packet, c *Client) error {
	var OnGround pk.UnsignedByte
	if err := p.Scan(&OnGround); err != nil {
		return err
	}
	c.Inputs.Lock()
	c.Inputs.OnGround = OnGround != 0
	c.Inputs.Unlock()
	return nil
}

func clientMoveVehicle(_ pk.Packet, _ *Client) error {
	return nil
}
