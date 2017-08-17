package pfsmodules

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/PaddlePaddle/cloud/go/utils/restclient"
	log "github.com/golang/glog"
	"github.com/google/subcommands"
)

const (
	mkdirCmdName = "mkdir"
)

// MkdirResult means Mkdir command's result.
type MkdirResult struct {
	Path string `json:"path"`
}

// MkdirCmd means Mkdir command.
type MkdirCmd struct {
	Method string   `json:"method"`
	Args   []string `json:"path"`
}

// ValidateLocalArgs checks the conditions when running on local.
func (p *MkdirCmd) ValidateLocalArgs() error {
	if len(p.Args) == 0 {
		return errors.New(StatusNotEnoughArgs)
	}
	return nil
}

// ValidateCloudArgs checks the conditions when running on cloud.
func (p *MkdirCmd) ValidateCloudArgs(userName string) error {
	return ValidatePfsPath(p.Args, userName, mkdirCmdName)
}

// ToURLParam need not to be implemented.
func (p *MkdirCmd) ToURLParam() url.Values {
	panic("not implemented")
}

// ToJSON encodes MkdirCmd to JSON string.
func (p *MkdirCmd) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// NewMkdirCmd returns a new MkdirCmd.
func NewMkdirCmd(path string) *MkdirCmd {
	return &MkdirCmd{
		Method: mkdirCmdName,
		Args:   []string{path},
	}
}

// newMkdirCmdFromFlag returns a new MkdirCmd from parsed flags.
func newMkdirCmdFromFlag(f *flag.FlagSet) (*MkdirCmd, error) {
	cmd := MkdirCmd{}

	cmd.Method = mkdirCmdName
	cmd.Args = make([]string, 0, f.NArg())

	for _, arg := range f.Args() {
		log.V(2).Info(arg)
		cmd.Args = append(cmd.Args, arg)
	}

	return &cmd, nil
}

// Run runs MkdirCmd.
func (p *MkdirCmd) Run() (interface{}, error) {
	var results []MkdirResult
	for _, path := range p.Args {
		fi, err := os.Stat(path)

		if os.IsExist(err) && !fi.IsDir() {
			return results, errors.New(StatusAlreadyExist)
		}

		if err := os.MkdirAll(path, 0700); err != nil {
			return results, err
		}

		results = append(results, MkdirResult{Path: path})
	}

	return results, nil
}

// Name returns name of MkdirComand.
func (*MkdirCmd) Name() string { return "mkdir" }

// Synopsis returns synopsis of MkdirCmd.
func (*MkdirCmd) Synopsis() string { return "mkdir directoies on PaddlePaddle Cloud" }

// Usage returns usage of MkdirCmd.
func (*MkdirCmd) Usage() string {
	return `mkdir <pfspath>:
	mkdir directories on PaddlePaddleCloud
	Options:
`
}

// SetFlags sets MkdirCmd's parameters.
func (p *MkdirCmd) SetFlags(f *flag.FlagSet) {
}

func formatMkdirPrint(results []MkdirResult, err error) {
	if err != nil {
		fmt.Println("\t" + err.Error())
	}
}

// RemoteMkdir creat a directory on cloud.
func RemoteMkdir(cmd *MkdirCmd) ([]MkdirResult, error) {
	j, err := cmd.ToJSON()
	if err != nil {
		return nil, err
	}

	t := fmt.Sprintf("%s/%s", Config.ActiveConfig.Endpoint, RESTFilesPath)
	log.V(2).Infoln(t)
	body, err := restclient.PostCall(t, j)
	if err != nil {
		return nil, err
	}

	log.V(3).Info(string(body[:]))

	type mkdirResponse struct {
		Err     string        `json:"err"`
		Results []MkdirResult `json:"results"`
	}

	resp := mkdirResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp.Results, err
	}

	log.V(1).Infof("%#v\n", resp)

	if len(resp.Err) == 0 {
		return resp.Results, nil
	}

	return resp.Results, errors.New(resp.Err)
}

func remoteMkdir(cmd *MkdirCmd) error {
	for _, arg := range cmd.Args {
		subcmd := NewMkdirCmd(arg)

		fmt.Printf("mkdir %s\n", arg)
		results, err := RemoteMkdir(subcmd)
		formatMkdirPrint(results, err)
	}
	return nil

}

// Execute runs a MkdirCmd.
func (p *MkdirCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd, err := newMkdirCmdFromFlag(f)
	if err != nil {
		return subcommands.ExitFailure
	}
	log.V(1).Infof("%#v\n", cmd)

	if err := remoteMkdir(cmd); err != nil {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
