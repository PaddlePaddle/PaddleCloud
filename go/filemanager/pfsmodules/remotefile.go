package pfsmodules

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/PaddlePaddle/cloud/go/utils/restclient"
	log "github.com/golang/glog"
)

// RFileHandle is a remote file's handle.
type RFileHandle struct {
	Path string
	Flag int
	Size int64
}

const (
	// ReadOnly means read only.
	ReadOnly = os.O_RDONLY
	// WriteOnly means write only.
	WriteOnly = os.O_WRONLY
	// ReadAndWrite means read and write.
	ReadAndWrite = os.O_RDWR
)

// Open file to read ,write or read-write.
// if flag == WriteOnly or flag == ReadAndWrite, this function will
// attempt to create a sized file on remote if it does't exist.
func (f *RFileHandle) Open(path string, flag int, size int64) error {
	if flag != ReadOnly &&
		flag != WriteOnly &&
		flag != ReadAndWrite {
		return errors.New("only support ReadOnly, WriteOnly, ReadAndWrite")
	}

	f.Path = path
	f.Flag = flag
	f.Size = size

	if flag == WriteOnly ||
		flag == ReadAndWrite {

		cmd := TouchCmd{
			Method:   TouchCmdName,
			Path:     path,
			FileSize: size,
		}
		// create sized file.
		if err := remoteTouch(&cmd); err != nil {
			return err
		}
	}
	return nil
}

func getChunkData(m ChunkParam) (*Chunk, error) {
	t := fmt.Sprintf("%s/api/v1/pfs/storage/chunks", Config.ActiveConfig.Endpoint)
	log.V(1).Info("target url: " + t)

	resp, err := restclient.GetChunk(t, m.ToURLParam())
	if err != nil {
		return nil, err
	}
	defer Close(resp.Body)

	if resp.Status != restclient.HTTPOK {
		return nil, errors.New("http server returned non-200 status: " + resp.Status)
	}

	var c *Chunk
	partReader := multipart.NewReader(resp.Body, DefaultMultiPartBoundary)
	for {
		part, err := partReader.NextPart()
		if err == io.EOF {
			break
		}

		if part.FormName() == "chunk" {
			m1, err := ParseChunkParam(part.FileName())
			if err != nil {
				return nil, errors.New(err.Error())
			}

			if m1.Size == 0 {
				return c, io.EOF
			}

			c = NewChunk(m1.Size)
			c.Len = m1.Size
			c.Offset = m1.Offset
			if _, err := part.Read(c.Data); err != nil {
				return nil, err
			}
		}
	}

	return c, nil
}

// ReadChunk reads Chunk data from f.
func (f *RFileHandle) ReadChunk(offset int64, len int64) (*Chunk, error) {
	m := ChunkParam{
		Path:   f.Path,
		Offset: offset,
		Size:   len,
	}

	return getChunkData(m)
}

// GetChunkMeta gets ChunkMeta info from f.
func (f *RFileHandle) GetChunkMeta(offset int64, len int64) (*ChunkMeta, error) {
	return remoteChunkMeta(f.Path, offset, len)
}

// WriteChunk writes chunk data to f.
func (f *RFileHandle) WriteChunk(c *Chunk) error {
	t := fmt.Sprintf("%s/api/v1/pfs/storage/chunks", Config.ActiveConfig.Endpoint)
	log.V(4).Infoln("chunk's URI:" + t)

	p := ChunkParam{
		Path:   f.Path,
		Offset: c.Offset,
		Size:   c.Len,
	}
	param := p.ToURLParam().Encode()

	body, err := restclient.PostChunk(t, param,
		bytes.NewReader(c.Data), c.Len, DefaultMultiPartBoundary)

	if err != nil {
		return err
	}

	log.V(5).Info("received body:" + string(body[:]))

	resp := uploadChunkResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if len(resp.Err) == 0 {
		return nil
	}

	return errors.New(resp.Err)
}

// Close is function not need implement.
func (f *RFileHandle) Close() {
}
