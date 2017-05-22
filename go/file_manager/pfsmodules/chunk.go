package pfsmodules

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

const (
	defaultMaxChunkSize = 4 * 1024 * 1024
	defaultMinChunkSize = 4 * 1024
	DefaultChunkSize    = 2 * 1024 * 1024 * 1024
)

type Chunk struct {
	Meta ChunkMeta
	Data []byte
}

type ChunkCmdAttr struct {
	Path   string `json:"path"`
	Offset int64  `json:"offset"`
	len    uint32 `json:len`
}

type ChunkCmd struct {
	cmdAttr *ChunkCmdAttr
	resp    *Chunk
}

func WriteChunk(chunk Chunk) {
	fd, err := os.OpenFile(chunk.Meta.Path, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.Seek(chunk.Meta.Offset, 0)
	if err != nil {
		return err
	}

	_, err = fd.Write([]chunk.Meta.Data[:chunk.Meta.Len])
	if err != nil {
		return err
	}

	return nil
}

func GetChunkWriter(path string, int offset) (*File, error) {
	fd, err := os.OpenFile(chunk.Meta.Path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	_, err = fd.Seek(chunk.Meta.Offset, 0)
	if err != nil {
		return nil, err
	}

	return fd, nil
}

func GetChunk(path string, offset int64, len uint32) (*Chunk, error) {
	f, err := os.Open(path) // For read access.
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if len > defaultMaxChunkSize || len < defaultMinChunkSize {
		return nil, errors.New("invalid len:" + string(len))
	}

	/*
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}
	*/

	/*
		if offset < 0 || offset > fi.Size() {
			return nil, Errors.New("offset > filesize:" + string(fi.Size()))
		}
	*/

	chunk := Chunk{}
	chunk.Data = make([]byte, len)
	m := &chunk.Meta

	if _, err := f.Seek(offset, os.SEEK_SET); err != nil {
		return nil, err
	}

	n, err := f.Read(chunk.Data)
	if err != nil && err != io.EOF {
		return nil, err
	}

	m.Offset = offset
	sum := md5.Sum(chunk.Data[:n])
	m.Checksum = hex.EncodeToString(sum[:])
	m.Len = uint32(n)

	return &chunk, nil
}

// https://github.com/gebi/go-fileupload-example/blob/master/main.go
// Streams upload directly from file -> mime/multipart -> pipe -> http-request
func streamingUploadFile(params map[string]string, paramName, path string, w *io.PipeWriter, file *os.File) {
	defer file.Close()
	defer w.Close()
	writer := multipart.NewWriter(w)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatal(err)
		return
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()
	go streamingUploadFile(params, paramName, path, w, file)
	return http.NewRequest("POST", uri, r)
}
