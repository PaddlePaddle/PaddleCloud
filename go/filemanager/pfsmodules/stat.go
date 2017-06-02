package pfsmodules

import (
	"errors"
	"net/http"
	"net/url"
	"os"
)

const (
	StatCmdName = "stat"
)

// StatCmd means stat command.
type StatCmd struct {
	Method string
	Path   string
}

// ToURLParam encodes StatCmd to URL Encoding string.
func (p *StatCmd) ToURLParam() string {
	parameters := url.Values{}
	parameters.Add("method", p.Method)
	parameters.Add("path", p.Path)

	return parameters.Encode()

}

// ToJSON here need not tobe implemented.
func (p *StatCmd) ToJSON() ([]byte, error) {
	return nil, nil
}

// NewStatCmdFromURLParam return a new StatCmd.
func NewStatCmdFromURLParam(path string) (*StatCmd, error) {
	cmd := StatCmd{}

	m, err := url.ParseQuery(path)
	if err != nil ||
		len(m["method"]) == 0 ||
		len(m["path"]) == 0 {
		return nil, errors.New(StatusNotEnoughArgs)
	}

	cmd.Method = m["method"][0]
	if cmd.Method != StatCmdName {
		return nil, errors.New(http.StatusText(http.StatusMethodNotAllowed) + ":" + cmd.Method)
	}

	cmd.Path = m["path"][0]
	return &cmd, nil
}

// LocalCheck checks the condition when running local.
func (p *StatCmd) LocalCheck() error {
	return nil
}

// CloudCheck checks the conditions when running on cloud.
func (p *StatCmd) CloudCheck() error {
	if !IsCloudPath(p.Path) {
		return errors.New(StatusShouldBePfsPath + ":" + p.Path)
	}

	if !CheckUser(p.Path) {
		return errors.New(StatusUnAuthorized + ":" + p.Path)
	}

	return nil
}

// Run runs the StatCmd.
func (p *StatCmd) Run() (interface{}, error) {
	fi, err := os.Stat(p.Path)
	if err != nil {
		return nil, err
	}

	return &LsResult{
		Path:    p.Path,
		ModTime: fi.ModTime().UnixNano(),
		IsDir:   fi.IsDir(),
		Size:    fi.Size(),
	}, nil
}
