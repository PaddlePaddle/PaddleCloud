package pfsmodules

import (
	"crypto/md5"
	"io"
	"os"
)

type ChunkMeta struct {
	fileOffset int64
	checksum   []byte
	len        uint32
}

type Chunk struct {
	Meta ChunkMeta
	Data []byte
}

type GetChunksMetaReq struct {
	Path string `json:"Path"`
}

type GetChunksMetaRep struct {
	Err   string      `json:"Err"`
	Metas []ChunkMeta `json:"Metas"`
}

type PostChunksRep struct {
	Err string `json:"Err"`
}

const (
	defaultMaxChunkSize = 2 * 1024 * 1024
)

func GetChunksMeta(path string, len uint32) ([]ChunkMeta, error) {
	f, err := os.Open(path) // For read access.
	if err != nil {
		return nil, err
	}

	defer f.Close()

	if len > defaultMaxChunkSize || len <= 1024 {
		len = defaultMaxChunkSize
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	metas := make([]ChunkMeta, 0, fi.Size()/int64(len)+1)
	data := make([]byte, len)
	offset := int64(0)

	for {
		n, err := f.Read(data)
		if err != nil && err != io.EOF {
			return metas, err
		}

		if err == io.EOF {
			break
		}

		m := ChunkMeta{}
		m.fileOffset = offset
		sum := md5.Sum(data)
		m.checksum = sum[:]
		m.len = uint32(n)

		metas = append(metas, m)

		offset += int64(n)
	}

	return metas, nil
}
