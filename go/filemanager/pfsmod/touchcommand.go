package pfsmod

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const (
	defaultMaxCreateFileSize = int64(4 * 1024 * 1024 * 1024)
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

func (p *TouchCmd) ToUrlParam() string {
	parameters := url.Values{}
	parameters.Add("method", p.Method)
	parameters.Add("path", p.Path)

	str := fmt.Sprint(p.FileSize)
	parameters.Add("path", str)

	return parameters.Encode()
}

func (p *TouchCmd) ToJson() ([]byte, error) {
	return json.Marshal(p)
}

func NewTouchCmdFromUrlParam(path string) (*TouchCmd, int32) {
	cmd := TouchCmd{}

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

	cmd.FileSize, err = strconv.ParseInt(m["filesize"][0], 0, 64)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	cmd.Path = m["path"][0]
	if !IsCloudPath(cmd.Path) {
		return nil, http.StatusBadRequest
	}

	return &cmd, http.StatusOK
}

func NewTouchCmd(path string, fileSize int64) *TouchCmd {
	return &TouchCmd{
		Method:   touchCmdName,
		Path:     path,
		FileSize: fileSize,
	}
}

func CreateSizedFile(path string, size int64) error {
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
	if p.FileSize < 0 || p.FileSize > defaultMaxCreateFileSize {
		return errors.New(StatusText(StatusBadFileSize))
	}

	fi, err := os.Stat(p.Path)
	if os.IsExist(err) && fi.IsDir() {
		return errors.New(StatusText(StatusDirectoryAlreadyExist))
	}

	if os.IsNotExist(err) || fi.Size() != p.FileSize {
		if err := CreateSizedFile(p.Path, p.FileSize); err != nil {
			return err
		}
	}

	return nil
}
