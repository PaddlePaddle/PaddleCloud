package pfsmod

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type ChunkCmd struct {
	Path      string
	Offset    int64
	ChunkSize int64
	Data      []byte
}

func NewChunkCmd(path string, offset, chunkSize int64) *ChunkCmd {
	return &ChunkCmd{
		Path:      path,
		Offset:    offset,
		ChunkSize: chunkSize,
	}
}

func (p *ChunkCmd) ToUrlParam() string {
	parameters := url.Values{}
	parameters.Add("path", p.Path)

	str := fmt.Sprint(p.Offset)
	parameters.Add("offset", str)

	str = fmt.Sprint(p.ChunkSize)
	parameters.Add("chunksize", str)

	return parameters.Encode()
}

func (p *ChunkCmd) ToJson() ([]byte, error) {
	return nil, nil
}

func (p *ChunkCmd) Run() (interface{}, error) {
	return nil, nil
}

// path example:
// 	  path=/pfs/datacenter1/1.txt&offset=4096&chunksize=4096
func NewChunkCmdFromUrlParam(path string) (*ChunkCmd, int) {
	cmd := ChunkCmd{}

	m, err := url.ParseQuery(path)
	if err != nil ||
		len(m["path"]) == 0 ||
		len(m["offset"]) == 0 ||
		len(m["chunksize"]) == 0 {
		return nil, http.StatusBadRequest
	}

	//var err error
	cmd.Path = m["path"][0]
	cmd.Offset, err = strconv.ParseInt(m["offset"][0], 10, 64)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	chunkSize, err := strconv.ParseInt(m["chunksize"][0], 10, 64)
	if err != nil {
		return nil, http.StatusBadRequest
	}
	cmd.ChunkSize = chunkSize

	return &cmd, http.StatusOK
}

func (p *ChunkCmd) WriteChunkData(w io.Writer) error {
	f, err := os.Open(p.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Seek(p.Offset, 0)
	if err != nil {
		return err
	}

	writer := multipart.NewWriter(w)
	defer writer.Close()

	writer.SetBoundary(DefaultMultiPartBoundary)

	fName := p.ToUrlParam()
	part, err := writer.CreateFormFile("chunk", fName)
	if err != nil {
		return err
	}

	_, err = io.CopyN(part, f, p.ChunkSize)
	if err != nil {
		return err
	}
	return nil
}

func (p *ChunkCmd) GetChunkData(r io.Reader) error {
	f, err := os.Open(p.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Seek(p.Offset, 0)
	if err != nil {
		return err
	}

	_, err = io.CopyN(f, r, p.ChunkSize)
	if err != nil {
		return err
	}
	return nil
}
