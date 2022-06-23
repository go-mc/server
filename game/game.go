package game

import (
	"context"
	"errors"
	"github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/server"
	"github.com/Tnze/go-mc/server/auth"
	"github.com/go-mc/server/client"
	"github.com/go-mc/server/player"
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

	playerProvider player.Provider
	overworld      *world.World

	*playerList
}

func NewGame(log *zap.Logger, config Config, pingList *server.PlayerList) *Game {
	overworld := world.NewProvider(filepath.Join(".", config.LevelName, "region"))
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

		playerProvider: player.NewProvider(filepath.Join(".", config.LevelName, "playerdata")),
		overworld:      world.New(log.Named("overworld"), overworld),

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
	logger.Info("Player join")
	defer logger.Info("Player left")

	p, err := g.playerProvider.GetPlayer(name, id, profilePubKey, properties)
	if errors.Is(err, os.ErrNotExist) {
		p = player.New(name, id, profilePubKey, properties)
		p.SetGamemode(1)
		p.SetPos([3]float64{48, 64, 35})
		p.SetViewDistance(15)
	} else if err != nil {
		logger.Error("Read player data error", zap.Error(err))
		return
	}
	c := client.New(logger, conn, p)

	g.playerList.addPlayer(c, p)
	defer g.playerList.removePlayer(c)

	if err := c.Spawn(g.overworld); err != nil {
		logger.Error("Spawn player error", zap.Error(err))
		return
	}
	defer g.overworld.RemovePlayer(c)

	c.Start()
}
