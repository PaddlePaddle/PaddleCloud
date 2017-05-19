package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/google/subcommands"
)

type logsCommand struct {
	n int
	w string
}

func (*logsCommand) Name() string     { return "logs" }
func (*logsCommand) Synopsis() string { return "Print logs of the job." }
func (*logsCommand) Usage() string {
	return `logs <job name>:
	Print logs of the job.
	Options:
`
}

func (p *logsCommand) SetFlags(f *flag.FlagSet) {
	f.IntVar(&p.n, "n", 10, "Number of lines to print from tail.")
	f.StringVar(&p.w, "w", "", "Print logs for a single worker.")
}

func (p *logsCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return subcommands.ExitFailure
	}
	token, err := token()
	if err != nil {
		return subcommands.ExitFailure
	}
	queryMap := make(map[string]string)

	queryMap["n"] = strconv.FormatInt(int64(p.n), 10)
	queryMap["w"] = p.w
	queryMap["jobname"] = f.Arg(0)
	respBody, err := getCall(config.ActiveConfig.Endpoint+"/api/v1/logs", queryMap, token)
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
