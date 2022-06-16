package main

import (
	"flag"

	"go.uber.org/zap"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/server"
	"github.com/go-mc/server/game"
)

var motd = chat.Message{Text: "A Minecraft Server ", Extra: []chat.Message{{Text: "Powered by go-mc", Color: "yellow"}}}
var addr = flag.String("Address", "127.0.0.1:25565", "Listening address")

func main() {
	flag.Parse()

	// 初始化日志库
	logger := unwrap(zap.NewDevelopment())
	defer logger.Sync()

	logger.Info("Program start")
	defer logger.Info("Program exit")

	// TODO：从服务器配置文件中提取最大人数、是否开启在线验证等设置项目，而不是在代码中写死

	// 初始化玩家列表和服务器状态信息模块，这两个模块相辅相成，用于服务器Ping&List信息的显示
	playerList := server.NewPlayerList(20)
	serverInfo, err := server.NewPingInfo(playerList, "GoMC "+server.ProtocolName, server.ProtocolVersion, motd, nil)
	if err != nil {
		logger.Panic("Init server info system fail", zap.Error(err))
	}

	s := server.Server{
		ListPingHandler: serverInfo,
		LoginHandler: &server.MojangLoginHandler{
			OnlineMode:   true,
			Threshold:    256,
			LoginChecker: playerList,
		},
		GamePlay: game.NewGame(logger),
	}
	logger.Info("Start listening", zap.String("address", *addr))
	err = s.Listen(*addr)
	if err != nil {
		logger.Error("Server listening error", zap.Error(err))
	}
}

func unwrap[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
