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

type ChunkMetaCmdAttr struct {
	Path      string
	BlockSize uint32
}

type ChunkMetaCmdResponse struct {
	Err   string      `json:"err"`
	Path  string      `json:"path"`
	Metas []ChunkMeta `json:"metas"`
}

type ChunkMetaCmd struct {
	cmdAttr *CmdAttr
	resp    *ChunkMetaCmdResponse
}

func GetChunksMeta(path string, len uint32) ([]ChunkMeta, error) {
	f, err := os.Open(path) // For read access.
	if err != nil {
		return nil, err
	}

	defer f.Close()

	if len > defaultMaxChunkSize || len < defaultMinChunkSize {
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
		sum := md5.Sum(data[:n])
		m.checksum = sum[:]
		m.len = uint32(n)

		metas = append(metas, m)

		offset += int64(n)
	}

	return metas, nil
}
