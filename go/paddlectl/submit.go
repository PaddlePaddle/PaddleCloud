package paddlectl

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/PaddlePaddle/cloud/go/utils/config"
	"github.com/PaddlePaddle/cloud/go/utils/restclient"
	"github.com/golang/glog"
	"github.com/google/subcommands"
)

const (
	invalidJobName   = "jobname can not contain '.' or '_'"
	trainingjobsPath = "/api/v1/trainingjobs"
)

// Config is global config object for paddlectl commandline
var Config = config.ParseDefaultConfig()

// SubmitCmd define the subcommand of submitting paddle training jobs.
type SubmitCmd struct {
	Jobname    string `json:"name"`
	Jobpackage string `json:"jobPackage"`
	CPU        int    `json:"cpu"`
	GPU        int    `json:"gpu"`
	Memory     string `json:"memory"`
	Pservers   int    `json:"pservers"`
	PSCPU      int    `json:"pscpu"`
	PSMemory   string `json:"psmemory"`
	Entry      string `json:"entry"`
	Topology   string `json:"topology"`
	Datacenter string `json:"datacenter"`
	Passes     int    `json:"passes"`

	// docker image to run jobs
	Image    string `json:"image"`
	Registry string `json:"registry"`

	// Alpha features:
	// TODO: separate API versions
	FaultTolerant bool `json:"faulttolerant"`
	MaxInstance   int  `json:"maxInstance"`
	MinInstance   int  `json:"minInstance"`

	// TODO(gongwb): create from yaml.
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
	f.IntVar(&p.CPU, "cpu", 1, "CPU resource each trainer will use. Defaults to 1.")
	f.IntVar(&p.GPU, "gpu", 0, "GPU resource each trainer will use. Defaults to 0.")
	f.StringVar(&p.Memory, "memory", "1Gi", " Memory resource each trainer will use. Defaults to 1Gi.")
	f.IntVar(&p.Pservers, "pservers", 1, "Number of parameter servers. Defaults equal to -p")
	f.IntVar(&p.PSCPU, "pscpu", 1, "Parameter server CPU resource. Defaults to 1.")
	f.StringVar(&p.PSMemory, "psmemory", "1Gi", "Parameter server momory resource. Defaults to 1Gi.")
	f.StringVar(&p.Entry, "entry", "", "Command of starting trainer process. Defaults to paddle train")
	f.StringVar(&p.Topology, "topology", "", "Will Be Deprecated .py file contains paddle v1 job configs")
	f.IntVar(&p.Passes, "passes", 1, "Pass count for training job")
	f.StringVar(&p.Image, "image", "paddlepaddle/paddlecloud-job", "Runtime Docker image for the job")
	f.StringVar(&p.Registry, "registry", "", "Registry secret name for the runtime Docker image")
	f.IntVar(&p.MinInstance, "min_instance", 1, "The minimum number of trainers"+
		"only used fo faulttolerant. Default to 1.")
	f.IntVar(&p.MaxInstance, "max_instance", 1, "The minimum number of trainers,"+
		"only used for faulttolerant, Default to 1.")
	f.BoolVar(&p.FaultTolerant, "faulttolerant", false, "if true, use new fault-tolerant pservers")
}

// Execute submit command.
func (p *SubmitCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	p.Jobpackage = f.Arg(0)
	p.Datacenter = Config.ActiveConfig.Name

	s := NewSubmitter(p)
	if err := s.Submit(f.Arg(0), p.Jobname); err != nil {
		fmt.Fprintf(os.Stderr, "error submiting job: %v\n", err)
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

func getJobPfsPath(jobPackage, jobName string) string {
	_, err := os.Stat(jobPackage)
	if os.IsNotExist(err) {
		return jobPackage
	}

	return path.Join("/pfs", Config.ActiveConfig.Name, "home", Config.ActiveConfig.Username, "jobs", jobName)
}

// putFiles puts files to pfs and
// if jobPackage is not a local dir, skip uploading package.
func putFilesToPfs(jobPackage, jobName string) error {
	_, pkgerr := os.Stat(jobPackage)
	if pkgerr == nil {
		dest := getJobPfsPath(jobPackage, jobName)
		if !strings.HasSuffix(jobPackage, "/") {
			jobPackage = jobPackage + "/"
		}
		err := putFiles(jobPackage, dest)
		if err != nil {
			return err
		}
	} else if os.IsNotExist(pkgerr) {
		glog.Warning("jobpackage not a local dir, skip upload.")
	}

	return nil
}

func (s *Submitter) createJobs() error {
	jsonString, err := json.Marshal(s.args)
	if err != nil {
		return err
	}

	apiPath := Config.ActiveConfig.Endpoint + trainingjobsPath
	respBody, err := restclient.PostCall(apiPath, jsonString)
	if err != nil {
		return err
	}
	var respObj interface{}
	if err = json.Unmarshal(respBody, &respObj); err != nil {
		return err
	}

	// FIXME: Return an error if error message is not empty. Use response code instead.
	errMsg := respObj.(map[string]interface{})["msg"].(string)
	if len(errMsg) > 0 {
		return errors.New(errMsg)
	}

	glog.Infof("Submitting job: %s\n", s.args.Jobname)
	return nil
}

// Submit current job.
func (s *Submitter) Submit(jobPackage string, jobName string) error {
	if err := checkJobName(jobName); err != nil {
		return err
	}

	if err := putFilesToPfs(jobPackage, jobName); err != nil {
		return err
	}

	// 2. call paddlecloud server to create TPR TraningJobs.
	return s.createJobs()
}

func checkJobName(jobName string) error {
	if strings.Contains(jobName, "_") || strings.Contains(jobName, ".") {
		return errors.New(invalidJobName)
	}

	return nil
}
