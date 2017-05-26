package paddlecloud

import (
	"context"
	"flag"
	"fmt"
	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
	"github.com/google/subcommands"
)

type LsCommand struct {
	cmd pfsmod.LsCommand
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

func RemoteLs(s *PfsSubmitter, cmd *LsCommand) (*pfsmod.LsCommandResult, error) {
	body, err := s.GetFiles(cmd)
	if err != nil {
		return nil, err
	}

	result := pfsmod.LsCommandResult{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	return &result.(pfsmod.LsCommandResult), nil
}

func (p *LsCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd, err := pfsmod.NewLsCommand(f)
	if err != nil {
		return err
	}

	s := NewPfsSubmitter(UserHomeDir() + "/.paddle/config")
	result, err := RemoteLs(s, cmd)
	if err != nil {
		return subcommands.ExitFailure
	}

	fmt.Println(result)
	return subcommands.ExitSuccess
}
