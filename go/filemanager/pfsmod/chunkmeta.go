package pfsmod

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
)

type ChunkMeta struct {
	Offset   int64  `json:"offset"`
	Checksum string `json:"checksum"`
	Len      int64  `json:"len"`
}

type ChunkMataCmd struct {
	Method    string `json:"method"`
	FilePath  string `json:"path"`
	ChunkSize int64  `json:"chunksize"`
}

func (p *ChunkMetaCmd) ToUrlParam() string {
	parameters := url.Values{}
	parameters.Add("method", p.Method)
	parameters.Add("path", p.Path)

	str := fmt.Sprint(p.ChunkSize)
	parameters.Add("chunksize", str)

	return parameters.Encode()
}

func (p *ChunkMetaCmd) ToJson() ([]byte, error) {
	return json.Marshal(p)
}

func (p *ChunkMetaCmd) Run() (interface{}, error) {
	metas, err := getChunkMeta(p.FilePath, p.ChunkSize)
	if err != nil {
		return nil, err
	}

	return metas, nil
}

func NewChunkMetaCmdFromUrl(r *http.Request) (*ChunkMetaCmd, int32) {
	method := r.URL.Query().Get("method")
	path := r.URL.Query().Get("path")
	chunkStr := r.URL.Query().Get("chunksize")

	if len(method) == 0 ||
		method != "GetChunkMeta" ||
		len(path) < 4 ||
		len(chunkStr) == 0 {
		return nil, http.StatusBadRequest
	}

	inputSize, err := strconv.ParseInt(chunkStr, 10, 64)
	if err != nil {
		return nil, http.StatusBadRequest
	}
	chunkSize = int64(inputSize)

	if chunkSize < defaultMinChunkSize || chunkSize > defaultMaxChunkSize {
		return nil, http.StatusBadRequest
	}

	return &ChunkMetaCmd{
		Method:    method,
		FilePath:  path,
		ChunkSize: chunkSize,
	}, 0
}

func NewChunkMetaCmd(path string, chunkSize int64) *ChunkMetaCmd {
	return &ChunMetaCmd{
		Method:    "GetChunkMeta",
		FilePath:  path,
		ChunkSize: chunkSize,
	}
}

type metaSlice []ChunkMeta

func (a metaSlice) Len() int           { return len(a) }
func (a metaSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a metaSlice) Less(i, j int) bool { return a[i].Offset < a[j].Offset }

func GetDiffChunkMeta(srcMeta []ChunkMeta, destMeta []ChunkMeta) ([]ChunkMeta, error) {
	if destMeta == nil || len(destMeta) == 0 || len(srcMeta) == 0 {
		return srcMeta, nil
	}

	if !sort.IsSorted(metaSlice(srcMeta)) {
		sort.Sort(metaSlice(srcMeta))
	}

	if !sort.IsSorted(metaSlice(destMeta)) {
		sort.Sort(metaSlice(destMeta))
	}

	dstIdx := 0
	srcIdx := 0
	diff := make([]ChunkMeta, 0, len(srcMeta))

	for {
		if srcMeta[srcIdx].Offset < destMeta[dstIdx].Offset {
			diff = append(diff, srcMeta[srcIdx])
			srcIdx += 1
		} else if srcMeta[srcIdx].Offset > destMeta[dstIdx].Offset {
			dstIdx += 1
		} else {
			if srcMeta[srcIdx].Checksum != destMeta[dstIdx].Checksum {
				diff = append(diff, srcMeta[srcIdx])
			}

			dstIdx += 1
			srcIdx += 1
		}

		if dstIdx >= len(destMeta) {
			break
		}

		if srcIdx >= len(srcMeta) {
			break
		}
	}

	if srcIdx < len(srcMeta) {
		diff = append(diff, srcMeta[srcIdx:len(srcMeta)]...)
	}

	return diff, nil
}

func getChunkMeta(path string, len int64) ([]ChunkMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if len > defaultMaxChunkSize || len < defaultMinChunkSize {
		return nil, errors.New(BadChunkSizeArguments)
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

		//log.Println(n)
		m := ChunkMeta{}
		m.Offset = offset
		sum := md5.Sum(data[:n])
		m.Checksum = hex.EncodeToString(sum[:])
		m.Len = int64(n)

		metas = append(metas, m)

		offset += int64(n)
	}

	return metas, nil
}
