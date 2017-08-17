package pfsmodules

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/PaddlePaddle/cloud/go/utils/restclient"
	log "github.com/golang/glog"
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
	return ValidatePfsPath([]string{p.Path}, userName, StatCmdName)
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

func remoteStat(cmd *StatCmd) (*LsResult, error) {
	t := fmt.Sprintf("%s/%s", Config.ActiveConfig.Endpoint, RESTFilesPath)
	log.V(3).Infoln("remotestat target URI:" + t)
	body, err := restclient.GetCall(t, cmd.ToURLParam())
	if err != nil {
		return nil, err
	}

	type statResponse struct {
		Err     string   `json:"err"`
		Results LsResult `json:"results"`
	}

	resp := statResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	log.V(1).Infof("result:%#v\n", resp)

	if len(resp.Err) != 0 {
		return nil, errors.New(resp.Err)
	}

	return &resp.Results, nil
}
