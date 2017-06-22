package paddlecloud

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/PaddlePaddle/cloud/go/utils"
	"github.com/google/subcommands"
)

// GetCommand exports get subcommand for fetching status
type GetCommand struct {
	a bool
}

// Name is subcommands name
func (*GetCommand) Name() string { return "get" }

// Synopsis is subcommands synopsis
func (*GetCommand) Synopsis() string { return "Print resources" }

// Usage is subcommands usage
func (*GetCommand) Usage() string {
	return `get [jobs|workers|registry [jobname]|quota]:
	Print resources.
`
}

// SetFlags registers subcommands flags
func (p *GetCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.a, "a", false, "Get all resources.")
}

// Execute get command
func (p *GetCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 || f.NArg() > 2 {
		f.Usage()
		return subcommands.ExitFailure
	}
	if f.Arg(0) == "jobs" {
		jobs()
	} else if f.Arg(0) == "quota" {
		quota()
	} else if f.Arg(0) == "registry" {
		registry()
	} else if f.Arg(0) == "workers" {
		if f.NArg() != 2 {
			f.Usage()
			return subcommands.ExitFailure
		}
		workers(f.Arg(1))
	} else {
		f.Usage()
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func workers(jobname string) error {
	var queryMap url.Values
	queryMap.Add("jobname", jobname)
	respBody, err := utils.GetCall(utils.Config.ActiveConfig.Endpoint+"/api/v1/workers/", queryMap)
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
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "NAME\tSTATUS\tSTART\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing: %s", err)
		return err
	}
	for _, item := range respObj.(map[string]interface{})["items"].([]interface{}) {
		fmt.Fprintf(w, "%s\t%s\t%v\t\n",
			item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string),
			item.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string),
			item.(map[string]interface{})["status"].(map[string]interface{})["start_time"])
	}
	w.Flush()
	return nil
}
func registry() error {
	respBody, err := utils.GetCall(utils.Config.ActiveConfig.Endpoint+"/api/v1/registry/", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err getting registry secret: %v\n", err)
		return err
	}
	var respObj interface{}
	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		return err
	}
	items := respObj.(map[string]interface{})["msg"].(map[string]interface{})["items"].([]interface{})
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if len(items) >= 0 {
		fmt.Fprintf(w, "ID\tNAME\tDATA\n")
	}
	idx := 0
	for _, r := range items {
		metadata := r.(map[string]interface{})["metadata"].(map[string]interface{})
		name := RegistryName(metadata["name"].(string))
		if len(name) != 0 {
			cTime := metadata["creation_timestamp"].(string)
			fmt.Fprintf(w, "%d\t%s\t%s\n", idx, name, cTime)
			idx++
		}
	}
	w.Flush()
	return err
}
func jobs() error {
	respBody, err := utils.GetCall(utils.Config.ActiveConfig.Endpoint+"/api/v1/jobs/", nil)
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
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if len(items) >= 0 {
		fmt.Fprintf(w, "NUM\tNAME\tSUCC\tFAIL\tSTART\tCOMP\tACTIVE\t\n")
	}
	for idx, j := range items {
		jobnameTrainer := j.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		jobnameParts := strings.Split(jobnameTrainer, "-")
		jobname := strings.Join(jobnameParts[0:len(jobnameParts)-1], "-")

		fmt.Fprintf(w, "%d\t%s\t%v\t%v\t%v\t%v\t%v\t\n", idx,
			jobname,
			j.(map[string]interface{})["status"].(map[string]interface{})["succeeded"],
			j.(map[string]interface{})["status"].(map[string]interface{})["failed"],
			j.(map[string]interface{})["status"].(map[string]interface{})["start_time"],
			j.(map[string]interface{})["status"].(map[string]interface{})["completion_time"],
			j.(map[string]interface{})["status"].(map[string]interface{})["active"])
	}
	w.Flush()

	return err
}

func quota() error {
	respBody, err := utils.GetCall(utils.Config.ActiveConfig.Endpoint+"/api/v1/quota/", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting quota: %v\n", err)
		return err
	}
	var respObj interface{}
	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "RESOURCE\tLIMIT\t\n")
	for _, item := range respObj.(map[string]interface{})["items"].([]interface{}) {
		fmt.Fprintf(w, "-----\t-----\t\n")
		hardLimits := item.(map[string]interface{})["status"].(map[string]interface{})["hard"].(map[string]interface{})
		for k, v := range hardLimits {
			fmt.Fprintf(w, "%s\t%s\t\n", k, v.(string))
		}
	}
	w.Flush()
	return nil
}
