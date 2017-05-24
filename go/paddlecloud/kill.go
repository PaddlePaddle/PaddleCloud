package paddlecloud

import (
	"context"
	"flag"

	"github.com/google/subcommands"
)

// KillCommand do job killings
type KillCommand struct {
	rm bool
}

// Name is subcommands name
func (*KillCommand) Name() string { return "kill" }

// Synopsis is subcommands synopsis
func (*KillCommand) Synopsis() string { return "Stop the job. -rm will remove the job from history." }

// Usage is subcommands usage
func (*KillCommand) Usage() string {
	return `kill:
	Stop the job. -rm will remove the job from history.
`
}

// SetFlags registers subcommands flags
func (p *KillCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.rm, "rm", false, "remove the job from history")
}

// Execute kill command
func (p *KillCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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
