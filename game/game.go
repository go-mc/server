package game

import (
	"github.com/Tnze/go-mc/net"
	"github.com/go-mc/server/client"
	"github.com/go-mc/server/player"
	"github.com/go-mc/server/world"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Game struct {
	log *zap.Logger

	Config
	chunkLoader *world.Loader

	overworld *world.World
}

func NewGame(log *zap.Logger) *Game {
	return &Game{log: log}
}

// AcceptPlayer 在新玩家登入时在单独的goroutine中被调用
func (g *Game) AcceptPlayer(name string, id uuid.UUID, protocol int32, conn *net.Conn) {
	c := client.New(g.log, conn)
	p := player.New(name, id)
	c.JoinWorld(p, g.overworld)
}
