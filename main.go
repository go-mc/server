package main

import (
	"flag"
	"runtime/debug"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/server"
	"github.com/go-mc/server/game"
	"go.uber.org/zap"
)

var isDebug = flag.Bool("debug", false, "Enable debug log output")

func main() {
	flag.Parse()
	// initialize log library
	var logger *zap.Logger
	if *isDebug {
		logger = unwrap(zap.NewDevelopment())
	} else {
		logger = unwrap(zap.NewProduction())
	}
	defer logger.Sync()

	logger.Info("Program start")
	printBuildInfo(logger)
	defer logger.Info("Program exit")

	// load server config
	config, err := readConfig()
	if err != nil {
		logger.Error("Read config fail", zap.Error(err))
		return
	}

	// initialize player list and server status module, the two modules work together to show server Ping&List information
	playerList := server.NewPlayerList(config.MaxPlayers)
	serverInfo := server.NewPingInfo(
		"Go-MC "+server.ProtocolName,
		server.ProtocolVersion,
		chat.Text(config.MessageOfTheDay),
		nil,
	)
	if err != nil {
		logger.Error("Init server info system fail", zap.Error(err))
		return
	}

	s := server.Server{
		Logger: zap.NewStdLog(logger),
		ListPingHandler: struct {
			*server.PlayerList
			*server.PingInfo
		}{playerList, serverInfo},
		LoginHandler: &server.MojangLoginHandler{
			OnlineMode:           config.OnlineMode,
			EnforceSecureProfile: config.EnforceSecureProfile,
			Threshold:            config.NetworkCompressionThreshold,
			LoginChecker:         playerList, // playerList implement LoginChecker interface to limit the maximum number of online players
		},
		GamePlay: game.NewGame(logger, config, playerList, serverInfo),
	}
	logger.Info("Start listening", zap.String("address", config.ListenAddress))
	err = s.Listen(config.ListenAddress)
	if err != nil {
		logger.Error("Server listening error", zap.Error(err))
	}
}

// printBuildInfo reading compile information of the binary program with runtime/debug package，and print it to log
func printBuildInfo(logger *zap.Logger) {
	binaryInfo, _ := debug.ReadBuildInfo()
	settings := make(map[string]string)
	for _, v := range binaryInfo.Settings {
		settings[v.Key] = v.Value
	}
	logger.Debug("Build info", zap.Any("settings", settings))
}

// readConfig read server config from config file. Throw error when meet unknown setting
func readConfig() (game.Config, error) {
	var c game.Config
	meta, err := toml.DecodeFile("config.toml", &c)
	if err != nil {
		return game.Config{}, err
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		var err errUnknownConfig
		for _, key := range undecoded {
			err = append(err, key.String())
		}
		return game.Config{}, err
	}

	return c, nil
}

type errUnknownConfig []string

func (e errUnknownConfig) Error() string {
	return "unknown config keys: [" + strings.Join(e, ", ") + "]"
}

func unwrap[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
