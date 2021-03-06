package game

import (
	"golang.org/x/time/rate"
	"time"
)

type Config struct {
	MaxPlayers                  int    `toml:"max-players"`
	ListenAddress               string `toml:"listen-address"`
	MessageOfTheDay             string `toml:"motd"`
	NetworkCompressionThreshold int    `toml:"network-compression-threshold"`
	OnlineMode                  bool   `toml:"online-mode"`
	LevelName                   string `toml:"level-name"`
	EnforceSecureProfile        bool   `toml:"enforce-secure-profile"`

	ChunkLoadingLimiter       Limiter `toml:"chunk-loading-limiter"`
	PlayerChunkLoadingLimiter Limiter `toml:"player-chunk-loading-limiter"`
}

type Limiter struct {
	Every duration `toml:"every"`
	N     int
}

// Limiter convert this to *rate.Limiter
func (l *Limiter) Limiter() *rate.Limiter {
	return rate.NewLimiter(rate.Every(l.Every.Duration), l.N)
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	return
}
