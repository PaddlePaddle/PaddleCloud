package main

import (
	"context"
	"flag"
	"fmt"
	pfsmod "github.com/cloud/go/file_manager/pfsmodules"
	"github.com/google/subcommands"
)

type rmCommand struct {
	r bool
}

func (*rmCommand) Name() string     { return "rm" }
func (*rmCommand) Synopsis() string { return "Rm files or directories on PaddlePaddle Cloud" }
func (*rmCommand) Usage() string {
	return `rm [-r] <pfspath>:
	Rm files or directories on PaddlePaddle Cloud
	Options:
`
}

func (p *rmCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.r, "r", false, "rm pfspath recursively")
}

func (p *rmCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmdAttr := pfsmod.NewCmdAttr(p.Name(), f)
	resp := pfsmod.RmCmdResponse{}
	s := NewCmdSubmitter(UserHomeDir() + "/.paddle/config")

	rmCmd := pfsmod.NewRmCmd(cmdAttr, &resp)
	_, err := s.SubmitCmdReqeust("POST", "api/v1/files", 8080, rmCmd)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
