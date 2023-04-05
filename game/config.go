// This file is part of go-mc/server project.
// Copyright (C) 2023.  Tnze
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package game

import (
	"time"

	"golang.org/x/time/rate"
)

type Config struct {
	MaxPlayers                  int    `toml:"max-players"`
	ViewDistance                int32  `toml:"view-distance"`
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
