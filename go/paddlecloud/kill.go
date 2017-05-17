package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
)

type killCommand struct {
	rm bool
}

func (*killCommand) Name() string     { return "kill" }
func (*killCommand) Synopsis() string { return "Stop the job. -rm will remove the job from history." }
func (*killCommand) Usage() string {
	return `kill:
	Stop the job. -rm will remove the job from history.
`
}

func (p *killCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.rm, "rm", false, "remove the job from history")
}

func (p *killCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return subcommands.ExitFailure
	}
	token, err := token()
	if err != nil {
		return subcommands.ExitFailure
	}

	deleteCall([]byte("{\"jobname\": \""+f.Arg(0)+"\"}"), config.ActiveConfig.Endpoint+"/api/v1/jobs/", token)
	return subcommands.ExitSuccess
}
