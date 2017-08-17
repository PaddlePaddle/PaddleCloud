package pfsmodules

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/PaddlePaddle/cloud/go/utils/restclient"
)

const (
	defaultMaxCreateFileSize = int64(1 * 1024 * 1024 * 1024 * 1024)
)

const (
	// TouchCmdName is the name of touch command.
	TouchCmdName = "touch"
)

// TouchResult represents touch-command's result.
type TouchResult struct {
	Path string `json:"path"`
}

// TouchCmd is holds touch command's variables.
type TouchCmd struct {
	Method   string `json:"method"`
	FileSize int64  `json:"filesize"`
	Path     string `json:"path"`
}

func (p *TouchCmd) checkFileSize() error {
	if p.FileSize < 0 || p.FileSize > defaultMaxCreateFileSize {
		return errors.New(StatusBadFileSize + ":" + fmt.Sprint(p.FileSize))
	}
	return nil
}

// ValidateLocalArgs check the conditions when running local.
func (p *TouchCmd) ValidateLocalArgs() error {
	return p.checkFileSize()
}

// ValidateCloudArgs checks the conditions when running on cloud.
func (p *TouchCmd) ValidateCloudArgs(userName string) error {
	if err := ValidatePfsPath([]string{p.Path}, userName, TouchCmdName); err != nil {
		return err
	}

	return p.checkFileSize()
}

// ToURLParam encodes a TouchCmd to a URL encoding string.
func (p *TouchCmd) ToURLParam() url.Values {
	parameters := url.Values{}
	parameters.Add("method", p.Method)
	parameters.Add("path", p.Path)

	str := fmt.Sprint(p.FileSize)
	parameters.Add("path", str)

	return parameters
}

// ToJSON encodes a TouchCmd to a JSON string.
func (p *TouchCmd) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// NewTouchCmdFromURLParam return a new TouchCmd with specified path.
func NewTouchCmdFromURLParam(path string) (*TouchCmd, int32) {
	cmd := TouchCmd{}

	m, err := url.ParseQuery(path)
	if err != nil ||
		len(m["method"]) == 0 ||
		len(m["filesize"]) == 0 ||
		len(m["path"]) == 0 {
		return nil, http.StatusBadRequest
	}

	cmd.Method = m["method"][0]
	if cmd.Method != TouchCmdName {
		return nil, http.StatusBadRequest
	}

	cmd.FileSize, err = strconv.ParseInt(m["filesize"][0], 0, 64)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	cmd.Path = m["path"][0]

	return &cmd, http.StatusOK
}

// CreateSizedFile creates a file with specified size.
func CreateSizedFile(path string, size int64) error {
	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	defer Close(fd)

	if size <= 0 {
		return nil
	}

	_, err = fd.Seek(size-1, 0)
	if err != nil {
		return err
	}

	_, err = fd.Write([]byte{0})
	return err
}

// Run is a function runs TouchCmd.
func (p *TouchCmd) Run() (interface{}, error) {
	if p.FileSize < 0 || p.FileSize > defaultMaxCreateFileSize {
		return nil, errors.New(StatusBadFileSize)
	}

	fi, err := os.Stat(p.Path)
	if os.IsExist(err) && fi.IsDir() {
		return nil, errors.New(StatusDirectoryAlreadyExist)
	}

	if os.IsNotExist(err) || fi.Size() != p.FileSize {
		if err := CreateSizedFile(p.Path, p.FileSize); err != nil {
			return nil, err
		}
	}

	return &TouchResult{
		Path: p.Path,
	}, nil
}

func localTouch(cmd *TouchCmd) error {
	if _, err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
func remoteTouch(cmd *TouchCmd) error {
	j, err := cmd.ToJSON()
	if err != nil {
		return err
	}

	t := fmt.Sprintf("%s/%s", Config.ActiveConfig.Endpoint, RESTFilesPath)
	body, err := restclient.PostCall(t, j)
	if err != nil {
		return err
	}

	type touchResponse struct {
		Err     string      `json:"err"`
		Results TouchResult `json:"results"`
	}

	resp := touchResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if len(resp.Err) == 0 {
		return nil
	}

	return errors.New(resp.Err)
}
