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

	"github.com/PaddlePaddle/cloud/go/utils/restclient"
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
	queryMap := url.Values{}
	queryMap.Add("jobname", jobname)
	respBody, err := restclient.GetCall(Config.ActiveConfig.Endpoint+"/api/v1/workers/", queryMap)
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

	fmt.Fprintln(w, "NAME\tSTATUS\tSTART\tEXIT_CODE\tMSG\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing: %s", err)
		return err
	}
	for _, item := range respObj.(map[string]interface{})["items"].([]interface{}) {
		var exitCode, msg interface{}
		terminateState := item.(map[string]interface{})["status"].(map[string]interface{})["container_statuses"].([]interface{})[0].(map[string]interface{})["state"].(map[string]interface{})["terminated"]

		if terminateState != nil {
			exitCode = terminateState.(map[string]interface{})["exit_code"]
			msg = terminateState.(map[string]interface{})["message"]
		}

		fmt.Fprintf(w, "%s\t%s\t%v\t%v\t%v\t\n",
			item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string),
			item.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string),
			item.(map[string]interface{})["status"].(map[string]interface{})["start_time"],
			exitCode, msg)
	}
	w.Flush()
	return nil
}
func registry() error {
	respBody, err := restclient.GetCall(Config.ActiveConfig.Endpoint+"/api/v1/registry/", nil)
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
	// NOTE: a job include pserver replicaset and a trainers job, display them
	// get pserver replicaset
	//          "status": {
	//              "available_replicas": 1,
	//              "conditions": null,
	//              "fully_labeled_replicas": 1,
	//              "observed_generation": 1,
	//              "ready_replicas": 1,
	//              "replicas": 1
	var respObj interface{}

	respBody, err := restclient.GetCall(Config.ActiveConfig.Endpoint+"/api/v1/pservers/", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting pservers: %v\n", err)
		return err
	}
	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		return err
	}
	pserverItems := respObj.(map[string]interface{})["items"].([]interface{})

	// get kubernetes jobs info
	respBody, err = restclient.GetCall(Config.ActiveConfig.Endpoint+"/api/v1/jobs/", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting jobs: %v\n", err)
		return err
	}

	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		return err
	}
	items := respObj.(map[string]interface{})["items"].([]interface{})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if len(items) >= 0 {
		fmt.Fprintf(w, "NAME\tACTIVE\tSUCC\tFAIL\tSTART\tCOMP\tPS_NAME\tPS_READY\tPS_TOTAL\t\n")
	}
	for _, j := range items {
		jobnameTrainer := j.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		jobnameParts := strings.Split(jobnameTrainer, "-")
		jobname := strings.Join(jobnameParts[0:len(jobnameParts)-1], "-")
		// get info for job pservers
		var psrsname string
		var readyReplicas, replicas interface{}
		for _, psrs := range pserverItems {
			psrsname = psrs.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			if psrsname == jobname+"-pserver" {
				readyReplicas = psrs.(map[string]interface{})["status"].(map[string]interface{})["ready_replicas"]
				replicas = psrs.(map[string]interface{})["status"].(map[string]interface{})["replicas"]
				break
			}
		}

		fmt.Fprintf(w, "%s\t%v\t%v\t%v\t%v\t%v\t%s\t%v\t%v\t\n",
			jobname,
			j.(map[string]interface{})["status"].(map[string]interface{})["active"],
			j.(map[string]interface{})["status"].(map[string]interface{})["succeeded"],
			j.(map[string]interface{})["status"].(map[string]interface{})["failed"],
			j.(map[string]interface{})["status"].(map[string]interface{})["start_time"],
			j.(map[string]interface{})["status"].(map[string]interface{})["completion_time"],
			psrsname, readyReplicas, replicas)
	}
	w.Flush()

	return err
}

func quota() error {
	respBody, err := restclient.GetCall(Config.ActiveConfig.Endpoint+"/api/v1/quota/", nil)
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
