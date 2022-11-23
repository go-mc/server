package world

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/save"
	"github.com/Tnze/go-mc/save/region"
	"github.com/Tnze/go-mc/server/auth"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// ChunkProvider implements chunk storage
type ChunkProvider struct {
	dir     string
	limiter *rate.Limiter
}

func NewProvider(dir string, limiter *rate.Limiter) ChunkProvider {
	return ChunkProvider{dir: dir, limiter: limiter}
}

func (p *ChunkProvider) GetChunk(pos [2]int32) (c *level.Chunk, errRet error) {
	if !p.limiter.Allow() {
		return nil, errors.New("reach time limit")
	}
	r, err := p.getRegion(region.At(int(pos[0]), int(pos[1])))
	if err != nil {
		return nil, fmt.Errorf("open region fail: %w", err)
	}
	defer func(r *region.Region) {
		err2 := r.Close()
		if errRet == nil && err2 != nil {
			errRet = fmt.Errorf("close region fail: %w", err2)
		}
	}(r)

	x, z := region.In(int(pos[0]), int(pos[1]))
	if !r.ExistSector(x, z) {
		return nil, errChunkNotExist
	}

	data, err := r.ReadSector(x, z)
	if err != nil {
		return nil, fmt.Errorf("read sector fail: %w", err)
	}

	var chunk save.Chunk
	if err := chunk.Load(data); err != nil {
		return nil, fmt.Errorf("parse chunk data fail: %w", err)
	}

	c, err = level.ChunkFromSave(&chunk)
	if err != nil {
		return nil, fmt.Errorf("load chunk data fail: %w", err)
	}
	return c, nil
}

func (p *ChunkProvider) getRegion(rx, rz int) (*region.Region, error) {
	filename := fmt.Sprintf("r.%d.%d.mca", rx, rz)
	path := filepath.Join(p.dir, filename)
	r, err := region.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		r, err = region.Create(path)
	}
	return r, err
}

func (p *ChunkProvider) PutChunk(pos [2]int32, c *level.Chunk) (err error) {
	//var chunk save.Chunk
	//err = level.ChunkToSave(c, &chunk)
	//if err != nil {
	//	return fmt.Errorf("encode chunk data fail: %w", err)
	//}
	//
	//data, err := chunk.Data(1)
	//if err != nil {
	//	return fmt.Errorf("record chunk data fail: %w", err)
	//}
	//
	//r, err := p.getRegion(region.At(int(pos[0]), int(pos[1])))
	//if err != nil {
	//	return fmt.Errorf("open region fail: %w", err)
	//}
	//defer func(r *region.Region) {
	//	err2 := r.Close()
	//	if err == nil && err2 != nil {
	//		err = fmt.Errorf("open region fail: %w", err)
	//	}
	//}(r)
	//
	//x, z := region.In(int(pos[0]), int(pos[1]))
	//err = r.WriteSector(x, z, data)
	//if err != nil {
	//	return fmt.Errorf("write sector fail: %w", err)
	//}

	return nil
}

var errChunkNotExist = errors.New("ErrChunkNotExist")

type PlayerProvider struct {
	dir string
}

func NewPlayerProvider(dir string) PlayerProvider {
	return PlayerProvider{dir: dir}
}

func (p *PlayerProvider) GetPlayer(name string, id uuid.UUID, pubKey *auth.PublicKey, properties []auth.Property) (player *Player, errRet error) {
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
	player = &Player{
		Entity: Entity{
			EntityID: NewEntityID(),
			Position: data.Pos,
			Rotation: data.Rotation,
		},
		Name:           name,
		UUID:           id,
		PubKey:         pubKey,
		Properties:     properties,
		ChunkPos:       [2]int32{int32(data.Pos[0]) >> 5, int32(data.Pos[1]) >> 5},
		Gamemode:       data.PlayerGameType,
		EntitiesInView: make(map[int32]*Entity),
		ViewDistance:   10,
	}
	return
}
