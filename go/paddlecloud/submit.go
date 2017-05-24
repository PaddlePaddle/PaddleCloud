package paddlecloud

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/google/subcommands"
)

// SubmitCmd define the subcommand of submitting paddle training jobs
type SubmitCmd struct {
	Jobname     string `json:"name"`
	Jobpackage  string `json:"jobPackage"`
	Parallelism int    `json:"parallelism"`
	CPU         int    `json:"cpu"`
	GPU         int    `json:"gpu"`
	Memory      string `json:"memory"`
	Pservers    int    `json:"pservers"`
	PSCPU       int    `json:"pscpu"`
	PSMemory    string `json:"psmemory"`
	Entry       string `json:"entry"`
	Topology    string `json:"topology"`
	Datacenter  string `json:"datacenter"`
}

// Name is subcommands name
func (*SubmitCmd) Name() string { return "submit" }

// Synopsis is subcommands synopsis
func (*SubmitCmd) Synopsis() string { return "Submit job to PaddlePaddle Cloud." }

// Usage is subcommands Usage
func (*SubmitCmd) Usage() string {
	return `submit [options] <package path>:
	Submit job to PaddlePaddle Cloud.
	Options:
`
}

// SetFlags registers subcommands flags
func (p *SubmitCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.Jobname, "jobname", "paddle-cluster-job", "Cluster job name.")
	f.IntVar(&p.Parallelism, "parallelism", 1, "Number of parrallel trainers. Defaults to 1.")
	f.IntVar(&p.CPU, "cpu", 1, "CPU resource each trainer will use. Defaults to 1.")
	f.IntVar(&p.GPU, "gpu", 0, "GPU resource each trainer will use. Defaults to 0.")
	f.StringVar(&p.Memory, "memory", "1Gi", " Memory resource each trainer will use. Defaults to 1Gi.")
	f.IntVar(&p.Pservers, "pservers", 0, "Number of parameter servers. Defaults equal to -p")
	f.IntVar(&p.PSCPU, "pscpu", 1, "Parameter server CPU resource. Defaults to 1.")
	f.StringVar(&p.PSMemory, "psmemory", "1Gi", "Parameter server momory resource. Defaults to 1Gi.")
	f.StringVar(&p.Entry, "entry", "paddle train", "Command of starting trainer process. Defaults to paddle train")
	f.StringVar(&p.Topology, "topology", "", "Will Be Deprecated .py file contains paddle v1 job configs")
}

// Execute submit command
func (p *SubmitCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return subcommands.ExitFailure
	}
	// default pservers count equals to trainers count
	if p.Pservers == 0 {
		p.Pservers = p.Parallelism
	}
	p.Jobpackage = f.Arg(0)
	p.Datacenter = config.ActiveConfig.Name

	s := NewSubmitter(p)
	errS := s.Submit(f.Arg(0))
	if errS != nil {
		fmt.Fprintf(os.Stderr, "error submiting job: %v\n", errS)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

// Submitter submit job to cloud
type Submitter struct {
	args *SubmitCmd
}

// NewSubmitter returns a submitter object
func NewSubmitter(cmd *SubmitCmd) *Submitter {
	s := Submitter{cmd}
	return &s
}

// Submit current job
func (s *Submitter) Submit(jobPackage string) error {
	token, err := token()
	if err != nil {
		return err
	}
	// 1. upload user job package to pfs
	filepath.Walk(jobPackage, func(path string, info os.FileInfo, err error) error {
		glog.V(10).Infof("Uploading %s...\n", path)
		return nil
		//return postFile(path, config.activeConfig.endpoint+"/api/v1/files")
	})
	// 2. call paddlecloud server to create kubernetes job
	jsonString, err := json.Marshal(s.args)
	if err != nil {
		return err
	}
	glog.V(10).Infof("Submitting job: %s to %s\n", jsonString, config.ActiveConfig.Endpoint+"/api/v1/jobs")
	respBody, err := postCall(jsonString, config.ActiveConfig.Endpoint+"/api/v1/jobs/",
		token)
	glog.V(10).Infof("got return body size: %d", len(respBody))
	return err
}
