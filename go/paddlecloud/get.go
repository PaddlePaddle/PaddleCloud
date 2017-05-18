package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

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
	if f.NArg() < 1 || f.NArg() > 2 {
		f.Usage()
		return subcommands.ExitFailure
	}
	if f.Arg(0) == "jobs" {
		jobs()
	} else if f.Arg(0) == "quota" {
		quota()
	} else if f.Arg(0) == "workers" {
		workers(f.Arg(1))
	}

	return subcommands.ExitSuccess
}

func workers(jobname string) error {
	token, err := token()
	if err != nil {
		return err
	}
	queryMap := make(map[string]string)
	queryMap["jobname"] = jobname
	respBody, err := getCall(config.ActiveConfig.Endpoint+"/api/v1/workers/", queryMap, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting workers: %v\n", err)
		return err
	}
	var respObj interface{}
	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad server return: %s", respBody)
		return err
	}

	fmt.Printf("NAME\tSTATUS\tSTART\n")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing: %s", err)
		return err
	}
	for _, item := range respObj.(map[string]interface{})["items"].([]interface{}) {
		fmt.Printf("%s\t%s\t%v\n", item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string),
			item.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string),
			item.(map[string]interface{})["status"].(map[string]interface{})["start_time"])
	}
	return nil
}

func jobs() error {
	token, err := token()
	if err != nil {
		return err
	}
	respBody, err := getCall(config.ActiveConfig.Endpoint+"/api/v1/jobs/", nil, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting jobs: %v\n", err)
		return err
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
		fmt.Printf("%d\t%s\t%v\t%v\t%v\t%v\t%v\n", idx,
			j.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string),
			j.(map[string]interface{})["status"].(map[string]interface{})["succeeded"],
			j.(map[string]interface{})["status"].(map[string]interface{})["failed"],
			j.(map[string]interface{})["status"].(map[string]interface{})["start_time"],
			j.(map[string]interface{})["status"].(map[string]interface{})["completion_time"],
			j.(map[string]interface{})["status"].(map[string]interface{})["active"])
	}

	return err
}

func quota() error {
	token, err := token()
	if err != nil {
		return err
	}
	respBody, err := getCall(config.ActiveConfig.Endpoint+"/api/v1/quota/", nil, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting quota: %v\n", err)
		return err
	}
	var respObj interface{}
	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		return err
	}
	fmt.Printf("RESOURCE\tLIMIT\n")
	for _, item := range respObj.(map[string]interface{})["items"].([]interface{}) {
		fmt.Printf("-----\t-----\n")
		hardLimits := item.(map[string]interface{})["status"].(map[string]interface{})["hard"].(map[string]interface{})
		for k, v := range hardLimits {
			fmt.Printf("%s\t%s\n", k, v.(string))
		}

	}
	return nil
}
