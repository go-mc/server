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
	pl.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		cc := c.(*client.Client)
		cc.SendPlayerInfoAdd([]*world.Player{p})
		players = append(players, cc.GetPlayer())
	})
	c.SendPlayerInfoAdd(players)
}

func (pl *playerList) updateLatency(c *client.Client, latency time.Duration) {
	p := c.GetPlayer()
	pl.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		cc := c.(*client.Client)
		cc.SendPlayerInfoUpdateLatency(p, latency)
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
