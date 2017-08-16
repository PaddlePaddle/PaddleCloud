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

type RFileHandle struct {
	Path string
	Flag int
	Size int64
}

const (
	ReadOnly, WriteOnly, ReadAndWrite = os.O_RDONLY, os.O_WRONLY, os.O_RDWR
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

func getChunkData(target string, m ChunkParam) (*Chunk, error) {
	log.V(1).Info("target url: " + target)

	resp, err := restclient.GetChunk(target, m.ToURLParam())
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
		part, error := partReader.NextPart()
		if error == io.EOF {
			break
		}

		if part.FormName() == "chunk" {
			m, err := ParseChunkParam(part.FileName())
			if err != nil {
				return nil, errors.New(err.Error())
			}

			c = NewChunk(m.Size)
			c.Len = m.Size
			c.Offset = m.Offset
			if _, err := part.Read(c.Data); err != nil {
				return nil, err
			}
		}
	}

	return c, nil
}
func (f *RFileHandle) Read(offset int64, len int64) (*Chunk, error) {
	/*
		m := &ChunkParam{
			Path:   f.Path,
			Offset: offset,
			Size:   len,
		}
	*/

	//return getChunkData(offset, len)
	return nil, nil
}

func (f *RFileHandle) GetChunkMeta(offset int64, len int64) (*ChunkMeta, error) {
	return remoteChunkMeta(f.Path, offset, len)
}

func (f *RFileHandle) Write(c *Chunk) error {
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

func (f *RFileHandle) Close() {
}
