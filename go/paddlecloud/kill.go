package paddlecloud

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

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

	body, err := deleteCall([]byte("{\"jobname\": \""+f.Arg(0)+"\"}"), config.ActiveConfig.Endpoint+"/api/v1/jobs/", token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error kill job: %v\n", err)
		return subcommands.ExitFailure
	}
	var jsonObj interface{}
	err = json.Unmarshal(body, &jsonObj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error kill job: %v\n", err)
		return subcommands.ExitFailure
	}
	respCode := jsonObj.(map[string]interface{})["code"].(float64)
	if respCode != 200 {
		fmt.Fprintf(os.Stderr, "error kill job: %s\n", jsonObj.(map[string]interface{})["msg"].(string))
	}

	return subcommands.ExitSuccess
}
