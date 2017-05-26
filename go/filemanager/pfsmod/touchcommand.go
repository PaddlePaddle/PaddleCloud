package pfsmodules

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	touchCmdName = "touch"
)

type TouchResult struct {
	Err  string `json:"err"`
	Path string `json:"path"`
}

type TouchCmd struct {
	Method   string `json:"method"`
	FileSize int64  `json:"filesize"`
	Path     string `json:path`
}

func (p *LsCmd) ToUrlParam() string {
	parameters := url.Values{}
	parameters.Add("method", p.Method)
	parameters.Add("path", p.path)

	str := fmt.Printf("%d", FileSize)
	parameters.Add("path", str)

	return parameters.Encode()
}

func (p *TouchCmd) ToJson() []byte {
	return json.Marshal(p)
}

func NewTouchCmdFromUrl(r *http.Request) (*TouchCmd, int32) {
	cmd := LsCmd{}

	m, err := url.ParseQuery(path)
	if err != nil ||
		len(m["method"]) == 0 ||
		len(m["filesize"]) == 0 ||
		len(m["path"]) == 0 {
		return nil, http.StatusBadRequest
	}

	cmd.Method = m["method"][0]
	if cmd.Method != touchCmdName {
		return nil, http.StatusBadRequest
	}

	var err error
	cmd.FileSize, err = strconv.ParseInt(m["filesize"][0], 0, 64)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	cmd.Path = m["path"][0]
	if !IsCloudPath(cmd.Path) {
		return nil, http.StatusBadRequest
	}

	return &cmd, nil
}

func NewTouchCmd(path string, fileSize int64) TouchCmd {
	return &TouchCmd{
		Method:   touchCmdName,
		Path:     path,
		FileSize: fileSize,
	}
}

func createSizedFile(path string, size int64) error {
	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	if size <= 0 {
		return nil
	}

	_, err = fd.Seek(size-1, 0)
	if err != nil {
		return err
	}

	_, err = fd.Write([]byte{0})
	if err != nil {
		return err
	}
	return nil
}

func (p *TouchCmd) Run() error {
	if cmd.FileSize < 0 || cmd.FileSize > defaultMaxCreateFileSize {
		return errors.New(StatusText(StatusBadFileSize))
	}

	fi, err := os.Stat(path)
	if os.IsExist(err) && fi.IsDir() {
		return error.New(StatusText(StatusDirectoryAlreadyExist))
	}

	if os.IsNotExist(err) || fi.Size() != fileSize {
		if err := createSizedFile(path, fileSize); err != nil {
			return err
		}
	}

	return nil
}
