package game

import (
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/go-mc/server/client"
	"go.uber.org/zap"
)

type chatPreview struct {
	log *zap.Logger
}

func (cp *chatPreview) Handle(p pk.Packet, c *client.Client) error {
	var (
		QueryID pk.Int
		Message pk.String
	)
	if err := p.Scan(&QueryID, &Message); err != nil {
		return err
	}
	player := c.GetPlayer()
	cp.log.Debug("Preview msg",
		zap.String("player", player.Name),
		zap.Int32("query id", int32(QueryID)),
		zap.String("msg", string(Message)),
	)
	c.SendChatPreview(int32(QueryID), nil)
	return nil
}
