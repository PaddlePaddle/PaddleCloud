package paddlecloud

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	"github.com/PaddlePaddle/cloud/go/utils/config"
	"github.com/PaddlePaddle/cloud/go/utils/restclient"
	"github.com/golang/glog"
	"github.com/google/subcommands"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

const (
	invalidJobName = "jobname can not contain '.' or '_'"
)

// Config is global config object for paddlecloud commandline
var Config = config.ParseDefaultConfig()

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
	// docker image to run jobs
	Image    string `json:"image"`
	Registry string `json:"registry"`
	// Alpha features:
	// TODO: separate API versions
	FaultTolerant bool `json:"faulttolerant"`
	MinInstance   int  `json:"min-instance"`
	MaxInstance   int  `json:"max-instance"`
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
	f.StringVar(&p.Image, "image", "", "Runtime Docker image for the job")
	f.StringVar(&p.Registry, "registry", "", "Registry secret name for the runtime Docker image")
	f.IntVar(&p.MinInstance, "min-instance", 1, "The minimum instance number for the trainer.")
	f.IntVar(&p.MaxInstance, "max-instance", 1, "The maximum instance number for the trainer.")
	f.BoolVar(&p.FaultTolerant, "faulttolerant", false, "if true, use new fault-tolerant pservers")
}

// ParseToTrainingJob parse command-line parameters to TrainingJob
func (p *SubmitCmd) ParseToTrainingJob() *paddlejob.TrainingJob {
	job := paddlejob.TrainingJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrainingJob",
			APIVersion: "paddlepaddle.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Name(),
		},
		Spec: paddlejob.TrainingJobSpec{
			Image:             p.Image,
			Port:              7164,
			PortsNum:          1,
			PortsNumForSparse: 1,
			FaultTolerant:     p.FaultTolerant,
			Passes:            p.Passes,
			Trainer: paddlejob.TrainerSpec{
				Entrypoint:  p.Entry,
				MinInstance: p.MinInstance,
				MaxInstance: p.MaxInstance,
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:       resource.MustParse(string(p.CPU)),
						v1.ResourceMemory:    resource.MustParse(p.Memory),
						v1.ResourceNvidiaGPU: resource.MustParse(string(p.GPU)),
					},
					Requests: v1.ResourceList{
						v1.ResourceCPU:       resource.MustParse(string(p.CPU)),
						v1.ResourceMemory:    resource.MustParse(p.Memory),
						v1.ResourceNvidiaGPU: resource.MustParse(string(p.GPU)),
					},
				},
			},
			Pserver: paddlejob.PserverSpec{
				MinInstance: p.Pservers,
				MaxInstance: p.Pservers,
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse(string(p.PSCPU)),
						v1.ResourceMemory: resource.MustParse(p.PSMemory),
					},
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse(string(p.PSCPU)),
						v1.ResourceMemory: resource.MustParse(p.PSMemory),
					},
				},
			},
		},
	}
	if p.FaultTolerant {
		job.Spec.Master = paddlejob.MasterSpec{
			Resources: v1.ResourceRequirements{
				// TODO(Yancey1989) caculate resource according trainingjob resource
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("1"),
					v1.ResourceMemory: resource.MustParse("2Gi"),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("1"),
					v1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
		}
	}
	return &job
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
	p.Datacenter = Config.ActiveConfig.Name

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
	if err := checkJobName(jobName); err != nil {
		return err
	}
	// if jobPackage is not a local dir, skip uploading package.
	_, pkgerr := os.Stat(jobPackage)
	if pkgerr == nil {
		dest := path.Join("/pfs", Config.ActiveConfig.Name, "home", Config.ActiveConfig.Username, "jobs", jobName)
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
	// 2. call paddlecloud server to create kubernetes job
	jsonString, err := json.Marshal(s.args.ParseToTrainingJob())
	if err != nil {
		return err
	}
	glog.V(10).Infof("Submitting job: %s to %s\n", jsonString, Config.ActiveConfig.Endpoint+"/api/v1/jobs")
	respBody, err := restclient.PostCall(Config.ActiveConfig.Endpoint+"/api/v1/jobs/", jsonString)
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
func checkJobName(jobName string) error {
	if strings.Contains(jobName, "_") || strings.Contains(jobName, ".") {
		return errors.New(invalidJobName)
	}
	return nil
}
