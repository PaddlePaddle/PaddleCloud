package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/subcommands"
)

type getCommand struct {
	a bool
}

func (*getCommand) Name() string     { return "get" }
func (*getCommand) Synopsis() string { return "Print resources" }
func (*getCommand) Usage() string {
	return `get:
	Print resources.
`
}

func (p *getCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.a, "a", false, "Get all resources.")
}

func (p *getCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return subcommands.ExitFailure
	}
	if f.Arg(0) == "jobs" {
		Jobs()
	} else if f.Arg(0) == "quota" {
		fmt.Println("show quota")
	} else if f.Arg(0) == "workers" {
		fmt.Println("show workers")
	}

	return subcommands.ExitSuccess
}

// Jobs prints current job list
func Jobs() error {
	token, err := token()
	if err != nil {
		return err
	}
	respBody, err := getCall(config.ActiveConfig.Endpoint+"/api/v1/jobs/", token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting jobs: %v\n", err)
	}
	var respObj interface{}
	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		return err
	}
	items := respObj.(map[string]interface{})["items"].([]interface{})
	if len(items) >= 0 {
		fmt.Printf("NUM\tNAME\tSUCC\tFAIL\tSTART\tCOMP\tACTIVE\n")
	}
	for idx, j := range items {
		jobnameTrainer := j.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		jobnameParts := strings.Split(jobnameTrainer, "-")
		jobname := strings.Join(jobnameParts[0:len(jobnameParts)-1], "-")

		fmt.Printf("%d\t%s\t%v\t%v\t%v\t%v\t%v\n", idx,
			jobname,
			j.(map[string]interface{})["status"].(map[string]interface{})["succeeded"],
			j.(map[string]interface{})["status"].(map[string]interface{})["failed"],
			j.(map[string]interface{})["status"].(map[string]interface{})["start_time"],
			j.(map[string]interface{})["status"].(map[string]interface{})["completion_time"],
			j.(map[string]interface{})["status"].(map[string]interface{})["active"])
	}

	return err
}
