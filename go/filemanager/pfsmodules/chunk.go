package pfsmodules

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"

	log "github.com/golang/glog"
)

// Chunk respresents a chunk info.
type ChunkParam struct {
	Path   string
	Offset int64
	Size   int64
}

// ToURLParam encodes variables to url encoding parameters.
func (p *ChunkParam) ToURLParam() url.Values {
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
func ParseChunkParam(path string) (*ChunkParam, error) {
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

type Chunk struct {
	Offset int64
	Len    int64
	Sum    string
	Data   []byte
}

func NewChunk() {
	return &ChunkData{
		Offset: -1,
		Len:    -1,
	}
}

type FileHandle struct {
	F      *File
	Offset int64
}

func NewFileHandle() {
	return &FileHandle{
		Offset: -1,
	}
}

// LoadChunkData loads a specified chunk to io.Writer.
func (f *FileHandle) Load(offset, len int64) (*Chunk, error) {
	if offset != f.Offset {
		_, err = f.Seek(offset, 0)
		if err != nil {
			return nil, err
		}
		f.Offset = offset
	}

	c = NewChunk()

	n, err := io.CopyN(c.Data, f.F, len)
	log.V(2).Infof("expect %d read %d\n", len, n)

	if err != nil {
		return nil, err
	}
	f.Offset += n

	c.Offset = offset
	c.Len = len
	sum := md5.Sum(c.Data[:n])
	c.Sum = hex.EncodeToString(sum[:])

	return c, nil
}

// Save save data to file
func (f *FileHandle) SaveChunk(c *Chunk) error {
	return f.Save(bytes.NewReader(c.Data), c.Offset, c.Len)
}

// SaveChunkData save data from io.Reader.
func (f *FileHanle) Save(r io.Reader, offset, len int64) error {
	if offset != f.Offset {
		_, err = f.Seek(offset, 0)
		if err != nil {
			return nil, err
		}
		f.Offset = offset
	}

	n, err := io.CopyN(f.F, c.Data, len)
	log.V(2).Infof("expect write %d writen:%d\n", len, n)
	if err == nil {
		f.Offset += n
	}

	return err
}
