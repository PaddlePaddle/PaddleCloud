package pfsmod

import (
	"encoding/json"
	"errors"
	//"fmt"
	//"net/http"
	//"net/url"
	"os"
	//"strconv"
	"flag"
	log "github.com/golang/glog"
)

const (
	mkdirCmdName = "mkdir"
)

type MkdirResult struct {
	Path string `json:"path"`
}

type MkdirCmd struct {
	Method string   `json:"method"`
	Args   []string `json:path`
}

func (p *MkdirCmd) LocalCheck() error {
	if len(p.Args) == 0 {
		return errors.New(StatusText(StatusNotEnoughArgs))
	}
	return nil
}

func (p *MkdirCmd) CloudCheck() error {
	if len(p.Args) == 0 {
		return errors.New(StatusText(StatusNotEnoughArgs))
	}

	for _, arg := range p.Args {
		if !IsCloudPath(arg) {
			return errors.New(StatusText(StatusShouldBePfsPath) + ":" + arg)
		}

		if !CheckUser(arg) {
			return errors.New(StatusText(StatusShouldBePfsPath) + ":" + arg)
		}
	}

	return nil

}

func (p *MkdirCmd) ToUrlParam() string {
	return ""
}

func (p *MkdirCmd) ToJson() ([]byte, error) {
	return json.Marshal(p)
}

func NewMkdirCmd(path string) *MkdirCmd {
	return &MkdirCmd{
		Method: mkdirCmdName,
		Args:   []string{path},
	}
}
func NewMkdirCmdFromFlag(f *flag.FlagSet) (*MkdirCmd, error) {
	cmd := MkdirCmd{}

	cmd.Method = mkdirCmdName
	cmd.Args = make([]string, 0, f.NArg())

	for _, arg := range f.Args() {
		log.V(2).Info(arg)
		cmd.Args = append(cmd.Args, arg)
	}

	return &cmd, nil
}

func (p *MkdirCmd) Run() (interface{}, error) {
	results := make([]MkdirResult, 0, 100)
	for _, path := range p.Args {
		fi, err := os.Stat(path)

		if os.IsExist(err) && !fi.IsDir() {
			return results, errors.New(StatusText(StatusAlreadyExist))
		}

		if err := os.MkdirAll(path, 0755); err != nil {
			return results, err
		}

		results = append(results, MkdirResult{Path: path})
	}

	return results, nil
}
