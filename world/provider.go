package world

import (
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/save"
	"github.com/Tnze/go-mc/save/region"
	"io/fs"
	"path/filepath"
)

// Provider 提供区块的存储功能
type Provider struct {
	dir string
}

func NewProvider(dir string) Provider {
	return Provider{dir: dir}
}

func (p *Provider) GetChunk(pos [2]int32) (c *level.Chunk, errRet error) {
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

func (p *Provider) getRegion(rx, rz int) (*region.Region, error) {
	filename := fmt.Sprintf("r.%d.%d.mca", rx, rz)
	path := filepath.Join(p.dir, filename)
	r, err := region.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		r, err = region.Create(path)
	}
	return r, err
}

func (p *Provider) PutChunk(pos [2]int32, c *level.Chunk) (err error) {
	var chunk save.Chunk
	err = level.ChunkToSave(c, &chunk)
	if err != nil {
		return fmt.Errorf("encode chunk data fail: %w", err)
	}

	data, err := chunk.Data(1)
	if err != nil {
		return fmt.Errorf("record chunk data fail: %w", err)
	}

	r, err := p.getRegion(region.At(int(pos[0]), int(pos[1])))
	if err != nil {
		return fmt.Errorf("open region fail: %w", err)
	}
	defer func(r *region.Region) {
		err2 := r.Close()
		if err == nil && err2 != nil {
			err = fmt.Errorf("open region fail: %w", err)
		}
	}(r)

	x, z := region.In(int(pos[0]), int(pos[1]))
	err = r.WriteSector(x, z, data)
	if err != nil {
		return fmt.Errorf("write sector fail: %w", err)
	}

	return nil
}

var errChunkNotExist = errors.New("ErrChunkNotExist")
