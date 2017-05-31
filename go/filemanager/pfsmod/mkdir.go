package pfsmod

import (
	"encoding/json"
	"errors"
	"flag"
	"os"

	log "github.com/golang/glog"
)

const (
	mkdirCmdName = "mkdir"
)

// MkdirResult means Mkdir command's result
type MkdirResult struct {
	Path string `json:"path"`
}

// MkdirCmd means Mkdir command
type MkdirCmd struct {
	Method string   `json:"method"`
	Args   []string `json:"path"`
}

// LocalCheck checks the conditions when running on local
func (p *MkdirCmd) LocalCheck() error {
	if len(p.Args) == 0 {
		return errors.New(StatusText(StatusNotEnoughArgs))
	}
	return nil
}

// CloudCheck checks the conditions when running on cloud
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

// ToURLParam need not to be implemented
func (p *MkdirCmd) ToURLParam() string {
	panic("not implemented")
}

// ToJSON encodes MkdirCmd to JSON string
func (p *MkdirCmd) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// NewMkdirCmd returns a new MkdirCmd
func NewMkdirCmd(path string) *MkdirCmd {
	return &MkdirCmd{
		Method: mkdirCmdName,
		Args:   []string{path},
	}
}

// NewMkdirCmdFromFlag returns a new MkdirCmd from parsed flags
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

// Run runs MkdirCmd
func (p *MkdirCmd) Run() (interface{}, error) {
	results := make([]MkdirResult, 0, 100)
	for _, path := range p.Args {
		fi, err := os.Stat(path)

		if os.IsExist(err) && !fi.IsDir() {
			return results, errors.New(StatusText(StatusAlreadyExist))
		}

		if err := os.MkdirAll(path, 0700); err != nil {
			return results, err
		}

		results = append(results, MkdirResult{Path: path})
	}

	return results, nil
}
