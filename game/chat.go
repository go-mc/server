package game

import (
	"github.com/Tnze/go-mc/chat"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/server"
	"github.com/go-mc/server/client"
	"go.uber.org/zap"
	"time"
)

type globalChat struct {
	log     *zap.Logger
	players *playerList
}

func (g *globalChat) broadcastSystemChat(msg chat.Message, typeID chat.Type) {
	g.log.Info("System message", zap.String("chat", msg.ClearString()), zap.Int32("type", int32(typeID)))
	g.players.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		c.(*client.Client).SendSystemChat(msg, typeID)
	})
}

func (g *globalChat) Handle(p pk.Packet, c *client.Client) error {
	var (
		message       pk.String
		timestamp     pk.Long
		salt          pk.Long
		signature     pk.ByteArray
		signedPreview pk.Boolean
	)
	err := p.Scan(&message, &timestamp, &salt, &signature, &signedPreview)
	if err != nil {
		return err
	}
	player := c.GetPlayer()
	unsignedMsg := chat.Text(string(message))
	g.log.Info(
		"Player message",
		zap.String("sender", player.Name),
		zap.String("chat", string(message)),
		zap.Time("timestamp", time.UnixMilli(int64(timestamp))),
	)
	g.players.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		c.(*client.Client).SendPlayerChat(
			player,
			string(message),
			&unsignedMsg,
			chat.Chat,
			int64(timestamp),
			int64(salt),
			signature,
		)
	})
	return nil
}
