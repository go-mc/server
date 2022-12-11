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
	g.log.Info(msg.String(), zap.Bool("overlay", overlay))
	g.players.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		c.(*client.Client).SendSystemChat(msg, overlay)
	})
}

func (g *globalChat) Handle(p pk.Packet, c *client.Client) error {
	var (
		message       pk.String
		timestampLong pk.Long
		salt          pk.Long
		signature     pk.Option[sign.Signature, *sign.Signature]
		lastSeen      sign.HistoryUpdate
	)
	err := p.Scan(
		&message,
		&timestampLong,
		&salt,
		&signature,
		&lastSeen,
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

	if existInvalidCharacter(string(message)) {
		c.SendDisconnect(chat.TranslateMsg("multiplayer.disconnect.illegal_characters"))
		return nil
	}

	if !player.SetLastChatTimestamp(timestamp) {
		c.SendDisconnect(chat.TranslateMsg("multiplayer.disconnect.out_of_order_chat"))
		return nil
	}

	// TODO: check if the client disable chatting
	if false {
		c.SendSystemChat(chat.TranslateMsg("chat.disabled.options").SetColor(chat.Red), false)
		return nil
	}

	// verify message
	//var playerMsg sign.PlayerMessage
	////if player.PubKey != nil {
	////}

	if time.Since(timestamp) > MsgExpiresTime {
		logger.Warn("Player send expired message", zap.String("msg", string(message)))
		return nil
	}
	chatTypeID, decorator := g.chatTypeCodec.Find("minecraft:chat")
	chatType := chat.Type{
		ID:         chatTypeID,
		SenderName: chat.Text(player.Name),
		TargetName: nil,
	}
	decorated := chatType.Decorate(chat.Text(string(message)), &decorator.Chat)
	logger.Info(decorated.String())

	g.players.pingList.Range(func(c server.PlayerListClient, _ server.PlayerSample) {
		c.(*client.Client).SendPlayerChat(
			player.UUID,
			0,
			signature,
			&sign.MessageBody{
				PlainMsg:  string(message),
				Timestamp: timestamp,
				Salt:      int64(salt),
				LastSeen:  []sign.PackedSignature{},
			},
			nil,
			&sign.FilterMask{Type: 0},
			&chatType,
		)
	})
	return nil
}

func existInvalidCharacter(msg string) bool {
	for _, c := range msg {
		if c == '§' || c < ' ' || c == '\x7F' {
			return true
		}
	}
	return false
}
