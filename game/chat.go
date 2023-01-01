// This file is part of go-mc/server project.
// Copyright (C) 2023.  Tnze
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

	"go.uber.org/zap"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/chat/sign"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/registry"
	"github.com/Tnze/go-mc/server"
	"github.com/go-mc/server/client"
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
			&sign.PackedMessageBody{
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
