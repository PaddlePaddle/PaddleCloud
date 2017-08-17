package pfsmodules

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PaddlePaddle/cloud/go/utils/restclient"
)

const (
	defaultMaxChunkSize = 4 * 1024 * 1024
	defaultMinChunkSize = 4 * 1024
)
const (
	// ChunkMetaCmdName is the name of GetChunkMeta command.
	ChunkMetaCmdName = "GetChunkMeta"
)

// ChunkMeta holds the chunk meta's info.
type ChunkMeta struct {
	Offset   int64  `json:"offset"`
	Checksum string `json:"checksum"`
	Len      int64  `json:"len"`
}

// ToString  pack a info tring of ChunkMeta.
func (m *ChunkMeta) String() string {
	return fmt.Sprintf("Offset:%d Checksum:%s Len:%d", m.Offset, m.Checksum, m.Len)
}

// ChunkMetaCmd is a command.
type ChunkMetaCmd struct {
	Method    string `json:"method"`
	FilePath  string `json:"path"`
	Offset    int64  `json:"offset"`
	ChunkSize int64  `json:"chunksize"`
}

// ToURLParam encodes ChunkMetaCmd to URL encoding string.
func (p *ChunkMetaCmd) ToURLParam() url.Values {
	parameters := url.Values{}
	parameters.Add("method", p.Method)
	parameters.Add("path", p.FilePath)

	str := fmt.Sprint(p.ChunkSize)
	parameters.Add("chunksize", str)

	str = fmt.Sprint(p.Offset)
	parameters.Add("offset", str)

	return parameters
}

// ToJSON encodes ChunkMetaCmd to JSON string.
func (p *ChunkMetaCmd) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// TestSeek tests seek.
func TestSeek(pos int64, len int64) {
	f, _ := os.OpenFile("/pfs/dlnel/home/gongwb/1.dat", os.O_RDONLY, 0666)
	f.Seek(pos, 0)
	//data := make([]byte, len)
	c := NewChunk(len)
	n, _ := f.Read(c.Data)
	fmt.Println(n)
	f.Close()
}

// Run is a functions which run ChunkMetaCmd.
func (p *ChunkMetaCmd) Run() (interface{}, error) {
	f := FileHandle{}
	if err := f.Open(p.FilePath, os.O_RDONLY, 0); err != nil {
		return nil, err
	}

	defer f.Close()

	c, err := f.ReadChunk(p.Offset, p.ChunkSize)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &ChunkMeta{
		Offset:   p.Offset,
		Checksum: c.Checksum,
		Len:      c.Len,
	}, err
}

func (p *ChunkMetaCmd) checkChunkSize() error {
	if p.ChunkSize < defaultMinChunkSize ||
		p.ChunkSize > defaultMaxChunkSize {
		return errors.New(StatusBadChunkSize)
	}

	return nil
}

// ValidateCloudArgs checks the conditions when running on cloud.
func (p *ChunkMetaCmd) ValidateCloudArgs(userName string) error {
	if err := ValidatePfsPath([]string{p.FilePath}, userName, ChunkMetaCmdName); err != nil {
		return err
	}

	return p.checkChunkSize()
}

// ValidateLocalArgs checks the conditions when running locally.
func (p *ChunkMetaCmd) ValidateLocalArgs() error {
	return p.checkChunkSize()
}

// NewChunkMetaCmdFromURLParam get a new ChunkMetaCmd.
func NewChunkMetaCmdFromURLParam(r *http.Request) (*ChunkMetaCmd, error) {
	method := r.URL.Query().Get("method")
	path := r.URL.Query().Get("path")
	chunkStr := r.URL.Query().Get("chunksize")
	offsetStr := r.URL.Query().Get("offset")

	if len(method) == 0 ||
		method != ChunkMetaCmdName ||
		len(path) == 0 ||
		len(chunkStr) == 0 ||
		len(offsetStr) == 0 {
		return nil, errors.New(http.StatusText(http.StatusBadRequest))
	}

	chunkSize, err := strconv.ParseInt(chunkStr, 10, 64)
	if err != nil {
		return nil, errors.New(StatusBadChunkSize)
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		return nil, errors.New(StatusBadChunkSize)
	}

	return &ChunkMetaCmd{
		Method:    method,
		FilePath:  path,
		ChunkSize: chunkSize,
		Offset:    offset,
	}, nil
}
func remoteChunkMeta(path string, offset int64,
	chunkSize int64) (*ChunkMeta, error) {
	cmd := ChunkMetaCmd{
		Method:    ChunkMetaCmdName,
		FilePath:  path,
		ChunkSize: chunkSize,
		Offset:    offset,
	}

	t := fmt.Sprintf("%s/api/v1/pfs/chunks", Config.ActiveConfig.Endpoint)
	ret, err := restclient.GetCall(t, cmd.ToURLParam())
	if err != nil {
		return nil, err
	}

	type chunkMetaResponse struct {
		Err     string    `json:"err"`
		Results ChunkMeta `json:"results"`
	}

	resp := chunkMetaResponse{}
	if err := json.Unmarshal(ret, &resp); err != nil {
		return nil, err
	}

	if len(resp.Err) == 0 {
		return &resp.Results, nil
	}

	if strings.Contains(resp.Err, StatusFileEOF) {
		return &resp.Results, io.EOF
	}

	return &resp.Results, errors.New(resp.Err)
}
