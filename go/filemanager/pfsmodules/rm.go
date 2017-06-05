package pfsmodules

import (
	"encoding/json"
	"errors"
	"flag"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/golang/glog"
)

const (
	rmCmdName = "rm"
)

// RmResult means Rm-command's result.
type RmResult struct {
	Path string `json:"path"`
}

// RmCmd means Rm command.
type RmCmd struct {
	Method string   `json:"method"`
	R      bool     `json:"r"`
	Args   []string `json:"path"`
}

// LocalCheck checks the conditions when running local.
func (p *RmCmd) ValidateLocalArgs() error {
	if len(p.Args) == 0 {
		return errors.New(StatusInvalidArgs)
	}
	return nil
}

// CloudCheck checks the conditions when running on cloud.
func (p *RmCmd) ValidateCloudArgs(userName string) error {
	return ValidatePfsPath(p.Args, userName)
}

// ToURLParam needs not to be implemented.
func (p *RmCmd) ToURLParam() url.Values {
	panic("not implemented")
}

// ToJSON encodes RmCmd to JSON string.
func (p *RmCmd) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// NewRmCmd returns a new RmCmd.
func NewRmCmd(r bool, path string) *RmCmd {
	return &RmCmd{
		Method: rmCmdName,
		R:      r,
		Args:   []string{path},
	}
}

// NewRmCmdFromFlag returns a new RmCmd from parsed flags.
func NewRmCmdFromFlag(f *flag.FlagSet) (*RmCmd, error) {
	cmd := RmCmd{}

	cmd.Method = rmCmdName
	cmd.Args = make([]string, 0, f.NArg())

	var err error
	f.Visit(func(flag *flag.Flag) {
		if flag.Name == "r" {
			cmd.R, err = strconv.ParseBool(flag.Value.String())
			if err != nil {
				panic(err)
			}
		}
	})

	for _, arg := range f.Args() {
		log.V(2).Info(arg)
		cmd.Args = append(cmd.Args, arg)
	}

	return &cmd, nil
}

// Run runs RmCmd.
func (p *RmCmd) Run() (interface{}, error) {
	var result []RmResult

	for _, path := range p.Args {
		list, err := filepath.Glob(path)
		if err != nil {
			return result, err
		}

		for _, arg := range list {
			fi, err := os.Stat(arg)
			if err != nil {
				return result, err
			}

			if fi.IsDir() && !p.R {
				return result, errors.New(StatusCannotDelDirectory + ":" + arg)
			}

			if err := os.RemoveAll(arg); err != nil {
				return result, err
			}

			result = append(result, RmResult{Path: arg})
		}
	}

	return result, nil
}
