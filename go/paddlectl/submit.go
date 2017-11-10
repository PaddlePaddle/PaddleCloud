package paddlectl

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/pkg/api/v1"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	"github.com/PaddlePaddle/cloud/go/utils/config"
	kubeutil "github.com/PaddlePaddle/cloud/go/utils/kubeutil"
	"github.com/golang/glog"
	"github.com/google/subcommands"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	invalidJobName = "jobname can not contain '.' or '_'"
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

	// TODO(gongwb): init config in memory.
	KubeConfig string `json:"kubeconfig"`

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

func (p *SubmitCmd) getTrainer() *paddlejob.TrainerSpec {
	return &paddlejob.TrainerSpec{
		Entrypoint: p.Entry,
		// FIXME(gongwb): workspace

		MinInstance: p.MinInstance,
		MaxInstance: p.MaxInstance,
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				"cpu":    *apiresource.NewQuantity(int64(p.CPU), apiresource.DecimalSI),
				"memory": apiresource.MustParse(p.Memory),
			},
			Requests: v1.ResourceList{
				"cpu":    *apiresource.NewQuantity(int64(p.CPU), apiresource.DecimalSI),
				"memory": apiresource.MustParse(p.Memory),
			},
		},
	}
}

func (p *SubmitCmd) getPserver() *paddlejob.PserverSpec {
	return &paddlejob.PserverSpec{
		// TODO(gongwb):Pserver can be auto-scaled?
		MinInstance: p.Pservers,
		MaxInstance: p.Pservers,
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				"cpu":    *apiresource.NewQuantity(int64(p.PSCPU), apiresource.DecimalSI),
				"memory": apiresource.MustParse(p.PSMemory),
			},
			Requests: v1.ResourceList{
				"cpu":    *apiresource.NewQuantity(int64(p.PSCPU), apiresource.DecimalSI),
				"memory": apiresource.MustParse(p.PSMemory),
			},
		},
	}
}

func (p *SubmitCmd) getMaster() *paddlejob.MasterSpec {
	return &paddlejob.MasterSpec{}
}

// GetTrainingJob get's paddlejob.TrainingJob struct filed by Submitcmd paramters.
func (p *SubmitCmd) GetTrainingJob() *paddlejob.TrainingJob {

	t := paddlejob.TrainingJob{
		metav1.TypeMeta{
			Kind:       "TrainingJob",
			APIVersion: "paddlepaddle.org/v1",
		},
		metav1.ObjectMeta{
			Name:      p.Jobname,
			Namespace: kubeutil.NameEscape(Config.ActiveConfig.Username),
		},

		// General job attributes.
		paddlejob.TrainingJobSpec{
			Image: p.Image,

			// TODO(gongwb): init them?

			FaultTolerant: p.FaultTolerant,
			Passes:        p.Passes,

			Trainer: *p.getTrainer(),
			Pserver: *p.getPserver(),
			Master:  *p.getMaster(),
		},
		paddlejob.TrainingJobStatus{},
	}

	if glog.V(3) {
		glog.Infof("GetTrainingJob: %s\n", t)
	}
	return &t
}

// SetFlags registers subcommands flags.
func (p *SubmitCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.KubeConfig, "kubeconfig", "", "Kubernetes config.")
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

// putFiles puts files to pfs and
// if jobPackage is not a local dir, skip uploading package.
func putFiles(jobPackage, jobName string) error {
	_, pkgerr := os.Stat(jobPackage)
	if pkgerr == nil {
		// FIXME: upload job package to paddle cloud.
	} else if os.IsNotExist(pkgerr) {
		return fmt.Errorf("stat jobpackage '%s' error: %v", jobPackage, pkgerr)
	}

	return nil
}

func (s *Submitter) getKubeConfig() (string, error) {
	kubeconfig := s.args.KubeConfig
	if _, err := os.Stat(kubeconfig); err != nil {
		return "", fmt.Errorf("can't access kubeconfig '%s' error: %v", kubeconfig, err)
	}

	return kubeconfig, nil
}

// Submit current job.
func (s *Submitter) Submit(jobPackage string, jobName string) error {
	if err := checkJobName(jobName); err != nil {
		return err
	}

	if err := putFiles(jobPackage, jobName); err != nil {
		return err
	}

	kubeconfig, err := s.getKubeConfig()
	if err != nil {
		return err
	}

	client, clientset, err := kubeutil.CreateClient(kubeconfig)
	if err != nil {
		return err
	}

	namespace := kubeutil.NameEscape(Config.ActiveConfig.Username)
	if err := kubeutil.FindNamespace(clientset, namespace); err != nil {
		return err
	}

	if err := kubeutil.CreateTrainingJob(client, namespace, s.args.GetTrainingJob()); err != nil {
		return err
	}

	return nil
}
func checkJobName(jobName string) error {
	if strings.Contains(jobName, "_") || strings.Contains(jobName, ".") {
		return errors.New(invalidJobName)
	}

	return nil
}
