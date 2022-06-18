package game

import (
	"github.com/Tnze/go-mc/net"
	"github.com/go-mc/server/client"
	"github.com/go-mc/server/player"
	"github.com/go-mc/server/world"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"path/filepath"
)

type Game struct {
	log *zap.Logger

	config Config

	overworld *world.World
}

func NewGame(log *zap.Logger, config Config) *Game {
	overworld := world.NewProvider(filepath.Join(".", config.LevelName, "region"))

	return &Game{
		log:       log,
		config:    config,
		overworld: world.New(log.Named("overworld"), overworld),
	}
}

// AcceptPlayer 在新玩家登入时在单独的goroutine中被调用
func (g *Game) AcceptPlayer(name string, id uuid.UUID, protocol int32, conn *net.Conn) {
	c := client.New(g.log, conn)
	p := player.New(name, id)
	c.JoinWorld(p, g.overworld)
}
