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
	"github.com/Tnze/go-mc/save"
	"github.com/Tnze/go-mc/server"
	"github.com/Tnze/go-mc/server/auth"
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
	overworld, err := createWorld(log, filepath.Join(".", config.LevelName), &config)
	if err != nil {
		log.Fatal("cannot load overworld", zap.Error(err))
	}
	keepAlive := server.NewKeepAlive()
	pl := playerList{pingList: pingList, keepAlive: keepAlive}
	keepAlive.AddPlayerDelayUpdateHandler(func(c server.KeepAliveClient, latency time.Duration) {
		pl.updateLatency(c.(*client.Client), latency)
	})
	go keepAlive.Run(context.TODO())
	playerProvider := world.NewPlayerProvider(filepath.Join(".", config.LevelName, "playerdata"))

	registryCodec := world.NetworkCodec
	return &Game{
		log: log.Named("game"),

		config:     config,
		serverInfo: serverInfo,

		playerProvider: playerProvider,
		overworld:      overworld,

		globalChat: globalChat{
			log:           log.Named("chat"),
			players:       &pl,
			chatTypeCodec: &registryCodec.ChatType,
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
		[3]int32{lv.Data.SpawnX, lv.Data.SpawnY, lv.Data.SpawnZ},
		lv.Data.SpawnAngle,
	)
	return overworld, nil
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

	joinMsg := chat.TranslateMsg("multiplayer.player.joined", chat.Text(p.Name)).SetColor(chat.Yellow)
	leftMsg := chat.TranslateMsg("multiplayer.player.left", chat.Text(p.Name)).SetColor(chat.Yellow)
	g.globalChat.broadcastSystemChat(joinMsg, false)
	defer g.globalChat.broadcastSystemChat(leftMsg, false)

	g.playerList.addPlayer(c, p)
	defer g.playerList.removePlayer(c)

	c.AddHandler(packetid.ServerboundChat, &g.globalChat)

	c.SendLogin(g.overworld, p)
	c.SendServerData(g.serverInfo.Description(), g.serverInfo.FavIcon(), g.config.EnforceSecureProfile)
	c.SendSetDefaultSpawnPosition(g.overworld.SpawnPositionAndAngle())
	c.SendPlayerPosition(p.Position, p.Rotation, true)
	g.overworld.AddPlayer(c, p, g.config.PlayerChunkLoadingLimiter.Limiter())
	defer g.overworld.RemovePlayer(c, p)

	c.Start()
}
