package pfsmod

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	log "github.com/golang/glog"
)

// ChunkCmd respresents
type ChunkCmd struct {
	Path      string
	Offset    int64
	ChunkSize int64
}

// NewChunkCmd get a new ChunkCmd
func NewChunkCmd(path string, offset, chunkSize int64) *ChunkCmd {
	return &ChunkCmd{
		Path:      path,
		Offset:    offset,
		ChunkSize: chunkSize,
	}
}

// ToURLParam encodes variables to url encoding parameters
func (p *ChunkCmd) ToURLParam() string {
	parameters := url.Values{}
	parameters.Add("path", p.Path)

	str := fmt.Sprint(p.Offset)
	parameters.Add("offset", str)

	str = fmt.Sprint(p.ChunkSize)
	parameters.Add("chunksize", str)

	return parameters.Encode()
}

// ToJSON encodes ChunkCmd to JSON string
func (p *ChunkCmd) ToJSON() ([]byte, error) {
	panic("not implemented")
}

// Run runs a ChunkCmd
func (p *ChunkCmd) Run() (interface{}, error) {
	panic("not implemented")
}

// NewChunkCmdFromURLParam get a ChunkCmd structure
// path example:
// 	  path=/pfs/datacenter1/1.txt&offset=4096&chunksize=4096
func NewChunkCmdFromURLParam(path string) (*ChunkCmd, error) {
	cmd := ChunkCmd{}

	m, err := url.ParseQuery(path)
	if err != nil ||
		len(m["path"]) == 0 ||
		len(m["offset"]) == 0 ||
		len(m["chunksize"]) == 0 {
		return nil, errors.New(StatusText(StatusJSONErr))
	}

	cmd.Path = m["path"][0]
	cmd.Offset, err = strconv.ParseInt(m["offset"][0], 10, 64)
	if err != nil {
		return nil, errors.New(StatusText(StatusJSONErr))
	}

	chunkSize, err := strconv.ParseInt(m["chunksize"][0], 10, 64)
	if err != nil {
		return nil, errors.New(StatusText(StatusBadChunkSize))
	}
	cmd.ChunkSize = chunkSize

	return &cmd, nil
}

// LoadChunkData loads a specified chunk to w
func (p *ChunkCmd) LoadChunkData(w io.Writer) error {
	f, err := os.Open(p.Path)
	if err != nil {
		return err
	}
	defer Close(f)

	_, err = f.Seek(p.Offset, 0)
	if err != nil {
		return err
	}

	writen, err := io.CopyN(w, f, p.ChunkSize)
	log.V(2).Infof("writen:%d\n", writen)
	return err
}

// SaveChunkData save data from r
func (p *ChunkCmd) SaveChunkData(r io.Reader) error {
	f, err := os.OpenFile(p.Path, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer Close(f)

	_, err = f.Seek(p.Offset, 0)
	if err != nil {
		return err
	}

	writen, err := io.CopyN(f, r, p.ChunkSize)
	log.V(2).Infof("chunksize:%d writen:%d\n", p.ChunkSize, writen)
	return err
}
