package updater

import (
	padv1 "github.com/PaddlePaddle/cloud/go/apis/paddlepaddle/v1"
	trainingJobClient "github.com/PaddlePaddle/cloud/go/client/clientset/versioned"
	log "github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	retryTimes            = 5
	convertedTimerTicker  = 10 * time.Second
	confirmResourceTicker = 5 * time.Second
)

type trainingJobEventType string

const (
	trainingJobEventDelete trainingJobEventType = "Delete"
	trainingJobEventModify trainingJobEventType = "Modify"
)

type trainingJobEvent struct {
	// pet is the TrainingJobEventType of TrainingJob
	pet trainingJobEventType
	// The job transfer the information fo job
	job *padv1.TrainingJob
}

// TrainingJobUpdater is to manager a specific TrainingJob
type TrainingJobUpdater struct {
	// Job is the job the TrainingJob manager.
	job *padv1.TrainingJob

	// kubeCli is standard kubernetes client.
	kubeCli kubernetes.Interface

	// TrainingJobClient is the client of TrainingJob.
	trainingJobClient trainingJobClient.Interface

	// Status is the status in memory, update when TrainingJob status changed and update the CRD resource status.
	status padv1.TrainingJobStatus

	// EventCh is the channel received by Controller, include Modify and Delete.
	// When trainingJobEvent is Delete it will delete all resources
	// The maximum is 1000.
	eventCh chan *trainingJobEvent
}

func initUpdater(job *padv1.TrainingJob, kubeCli kubernetes.Interface, trainingJobClient trainingJobClient.Interface) (*TrainingJobUpdater,
	error) {
	jobber := &TrainingJobUpdater{
		job:               job,
		kubeCli:           kubeCli,
		trainingJobClient: trainingJobClient,
		status:            job.Status,
		eventCh:           make(chan *trainingJobEvent, 1000),
	}
	return jobber, nil
}

// NewUpdater return a trainingJobUpdater for controller.
func NewUpdater(job *padv1.TrainingJob, kubeCli kubernetes.Interface, trainingJobClient trainingJobClient.Interface) (*TrainingJobUpdater,
	error) {
	log.Infof("NewJobber namespace=%v name=%v", job.Namespace, job.Name)
	jobber, err := initUpdater(job, kubeCli, trainingJobClient)
	if err != nil {
		return nil, err
	}

	go jobber.start()
	return jobber, nil
}

// Start is the main process of life cycle of a TrainingJob, including create resources, event process handle and
// status convert.
func (updater *TrainingJobUpdater) start() {
	//	TODO(zhengqi): this will commit in the next pr

}
