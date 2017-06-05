package paddlecloud

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
	log "github.com/golang/glog"
	"github.com/google/subcommands"
)

// RmCommand represents remove command.
type RmCommand struct {
	cmd pfsmod.RmCmd
}

// Name returns RmCommand's name.
func (*RmCommand) Name() string { return "rm" }

// Synopsis returns synopsis of RmCommand.
func (*RmCommand) Synopsis() string { return "rm files on PaddlePaddle Cloud" }

// Usage returns usage of RmCommand.
func (*RmCommand) Usage() string {
	return `rm -r <pfspath>:
	rm files on PaddlePaddleCloud
	Options:
`
}

// SetFlags sets RmCommand's parameters.
func (p *RmCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.cmd.R, "r", false, "rm files recursively")
}

func formatRmPrint(results []pfsmod.RmResult, err error) {
	for _, result := range results {
		fmt.Printf("rm %s\n", result.Path)
	}

	if err != nil {
		fmt.Println("\t" + err.Error())
	}

	return
}

// RemoteRm gets RmCmd Result from cloud.
func RemoteRm(cmd *pfsmod.RmCmd) ([]pfsmod.RmResult, error) {
	j, err := cmd.ToJSON()
	if err != nil {
		return nil, err
	}

	t := fmt.Sprintf("%s/api/v1/files", config.ActiveConfig.Endpoint)
	body, err := PostCall(t, j)
	if err != nil {
		return nil, err
	}

	log.V(3).Info(string(body[:]))

	type rmResponse struct {
		Err     string            `json:"err"`
		Results []pfsmod.RmResult `json:"path"`
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

func remoteRm(cmd *pfsmod.RmCmd) error {
	for _, arg := range cmd.Args {
		subcmd := pfsmod.NewRmCmd(
			cmd.R,
			arg,
		)

		fmt.Printf("rm %s\n", arg)
		result, err := RemoteRm(subcmd)
		formatRmPrint(result, err)
	}
	return nil

}

// Execute runs a RmCommand.
func (p *RmCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd, err := pfsmod.NewRmCmdFromFlag(f)
	if err != nil {
		return subcommands.ExitFailure
	}
	log.V(1).Infof("%#v\n", cmd)

	if err := remoteRm(cmd); err != nil {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
