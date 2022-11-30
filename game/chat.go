package game

import (
	"time"

	"github.com/Tnze/go-mc/registry"

	"github.com/Tnze/go-mc/chat/sign"

	"github.com/Tnze/go-mc/chat"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/server"
	"github.com/go-mc/server/client"
	"go.uber.org/zap"
)

type globalChat struct {
	log           *zap.Logger
	players       *playerList
	chatTypeCodec *registry.Registry[registry.ChatType]
}

func (g *globalChat) broadcastSystemChat(msg chat.Message, overlay bool) {
	g.log.Info(msg.String(), zap.Bool("type", overlay))
	g.players.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		c.(*client.Client).SendSystemChat(msg, overlay)
	})
}

func (g *globalChat) Handle(p pk.Packet, c *client.Client) error {
	var (
		message         pk.String
		timestamp       pk.Long
		salt            pk.Long
		signature       pk.ByteArray
		signedPreview   pk.Boolean
		prevMsg         []sign.HistoryMessage
		lastReceived    pk.Boolean
		lastReceivedMsg sign.HistoryMessage
	)
	err := p.Scan(
		&message,
		&timestamp,
		&salt,
		&signature,
		&signedPreview,
		pk.Array(&prevMsg),
		&lastReceived, pk.Opt{
			Has:   &lastReceived,
			Field: &lastReceivedMsg,
		},
	)
	if err != nil {
		return err
	}

	player := c.GetPlayer()
	g.log.Info(
		string(message),
		zap.String("sender", player.Name),
		zap.Time("timestamp", time.UnixMilli(int64(timestamp))),
	)
	unsignedMsg := chat.Text(string(message))
	chatTypeID, d := g.chatTypeCodec.Find("minecraft:chat")
	chatType := chat.Type{
		ID:         chatTypeID,
		SenderName: chat.Text(player.Name),
		TargetName: nil,
	}
	g.players.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		c.(*client.Client).SendSystemChat(chatType.Decorate(unsignedMsg, &d.Chat), false)
	})
	return nil
}
