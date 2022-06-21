package player

import (
	"compress/gzip"
	"fmt"
	"github.com/Tnze/go-mc/save"
	"github.com/google/uuid"
	"os"
	"path/filepath"
)

type Provider struct {
	dir string
}

func NewProvider(dir string) Provider {
	return Provider{dir: dir}
}

func (p *Provider) GetPlayer(name string, id uuid.UUID) (player *Player, errRet error) {
	f, err := os.Open(filepath.Join(p.dir, id.String()+".dat"))
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err2 := f.Close()
		if errRet == nil && err2 != nil {
			errRet = fmt.Errorf("close player data fail: %w", err2)
		}
	}(f)
	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("open gzip reader fail: %w", err)
	}
	data, err := save.ReadPlayerData(r)
	if err != nil {
		return nil, fmt.Errorf("read player data fail: %w", err)
	}
	if err := r.Close(); err != nil {
		return nil, fmt.Errorf("close gzip reader fail: %w", err)
	}
	player = New(name, id)
	player.pos = data.Pos
	player.gamemode = data.PlayerGameType
	player.viewDistance = 20
	return
}
