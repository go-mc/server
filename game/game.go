package game

import (
	"context"
	"errors"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/server"
	"github.com/Tnze/go-mc/server/auth"
	"github.com/go-mc/server/client"
	"github.com/go-mc/server/world"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

type Game struct {
	log *zap.Logger

	config Config

	playerProvider world.PlayerProvider
	overworld      *world.World

	globalChat globalChat
	*playerList
}

func NewGame(log *zap.Logger, config Config, pingList *server.PlayerList) *Game {
	overworld := world.NewProvider(filepath.Join(".", config.LevelName, "region"), config.ChunkLoadingLimiter.Limiter())
	keepAlive := server.NewKeepAlive()
	pl := playerList{pingList: pingList, keepAlive: keepAlive}
	keepAlive.AddPlayerDelayUpdateHandler(func(c server.KeepAliveClient, latency time.Duration) {
		cc := c.(*client.Client)
		cc.GetPlayer().SetLatency(latency)
		pl.updateLatency(cc, latency)
	})
	go keepAlive.Run(context.TODO())
	return &Game{
		log: log,

		config: config,

		playerProvider: world.NewPlayerProvider(filepath.Join(".", config.LevelName, "playerdata")),
		overworld:      world.New(log.Named("overworld"), overworld),

		globalChat: globalChat{log.Named("Chat"), &pl},
		playerList: &pl,
	}
}

// AcceptPlayer will be called in an independent goroutine when new player login
func (g *Game) AcceptPlayer(name string, id uuid.UUID, profilePubKey *auth.PublicKey, properties []auth.Property, protocol int32, conn *net.Conn) {
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
				Position: [3]float64{48, 64, 35},
				Rotation: [2]float32{},
			},
			Name:           name,
			UUID:           id,
			PubKey:         profilePubKey,
			Properties:     properties,
			Gamemode:       1,
			ChunkPos:       [2]int32{48 >> 5, 35 >> 5},
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

	g.playerList.addPlayer(c, p)
	defer g.playerList.removePlayer(c)

	joinMsg := chat.TranslateMsg("multiplayer.player.joined", chat.Text(p.Name)).SetColor(chat.Yellow)
	leftMsg := chat.TranslateMsg("multiplayer.player.left", chat.Text(p.Name)).SetColor(chat.Yellow)
	g.globalChat.broadcastSystemChat(joinMsg, chat.System)
	defer g.globalChat.broadcastSystemChat(leftMsg, chat.System)

	c.AddHandler(packetid.ServerboundChat, &g.globalChat)

	c.SendLogin(g.overworld, p)
	g.overworld.AddPlayer(c, p, g.config.PlayerChunkLoadingLimiter.Limiter())
	defer g.overworld.RemovePlayer(c, p)

	c.Start()
}
