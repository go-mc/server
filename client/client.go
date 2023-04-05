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
	"go.uber.org/zap"

	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/net/queue"
	"github.com/Tnze/go-mc/server"
	"github.com/go-mc/server/world"
)

type Client struct {
	log      *zap.Logger
	conn     *net.Conn
	player   *world.Player
	queue    server.PacketQueue
	handlers []packetHandler
	// pointer to the Player.Input
	*world.Inputs
}

type packetHandler interface {
	Handle(p pk.Packet, c *Client) error
}

func New(log *zap.Logger, conn *net.Conn, player *world.Player) *Client {
	return &Client{
		log:      log,
		conn:     conn,
		player:   player,
		queue:    queue.NewChannelQueue[pk.Packet](256),
		handlers: defaultHandlers[:],
		Inputs:   &player.Inputs,
	}
}

func (c *Client) Start() {
	stopped := make(chan struct{}, 2)
	done := func() {
		stopped <- struct{}{}
	}
	// Exit when any error is thrown
	go c.startSend(done)
	go c.startReceive(done)
	<-stopped
}

func (c *Client) startSend(done func()) {
	defer done()
	for {
		p, ok := c.queue.Pull()
		if !ok {
			return
		}
		err := c.conn.WritePacket(p)
		if err != nil {
			c.log.Debug("Send packet fail", zap.Error(err))
			return
		}
		if packetid.ClientboundPacketID(p.ID) == packetid.ClientboundDisconnect {
			return
		}
	}
}

func (c *Client) startReceive(done func()) {
	defer done()
	var packet pk.Packet
	for {
		err := c.conn.ReadPacket(&packet)
		if err != nil {
			c.log.Debug("Receive packet fail", zap.Error(err))
			return
		}
		if packet.ID < 0 || packet.ID >= int32(len(c.handlers)) {
			c.log.Debug("Invalid packet id", zap.Int32("id", packet.ID), zap.Int("len", len(packet.Data)))
			return
		}
		if handler := c.handlers[packet.ID]; handler != nil {
			err = handler.Handle(packet, c)
			if err != nil {
				c.log.Error("Handle packet error", zap.Int32("id", packet.ID), zap.Error(err))
				return
			}
		}
	}
}

func (c *Client) AddHandler(id packetid.ServerboundPacketID, handler packetHandler) {
	c.handlers[id] = handler
}
func (c *Client) GetPlayer() *world.Player { return c.player }

var defaultHandlers = [packetid.ServerboundPacketIDGuard]packetHandler{
	packetid.ServerboundAcceptTeleportation:  clientAcceptTeleportation{},
	packetid.ServerboundClientInformation:    clientInformation{},
	packetid.ServerboundMovePlayerPos:        clientMovePlayerPos{},
	packetid.ServerboundMovePlayerPosRot:     clientMovePlayerPosRot{},
	packetid.ServerboundMovePlayerRot:        clientMovePlayerRot{},
	packetid.ServerboundMovePlayerStatusOnly: clientMovePlayerStatusOnly{},
	packetid.ServerboundMoveVehicle:          clientMoveVehicle{},
}
