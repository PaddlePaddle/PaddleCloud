// trainingjober is a package for managing a TrainingJob
package trainingjober

import (
	"time"
	"github.com/PaddlePaddle/cloud/go/pkg/apis/paddlepaddle/v1alpha1"
	trainingJobClient "github.com/PaddlePaddle/cloud/go/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
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
	job *v1alpha1.TrainingJob
}

// TrainingJober is to manager a specific TrainingJob
type TrainingJober struct {
	// job is the job the TrainingJober manager.
	job *v1alpha1.TrainingJob

	// kubeCli is standard kubernetes client.
	kubeCli kubernetes.Interface

	// trainingJobClient is the client of TrainingJob.
	trainingJobClient trainingJobClient.Interface

	// status is the status in memory, update when TrainingJob status changed and update the CRD resource status.
	status v1alpha1.TrainingJobStatus

	// eventCh is the channel received by Controller, include Modify and Delete.
	// When trainingJobEvent is Delete it will delete all resources
	// The maximum is 1000.
	eventCh chan *trainingJobEvent
}

// initJobber init a TrainingJober to manager a specific training job.
func initJobber(job *v1alpha1.TrainingJob, kubeCli kubernetes.Interface, trainingJobClient trainingJobClient.Interface) (*TrainingJober,
error) {
	jobber := &TrainingJober{
		job:               job,
		kubeCli:           kubeCli,
		trainingJobClient: trainingJobClient,
		status:            job.Status,
		eventCh:           make(chan *trainingJobEvent, 1000),
	}
	return jobber, nil
}

