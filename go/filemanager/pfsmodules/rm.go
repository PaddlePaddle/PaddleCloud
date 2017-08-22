package pfsmodules

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/PaddlePaddle/cloud/go/utils/restclient"
	log "github.com/golang/glog"
	"github.com/google/subcommands"
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

// ValidateLocalArgs checks the conditions when running local.
func (p *RmCmd) ValidateLocalArgs() error {
	if len(p.Args) == 0 {
		return errors.New(StatusInvalidArgs)
	}
	return nil
}

// ValidateCloudArgs checks the conditions when running on cloud.
func (p *RmCmd) ValidateCloudArgs(userName string) error {
	return ValidatePfsPath(p.Args, userName, rmCmdName)
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

func newRmCmdFromFlag(f *flag.FlagSet) (*RmCmd, error) {
	cmd := RmCmd{}

	cmd.Method = rmCmdName
	cmd.Args = make([]string, 0, f.NArg())

	var err error
	f.Visit(func(flag *flag.Flag) {
		if flag.Name == "r" {
			cmd.R, err = strconv.ParseBool(flag.Value.String())
			if err != nil {
				log.Errorln("meets error when parsing argument r")
				return
			}
		}
	})

	if err != nil {
		return nil, err
	}

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

// Name returns RmCmd's name.
func (*RmCmd) Name() string { return "rm" }

// Synopsis returns synopsis of RmCmd.
func (*RmCmd) Synopsis() string { return "rm files on PaddlePaddle Cloud" }

// Usage returns usage of RmCmd.
func (*RmCmd) Usage() string {
	return `rm -r <pfspath>:
	rm files on PaddlePaddleCloud
	Options:
`
}

// SetFlags sets RmCmd's parameters.
func (p *RmCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.R, "r", false, "rm files recursively")
}

func formatRmPrint(results []RmResult, err error) {
	for _, result := range results {
		fmt.Printf("rm %s\n", result.Path)
	}

	if err != nil {
		fmt.Println("\t" + err.Error())
	}

	return
}

// RemoteRm gets RmCmd Result from cloud.
func RemoteRm(cmd *RmCmd) ([]RmResult, error) {
	j, err := cmd.ToJSON()
	if err != nil {
		return nil, err
	}

	t := fmt.Sprintf("%s/%s", Config.ActiveConfig.Endpoint, RESTFilesPath)
	body, err := restclient.DeleteCall(t, j)
	if err != nil {
		return nil, err
	}

	log.V(3).Info(string(body[:]))

	type rmResponse struct {
		Err     string     `json:"err"`
		Results []RmResult `json:"path"`
	}

	resp := rmResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp.Results, err
	}

	log.V(1).Infof("%#v\n", resp)

	if len(resp.Err) == 0 {
		return resp.Results, nil
	}

	return resp.Results, errors.New(resp.Err)
}

func remoteRm(cmd *RmCmd) error {
	for _, arg := range cmd.Args {
		subcmd := NewRmCmd(
			cmd.R,
			arg,
		)

		fmt.Printf("rm %s\n", arg)
		result, err := RemoteRm(subcmd)
		formatRmPrint(result, err)
	}
	return nil

}

// Execute runs a RmCmd.
func (p *RmCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd, err := newRmCmdFromFlag(f)
	if err != nil {
		return subcommands.ExitFailure
	}
	log.V(1).Infof("%#v\n", cmd)

	if err := remoteRm(cmd); err != nil {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
