package game

import (
	"time"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/chat/sign"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/registry"
	"github.com/Tnze/go-mc/server"
	"github.com/go-mc/server/client"
	"go.uber.org/zap"
)

const MsgExpiresTime = time.Minute * 5

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
		timestampLong   pk.Long
		salt            pk.Long
		signature       pk.ByteArray
		signedPreview   pk.Boolean
		prevMsg         []sign.HistoryMessage
		lastReceived    pk.Boolean
		lastReceivedMsg sign.HistoryMessage
	)
	err := p.Scan(
		&message,
		&timestampLong,
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
	timestamp := time.UnixMilli(int64(timestampLong))
	logger := g.log.With(
		zap.String("sender", player.Name),
		zap.Time("timestamp", timestamp),
	)
	if time.Since(timestamp) > MsgExpiresTime {
		logger.Debug("Player send expired message",
			zap.String("msg", string(message)),
		)
		return nil
	}
	// auth.VerifySignature(player.PubKey.PubKey)
	logger.Info(string(message))
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
