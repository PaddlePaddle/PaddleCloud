package paddlecloud

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
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
	result, err := RemoteLs(s, cmd)
	if err != nil {
		return subcommands.ExitFailure
	}

	fmt.Println(result)
	return subcommands.ExitSuccess
}
