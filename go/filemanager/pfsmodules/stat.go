package pfsmodules

import (
	"errors"
	"net/http"
	"net/url"
	"os"
)

const (
	// StatCmdName means stat command name.
	StatCmdName = "stat"
)

// StatCmd means stat command.
type StatCmd struct {
	Method string
	Path   string
}

// ToURLParam encodes StatCmd to URL Encoding string.
func (p *StatCmd) ToURLParam() url.Values {
	parameters := url.Values{}
	parameters.Add("method", p.Method)
	parameters.Add("path", p.Path)

	return parameters

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

// ValidateLocalArgs checks the condition when running local.
func (p *StatCmd) ValidateLocalArgs() error {
	panic("not implement")
}

// ValidateCloudArgs checks the conditions when running on cloud.
func (p *StatCmd) ValidateCloudArgs(userName string) error {
	return ValidatePfsPath([]string{p.Path}, userName)
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
