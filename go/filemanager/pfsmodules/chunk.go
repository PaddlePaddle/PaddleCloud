package pfsmodules

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// ChunkParam respresents a chunk info.
type ChunkParam struct {
	Path   string
	Offset int64
	Size   int64
}

// String packs info of ChunkParam.
func (p *ChunkParam) String() string {
	return fmt.Sprintf("Path:%s Offset:%d Size:%d", p.Path, p.Offset, p.Size)
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

// ParseChunkParam get a Chunk struct from path.
// path example:
// 	  path=/pfs/datacenter1/1.txt&offset=4096&chunksize=4096
func ParseChunkParam(path string) (*ChunkParam, error) {
	cmd := ChunkParam{}

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

// Chunk is struct.
type Chunk struct {
	Offset   int64
	Len      int64
	Checksum string
	Data     []byte
}

// String packs info of Chunk
func (c *Chunk) String() string {
	return fmt.Sprintf("Offset:%d Len:%d Checksum:%s", c.Offset, c.Len, c.Checksum)
}

// NewChunk make a Chunk struct.
func NewChunk(capcity int64) *Chunk {
	c := Chunk{}
	c.Data = make([]byte, capcity)
	return &c
}
