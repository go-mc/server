package main

import (
	"github.com/BurntSushi/toml"
	"github.com/Tnze/go-mc/chat"
	"go.uber.org/zap"
	"runtime/debug"
	"strings"

	"github.com/Tnze/go-mc/server"
	"github.com/go-mc/server/game"
)

func main() {
	// 初始化日志库
	logger := unwrap(zap.NewDevelopment())
	//logger := unwrap(zap.NewProduction())
	defer logger.Sync()

	logger.Info("Program start")
	printBuildInfo(logger)
	defer logger.Info("Program exit")

	// 读取服务器配置文件
	config, err := readConfig()
	if err != nil {
		logger.Error("Read config fail", zap.Error(err))
		return
	}

	// 初始化玩家列表和服务器状态信息模块，这两个模块相辅相成，用于服务器Ping&List信息的显示
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
			LoginChecker:         playerList, // playerList实现了LoginChecker接口，用于限制服务器最大人数
		},
		GamePlay: game.NewGame(logger, config, playerList),
	}
	logger.Info("Start listening", zap.String("address", config.ListenAddress))
	err = s.Listen(config.ListenAddress)
	if err != nil {
		logger.Error("Server listening error", zap.Error(err))
	}
}

// printBuildInfo 通过runtime/debug包读取二进制程序编译信息，并输出至日志
func printBuildInfo(logger *zap.Logger) {
	binaryInfo, _ := debug.ReadBuildInfo()
	settings := make(map[string]string)
	for _, v := range binaryInfo.Settings {
		settings[v.Key] = v.Value
	}
	logger.Debug("Build info", zap.Any("settings", settings))
}

// readConfig 从配置文件中读取服务器设置，当配置文件中出现无法识别的未知设置时报错
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
