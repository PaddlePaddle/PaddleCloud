package paddlecloud

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/google/subcommands"
)

// LogsCommand print aggregated job logs
type LogsCommand struct {
	n int
	w string
}

// Name is subcommands name
func (*LogsCommand) Name() string { return "logs" }

// Synopsis is subcommands synopsis
func (*LogsCommand) Synopsis() string { return "Print logs of the job." }

// Usage is subcommands usage
func (*LogsCommand) Usage() string {
	return `logs <job name>:
	Print logs of the job.
	Options:
`
}

// SetFlags registers subcommands flags
func (p *LogsCommand) SetFlags(f *flag.FlagSet) {
	f.IntVar(&p.n, "n", 10, "Number of lines to print from tail.")
	f.StringVar(&p.w, "w", "", "Print logs for a single worker.")
}

// Execute logs command
func (p *LogsCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	var queryMap url.Values
	queryMap.Add("n", strconv.FormatInt(int64(p.n), 10))
	queryMap.Add("w", p.w)
	queryMap.Add("jobname", f.Arg(0))

	respBody, err := GetCall(config.ActiveConfig.Endpoint+"/api/v1/logs", queryMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "call paddle cloud error %v", err)
	}
	var respObj interface{}
	errJSON := json.Unmarshal(respBody, &respObj)
	if errJSON != nil {
		fmt.Fprintf(os.Stderr, "bad server return: %s", respBody)
	}
	fmt.Printf("%s\n", respObj.(map[string]interface{})["msg"].(string))
	return subcommands.ExitSuccess
}
