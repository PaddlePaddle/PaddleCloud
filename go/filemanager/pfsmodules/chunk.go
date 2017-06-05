package pfsmodules

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	log "github.com/golang/glog"
)

// Chunk respresents a chunk info.
type Chunk struct {
	Path   string
	Offset int64
	Size   int64
}

// ToURLParam encodes variables to url encoding parameters.
func (p *Chunk) ToURLParam() url.Values {
	parameters := url.Values{}
	parameters.Add("path", p.Path)

	str := fmt.Sprint(p.Offset)
	parameters.Add("offset", str)

	str = fmt.Sprint(p.Size)
	parameters.Add("chunksize", str)

	return parameters
}

// ParseChunk get a Chunk struct from path.
// path example:
// 	  path=/pfs/datacenter1/1.txt&offset=4096&chunksize=4096
func ParseChunk(path string) (*Chunk, error) {
	cmd := Chunk{}

	m, err := url.ParseQuery(path)
	if err != nil ||
		len(m["path"]) == 0 ||
		len(m["offset"]) == 0 ||
		len(m["chunksize"]) == 0 {
		return nil, errors.New(StatusJSONErr)
	}

	cmd.Path = m["path"][0]
	cmd.Offset, err = strconv.ParseInt(m["offset"][0], 10, 64)
	if err != nil {
		return nil, errors.New(StatusJSONErr)
	}

	chunkSize, err := strconv.ParseInt(m["chunksize"][0], 10, 64)
	if err != nil {
		return nil, errors.New(StatusBadChunkSize)
	}
	cmd.Size = chunkSize

	return &cmd, nil
}

// LoadChunkData loads a specified chunk to io.Writer.
func (p *Chunk) LoadChunkData(w io.Writer) error {
	f, err := os.Open(p.Path)
	if err != nil {
		return err
	}
	defer Close(f)

	_, err = f.Seek(p.Offset, 0)
	if err != nil {
		return err
	}

	loaded, err := io.CopyN(w, f, p.Size)
	log.V(2).Infof("loaded:%d\n", loaded)
	return err
}

// SaveChunkData save data from io.Reader.
func (p *Chunk) SaveChunkData(r io.Reader) error {
	f, err := os.OpenFile(p.Path, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer Close(f)

	_, err = f.Seek(p.Offset, 0)
	if err != nil {
		return err
	}

	writen, err := io.CopyN(f, r, p.Size)
	log.V(2).Infof("chunksize:%d writen:%d\n", p.Size, writen)
	return err
}
