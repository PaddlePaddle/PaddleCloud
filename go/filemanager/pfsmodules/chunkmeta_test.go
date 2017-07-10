package pfsmodules

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func TestGetDiffMeta(t *testing.T) {
	var src []ChunkMeta
	var dst []ChunkMeta

	var data []ChunkMeta

	for i := 0; i < 4; i++ {
		data = append(data, ChunkMeta{
			Offset:   int64(i),
			Checksum: fmt.Sprintf("%x", md5.Sum([]byte("1"))),
			Len:      1})
	}

	src = data[0:3]
	// check when len(dst) is 0
	diff, _ := GetDiffChunkMeta(src, dst)
	if len(diff) != 3 {
		t.Error(len(diff))
	}

	for i, _ := range diff {
		if diff[i] != src[i] {
			t.Error(i)
		}
	}

	// check when dst is same as src
	dst = src
	diff, _ = GetDiffChunkMeta(src, dst)
	if len(diff) != 0 {
		t.Error(len(diff))
	}

	// check when dst is small than src
	dst = dst[:0]
	dst = append(dst, data[0])
	dst = append(dst, data[2])

	diff, _ = GetDiffChunkMeta(src, dst)
	if len(diff) != 1 {
		t.Error(len(diff))
	}

	if diff[0] != data[1] {
		t.Error(0)
	}

	// check when dst is large then src
	dst = data
	diff, _ = GetDiffChunkMeta(src, dst)
	if len(diff) != 0 {
		t.Error(len(diff))
	}
}
