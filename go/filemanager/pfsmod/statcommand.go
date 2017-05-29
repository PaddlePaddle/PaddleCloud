package pfsmod

import (
	"errors"
	"flag"
	log "github.com/golang/glog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

const (
	statCmdName = "ls"
)

type StatCmd struct {
	Method string
	Path   string
}

func (p *StatCmd) ToUrlParam() string {
	parameters := url.Values{}
	parameters.Add("method", p.Method)
	parameters.Add("path", Path)

	return parameters.Encode()

}

func (p *StatCmd) ToJson() ([]byte, error) {
	return nil, nil
}

func NewStatCmdFromUrlParam(path string) (*StatCmd, error) {
	cmd := StatCmd{}

	m, err := url.ParseQuery(path)
	if err != nil ||
		len(m["method"]) == 0 ||
		len(m["path"]) == 0 {
		return nil, errors.New(StatusText(StatusNotEnoughArgs))
	}

	cmd.Method = m["method"][0]
	if cmd.Method != statCmdName {
		return nil, errors.New(http.StatusText(http.StatusMethodNotAllowed) + ":" + cmd.Method)
	}

	cmd.Path = m["path"][0]
	return &cmd, nil
}

func NewStatCmd(path string) *StatCmd {
	return &StatCmd{
		Method: statCmdName,
		Path:   path,
	}
}

func (p *StatCmd) Check() error {
	if !IsCloudPath(p.Path) {
		return errors.New(StatusText(StatusShouldBePfsPath) + ":" + p.Path)
	}

	if !CheckUser(p.Path) {
		return errors.New(StatusText(StatusUnAuthorized) + ":" + arg)
	}

	return nil
}

func (p *StatCmd) Run() (interface{}, error) {
	result := LsResult{}

	fi, err := os.Stat(p.Path)
	if err != nil {
		return nil, err
	}

	return &LsResult{
		Path:    p.Path,
		ModTime: fi.ModTime().Format("2006-01-02 15:04:05"),
		IsDir:   fi.IsDir(),
		Size:    fi.Size(),
	}, nil
}

func IsNotExist(err error) bool {
	return err.Error() == StatusText(StatusFileNotFound)
}
