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

// RemoteFile is a remote file's handle.
type RemoteFile struct {
	Path string
	Flag int
	Size int64
}

// Open file to read ,write or read-write.
// if flag == WriteOnly or flag == ReadAndWrite, this function will
// attempt to create a sized file on remote if it does't exist.
func (f *RemoteFile) Open(path string, flag int, size int64) error {
	if flag != os.O_RDONLY &&
		flag != os.O_WRONLY &&
		flag != os.O_RDWR {
		return errors.New("only support os.O_RDONLY, os.O_WRONLY, os.O_RDWR")
	}

	f.Path = path
	f.Flag = flag
	f.Size = size

	if flag == os.O_WRONLY ||
		flag == os.O_RDWR {

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
	t := fmt.Sprintf("%s/%s", Config.ActiveConfig.Endpoint, RESTChunksStoragePath)
	log.V(1).Info("target url: " + t)

	resp, err := restclient.GetChunk(t, m.ToURLParam())
	if err != nil {
		return nil, err
	}
	defer Close(resp.Body)

	if resp.Status != restclient.HTTPOK {
		return nil, errors.New("http server returned non-200 status: " + resp.Status)
	}

	var c = &Chunk{}
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

			c = NewChunk(m1.Size)
			c.Len = m1.Size
			c.Offset = m1.Offset
			if _, err := part.Read(c.Data); err != nil && err != io.EOF {
				return nil, err
			}
		}
	}

	return c, nil
}

// ReadChunk reads Chunk data from f.
func (f *RemoteFile) ReadChunk(offset int64, len int64) (*Chunk, error) {
	if len == 0 {
		return &Chunk{}, nil
	}

	m := ChunkParam{
		Path:   f.Path,
		Offset: offset,
		Size:   len,
	}

	return getChunkData(m)
}

// GetChunkMeta gets ChunkMeta info from f.
func (f *RemoteFile) GetChunkMeta(offset int64, len int64) (*ChunkMeta, error) {
	if len == 0 {
		return &ChunkMeta{}, nil
	}

	return remoteChunkMeta(f.Path, offset, len)
}

// WriteChunk writes chunk data to f.
func (f *RemoteFile) WriteChunk(c *Chunk) error {
	t := fmt.Sprintf("%s/%s", Config.ActiveConfig.Endpoint, RESTChunksStoragePath)
	log.V(3).Infoln("chunk's URI:" + t)

	p := ChunkParam{
		Path:   f.Path,
		Offset: c.Offset,
		Size:   c.Len,
	}
	log.V(3).Infof("write chunk param:%v\n", p)
	param := p.ToURLParam().Encode()

	body, err := restclient.PostChunk(t, param,
		bytes.NewReader(c.Data), c.Len, DefaultMultiPartBoundary)

	if err != nil {
		return err
	}

	log.V(3).Info("received body:" + string(body[:]))

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
func (f *RemoteFile) Close() {
}
