package game

type Config struct {
	MaxPlayers                  int    `toml:"max-players"`
	ListenAddress               string `toml:"listen-address"`
	MessageOfTheDay             string `toml:"motd"`
	NetworkCompressionThreshold int    `toml:"network-compression-threshold"`
	OnlineMode                  bool   `toml:"online-mode"`
	LevelName                   string `toml:"level-name"`
}
