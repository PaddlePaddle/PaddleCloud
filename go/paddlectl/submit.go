package paddlectl

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	"github.com/PaddlePaddle/cloud/go/utils/config"
	"github.com/golang/glog"
	"github.com/google/subcommands"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	MaxInstance   int  `json:"maxInstance"`
	MinInstance   int  `json:"minInstance"`

	// TODO: init config in memory.
	//KubeConfig string `json:"kubeconfig"`
	//JobYaml    string `json:"jobyaml"`
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

func (p *SubmitCmd) getTrainer() *paddlejob.TrainerSpec {
	return &paddlejob.TrainerSpec{
		Entrypoint: p.Entry,
		// FIXME:(gongwb) workspace
		//Workspace:   p.Workspace,
		MinInstance: p.MinInstance,
		MaxInstance: p.MaxInstance,
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				"cpu":    strconv.Itoa(p.CPU),
				"memory": p.Memory,
			},
			Requests: v1.ResourceList{
				"cpu":    strconv.Itoa(p.CPU),
				"memory": p.Memory,
			},
		},
	}
}

func (p *SubmitCmd) getPserver() *paddlejob.PserverSpec {
	return &paddlejob.PserverSpec{
		// TODO:Pserver can be auto-scaled?
		MinInstance: p.Pservers,
		MaxInstance: p.Pservers,
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				"cpu":    strconv.Itoa(p.PSCPU),
				"memory": p.PSMemory,
			},
			Requests: v1.ResourceList{
				"cpu":    strconv.Itoa(p.PSCPU),
				"memory": p.PSMemory,
			},
		},
	}
}

func (p *SubmitCmd) getMaster() *paddlejob.MasterSpec {
	return &paddlejob.MasterSpec{}
}

// GetTrainingJob get's paddlejob.TrainingJob struct filed by Submitcmd paramters.
func (p *SubmitCmd) GetTrainingJob() paddlejob.TrainingJob {
	return paddlejob.TrainingJob{
		metav1.TypeMeta{
			Kind:       "TrainingJob",
			APIVersion: "paddlepaddle.org/v1",
		},
		metav1.ObjectMeta{
			Name:      p.Jobname,
			Namespace: Config.ActiveConfig.Username,
		},
		// General job attributes.
		paddlejob.TrainingJobSpec{
			Image: p.Image,

			// TODO: init them
			//Port:              p.Port,
			//PortsNum:          p.PortNum,
			//PortsNumForSparse: p.PortsNumForSparse,

			FaultTolerant: p.FaultTolerant,
			Passes:        p.Passes,
			// Job components.
			Trainer: p.getTrainer(),
			Pserver: p.getPserver(),
			Master:  p.getMaster(),
		},
		paddlejob.TrainingJobStatus{},
	}
}

// SetFlags registers subcommands flags.
func (p *SubmitCmd) SetFlags(f *flag.FlagSet) {
	//f.StringVar(&p.KubeConfig, "kubeconfig", "", "Kubernetes config.")
	//f.StringVar(&p.yaml, "yaml", "", "Job's yaml.")
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

func putFiles(jobPackage string, dest string) error {
	return nil
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
		// FIXME: upload job package to paddle cloud.
		err := putFiles(jobPackage, dest)
		if err != nil {
			return err
		}
	} else if os.IsNotExist(pkgerr) {
		glog.Warning("jobpackage not a local dir, skip upload.")
	}

	kubeconfig := "/Users/gongwb/.kube/config"
	resource := "training-job.paddlepaddle.org"
	namespace := "gongweibao-baidu-com"
	apiversion := "v1"
	client, clientset := createClient(kubeconfig)
	//ensureNamespace(clientset, namespace)
	ensureTPR(clientset, resource, namespace, apiversion)
	createDemo(client)
	/*
		fmt.Println("kubeconfig:%v resource:%v, namespace:%v apiversion:%v",
			kubeconfig, resource, namespace, apiversion)
	*/

	/*
		job := paddlejob.TrainingJob{}
		// TODO:unmarshel yaml and check job.
		if err := parser.Validate(&job); err != nil {
			fmt.Printf("valid arguments error: %v\n", err)
			return err
		}

		// TODO: parse to corresponding yaml and create TPR with kubernetes API.
		var parser controller.DefaultJobParser
		master := parser.ParseToMaster(&job)
		if err != nil {
			fmt.Printf("create error: %v\n", err)
			return err
		}

		fmt.Printf("result:%v err:%v\n", result, err)
	*/
	return nil
}
func checkJobName(jobName string) error {
	if strings.Contains(jobName, "_") || strings.Contains(jobName, ".") {
		return errors.New(invalidJobName)
	}

	return nil
}

func checkJob(nameSpace, jobName string, client *rest.RESTClient) error {
	/*
		type DemoSpec struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		type Demo struct {
			unversioned.TypeMeta `json:",inline"`
			api.ObjectMeta       `json:"metadata,omitempty"`

			Spec DemoSpec `json:"spec"`
		}

		demo := Demo{
			Spec: DemoSpec{
				Name:        jobName,
				Description: "Description for " + jobName + ".",
			},
		}
		// TODO: check jobName not exists.
		err := client.Get().
			Resource("TrainJob").
			Namespace(nameSpace).
			Name(jobName).
			Do().Into(demo)
	*/
	return nil

}
