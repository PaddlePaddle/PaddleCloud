package paddlecloud

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	//log "github.com/golang/glog"
	"github.com/google/subcommands"
)

type LsCommand struct {
	cmd pfsmod.LsCmd
}

func (*LsCommand) Name() string     { return "ls" }
func (*LsCommand) Synopsis() string { return "List files on PaddlePaddle Cloud" }
func (*LsCommand) Usage() string {
	return `ls [-r] <pfspath>:
	List files on PaddlePaddleCloud
	Options:
`
}

func (p *LsCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.cmd.R, "r", false, "list files recursively")
}

func formatPrint(result []pfsmod.LsResult) {
	fmt.Println(result)
}

func RemoteLs(s *PfsSubmitter, cmd *pfsmod.LsCmd) ([]pfsmod.LsResult, error) {
	body, err := s.GetFiles(cmd)
	if err != nil {
		return nil, err
	}

	resp := pfsmod.LsResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp.Results, err
	}

	if len(resp.Err) == 0 {
		return resp.Results, nil
	}

	return resp.Results, errors.New(resp.Err)
}

func remoteLs(s *PfsSubmitter, cmd *pfsmod.LsCmd) error {
	for _, arg := range cmd.Args {
		subcmd := pfsmod.NewLsCmd(
			cmd.R,
			arg,
		)
		result, err := RemoteLs(s, subcmd)

		//fmt.Printf("ls -r=%v %s\n", cmd.R, arg)
		fmt.Printf("%s :\n", arg)
		if err != nil {
			fmt.Printf("  error:%s\n", err.Error())
			return err
		}

		formatPrint(result)
	}
	return nil
}

func (p *LsCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd, err := pfsmod.NewLsCmdFromFlag(f)
	if err != nil {
		return subcommands.ExitFailure
	}

	s := NewPfsCmdSubmitter(UserHomeDir() + "/.paddle/config")
	if err := remoteLs(s, cmd); err != nil {
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
