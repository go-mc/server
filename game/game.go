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
	"compress/gzip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/save"
	"github.com/Tnze/go-mc/server"
	"github.com/Tnze/go-mc/yggdrasil/user"
	"github.com/go-mc/server/client"
	"github.com/go-mc/server/world"
)

type Game struct {
	log *zap.Logger

	config     Config
	serverInfo *server.PingInfo

	playerProvider world.PlayerProvider
	overworld      *world.World

	globalChat globalChat
	*playerList
}

func NewGame(log *zap.Logger, config Config, pingList *server.PlayerList, serverInfo *server.PingInfo) *Game {
	// providers
	overworld, err := createWorld(log, filepath.Join(".", config.LevelName), &config)
	if err != nil {
		log.Fatal("cannot load overworld", zap.Error(err))
	}
	playerProvider := world.NewPlayerProvider(filepath.Join(".", config.LevelName, "playerdata"))

	// keepalive
	keepAlive := server.NewKeepAlive()
	pl := playerList{pingList: pingList, keepAlive: keepAlive}
	keepAlive.AddPlayerDelayUpdateHandler(func(c server.KeepAliveClient, latency time.Duration) {
		pl.updateLatency(c.(*client.Client), latency)
	})
	go keepAlive.Run(context.TODO())

	return &Game{
		log: log.Named("game"),

		config:     config,
		serverInfo: serverInfo,

		playerProvider: playerProvider,
		overworld:      overworld,

		globalChat: globalChat{
			log:           log.Named("chat"),
			players:       &pl,
			chatTypeCodec: &world.NetworkCodec.ChatType,
		},
		playerList: &pl,
	}
}

func createWorld(logger *zap.Logger, path string, config *Config) (*world.World, error) {
	f, err := os.Open(filepath.Join(path, "level.dat"))
	if err != nil {
		return nil, err
	}
	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	lv, err := save.ReadLevel(r)
	if err != nil {
		return nil, err
	}
	overworld := world.New(
		logger.Named("overworld"),
		world.NewProvider(filepath.Join(path, "region"), config.ChunkLoadingLimiter.Limiter()),
		world.Config{
			ViewDistance:  config.ViewDistance,
			SpawnAngle:    lv.Data.SpawnAngle,
			SpawnPosition: [3]int32{lv.Data.SpawnX, lv.Data.SpawnY, lv.Data.SpawnZ},
		},
	)
	return overworld, nil
}

// AcceptPlayer will be called in an independent goroutine when new player login
func (g *Game) AcceptPlayer(name string, id uuid.UUID, profilePubKey *user.PublicKey, properties []user.Property, protocol int32, conn *net.Conn) {
	logger := g.log.With(
		zap.String("name", name),
		zap.String("uuid", id.String()),
		zap.Int32("protocol", protocol),
	)

	p, err := g.playerProvider.GetPlayer(name, id, profilePubKey, properties)
	if errors.Is(err, os.ErrNotExist) {
		p = &world.Player{
			Entity: world.Entity{
				EntityID: world.NewEntityID(),
				Position: [3]float64{48, 100, 35},
				Rotation: [2]float32{},
			},
			Name:           name,
			UUID:           id,
			PubKey:         profilePubKey,
			Properties:     properties,
			Gamemode:       1,
			ChunkPos:       [3]int32{48 >> 4, 64 >> 4, 35 >> 4},
			EntitiesInView: make(map[int32]*world.Entity),
			ViewDistance:   10,
		}
	} else if err != nil {
		logger.Error("Read player data error", zap.Error(err))
		return
	}
	c := client.New(logger, conn, p)

	logger.Info("Player join", zap.Int32("eid", p.EntityID))
	defer logger.Info("Player left")

	c.SendLogin(g.overworld, p)
	c.SendServerData(g.serverInfo.Description(), g.serverInfo.FavIcon(), g.config.EnforceSecureProfile)

	joinMsg := chat.TranslateMsg("multiplayer.player.joined", chat.Text(p.Name)).SetColor(chat.Yellow)
	leftMsg := chat.TranslateMsg("multiplayer.player.left", chat.Text(p.Name)).SetColor(chat.Yellow)
	g.globalChat.broadcastSystemChat(joinMsg, false)
	defer g.globalChat.broadcastSystemChat(leftMsg, false)
	c.AddHandler(packetid.ServerboundChat, &g.globalChat)

	g.playerList.addPlayer(c, p)
	defer g.playerList.removePlayer(c)

	c.SendPlayerPosition(p.Position, p.Rotation)
	g.overworld.AddPlayer(c, p, g.config.PlayerChunkLoadingLimiter.Limiter())
	defer g.overworld.RemovePlayer(c, p)
	c.SendPacket(packetid.ClientboundUpdateTags, pk.Array(defaultTags))
	c.SendSetDefaultSpawnPosition(g.overworld.SpawnPositionAndAngle())

	c.Start()
}
