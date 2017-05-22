package pfsmodules

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

const (
	defaultMaxChunkSize = 2 * 1024 * 1024
	defaultMinChunkSize = 4 * 1024
	DefaultChunkSize    = 1 * 1024 * 1024 * 1024
)

type Chunk struct {
	Meta ChunkMeta
	Data []byte
}

type ChunkCmdAttr struct {
	Path   string `json:"path"`
	Offset int64  `json:"offset"`
	len    uint32 `json:len`
}

type ChunkCmd struct {
	cmdAttr *ChunkCmdAttr
	resp    *Chunk
}

func GetChunk(path string, offset int64, len uint32) (*Chunk, error) {
	f, err := os.Open(path) // For read access.
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if len > defaultMaxChunkSize || len < defaultMinChunkSize {
		return nil, errors.New("invalid len:" + string(len))
	}

	/*
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}
	*/

	/*
		if offset < 0 || offset > fi.Size() {
			return nil, Errors.New("offset > filesize:" + string(fi.Size()))
		}
	*/

	chunk := Chunk{}
	chunk.Data = make([]byte, len)
	m := &chunk.Meta

	if _, err := f.Seek(offset, os.SEEK_SET); err != nil {
		return nil, err
	}

	n, err := f.Read(chunk.Data)
	if err != nil && err != io.EOF {
		return nil, err
	}

	m.Offset = offset
	sum := md5.Sum(chunk.Data[:n])
	m.Checksum = hex.EncodeToString(sum[:])
	m.Len = uint32(n)

	return &chunk, nil
}
