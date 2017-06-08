package paddlecloud

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/PaddlePaddle/cloud/go/utils"
	"github.com/golang/glog"
	"github.com/google/subcommands"
)

// SubmitCmd define the subcommand of submitting paddle training jobs.
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
	Passes      int    `json:"passes"`
}

// Name is subcommands name.
func (*SubmitCmd) Name() string { return "submit" }

// Synopsis is subcommands synopsis.
func (*SubmitCmd) Synopsis() string { return "Submit job to PaddlePaddle Cloud." }

// Usage is subcommands Usage.
func (*SubmitCmd) Usage() string {
	return `submit [options] <package path>:
	Submit job to PaddlePaddle Cloud.
	Options:
`
}

// SetFlags registers subcommands flags.
func (p *SubmitCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.Jobname, "jobname", "paddle-cluster-job", "Cluster job name.")
	f.IntVar(&p.Parallelism, "parallelism", 1, "Number of parrallel trainers. Defaults to 1.")
	f.IntVar(&p.CPU, "cpu", 1, "CPU resource each trainer will use. Defaults to 1.")
	f.IntVar(&p.GPU, "gpu", 0, "GPU resource each trainer will use. Defaults to 0.")
	f.StringVar(&p.Memory, "memory", "1Gi", " Memory resource each trainer will use. Defaults to 1Gi.")
	f.IntVar(&p.Pservers, "pservers", 0, "Number of parameter servers. Defaults equal to -p")
	f.IntVar(&p.PSCPU, "pscpu", 1, "Parameter server CPU resource. Defaults to 1.")
	f.StringVar(&p.PSMemory, "psmemory", "1Gi", "Parameter server momory resource. Defaults to 1Gi.")
	f.StringVar(&p.Entry, "entry", "", "Command of starting trainer process. Defaults to paddle train")
	f.StringVar(&p.Topology, "topology", "", "Will Be Deprecated .py file contains paddle v1 job configs")
	f.IntVar(&p.Passes, "passes", 1, "Pass count for training job")
}

// Execute submit command.
func (p *SubmitCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return subcommands.ExitFailure
	}
	// default pservers count equals to trainers count.
	if p.Pservers == 0 {
		p.Pservers = p.Parallelism
	}
	p.Jobpackage = f.Arg(0)
	p.Datacenter = utils.Config.ActiveConfig.Name

	s := NewSubmitter(p)
	errS := s.Submit(f.Arg(0), p.Jobname)
	if errS != nil {
		fmt.Fprintf(os.Stderr, "error submiting job: %v\n", errS)
		return subcommands.ExitFailure
	}
	fmt.Printf("%s submited.\n", p.Jobname)
	return subcommands.ExitSuccess
}

// Submitter submit job to cloud.
type Submitter struct {
	args *SubmitCmd
}

// NewSubmitter returns a submitter object.
func NewSubmitter(cmd *SubmitCmd) *Submitter {
	s := Submitter{cmd}
	return &s
}

// Submit current job.
func (s *Submitter) Submit(jobPackage string, jobName string) error {
	// 1. upload user job package to pfs
	err := filepath.Walk(jobPackage, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		glog.V(10).Infof("Uploading %s...\n", filePath)
		dest := path.Join("/pfs", utils.Config.ActiveConfig.Name, "home", utils.Config.ActiveConfig.Username, "jobs", jobName, filepath.Base(filePath))
		fmt.Printf("uploading: %s...\n", filePath)
		return putFile(filePath, dest)
	})
	if err != nil {
		return err
	}
	// 2. call paddlecloud server to create kubernetes job
	jsonString, err := json.Marshal(s.args)
	if err != nil {
		return err
	}
	glog.V(10).Infof("Submitting job: %s to %s\n", jsonString, utils.Config.ActiveConfig.Endpoint+"/api/v1/jobs")
	respBody, err := utils.PostCall(utils.Config.ActiveConfig.Endpoint+"/api/v1/jobs/", jsonString)
	if err != nil {
		return err
	}
	var respObj interface{}
	if err = json.Unmarshal(respBody, &respObj); err != nil {
		return err
	}
	// FIXME: Return an error if error message is not empty. Use response code instead
	errMsg := respObj.(map[string]interface{})["msg"].(string)
	if len(errMsg) > 0 {
		return errors.New(errMsg)
	}
	return nil
}
