// This file is part of go-mc/server project.
// Copyright (C) 2022.  Tnze
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package game

import (
	"time"

	"github.com/Tnze/go-mc/data/packetid"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/server"
	"github.com/go-mc/server/client"
	"github.com/go-mc/server/world"
)

type playerList struct {
	keepAlive *server.KeepAlive
	pingList  *server.PlayerList
}

func (pl *playerList) addPlayer(c *client.Client, p *world.Player) {
	pl.pingList.ClientJoin(c, server.PlayerSample{
		Name: p.Name,
		ID:   p.UUID,
	})
	pl.keepAlive.ClientJoin(c)
	c.AddHandler(packetid.ServerboundKeepAlive, keepAliveHandler{pl.keepAlive})
	players := make([]*world.Player, 0, pl.pingList.Len()+1)
	players = append(players, p)
	addPlayerAction := client.NewPlayerInfoAction(
		client.PlayerInfoAddPlayer,
		client.PlayerInfoUpdateListed,
	)
	pl.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		cc := c.(*client.Client)
		cc.SendPlayerInfoUpdate(addPlayerAction, []*world.Player{p})
		players = append(players, cc.GetPlayer())
	})
	c.SendPlayerInfoUpdate(addPlayerAction, players)
}

func (pl *playerList) updateLatency(c *client.Client, latency time.Duration) {
	updateLatencyAction := client.NewPlayerInfoAction(client.PlayerInfoUpdateLatency)
	p := c.GetPlayer()
	p.Inputs.Lock()
	p.Inputs.Latency = latency
	p.Inputs.Unlock()
	pl.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		c.(*client.Client).SendPlayerInfoUpdate(updateLatencyAction, []*world.Player{p})
	})
}

func (pl *playerList) removePlayer(c *client.Client) {
	pl.pingList.ClientLeft(c)
	pl.keepAlive.ClientLeft(c)
	p := c.GetPlayer()
	pl.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		c.(*client.Client).SendPlayerInfoRemove([]*world.Player{p})
	})
}

type keepAliveHandler struct{ *server.KeepAlive }

func (k keepAliveHandler) Handle(p pk.Packet, c *client.Client) error {
	var req pk.Long
	if err := p.Scan(&req); err != nil {
		return err
	}
	k.KeepAlive.ClientTick(c)
	return nil
}
