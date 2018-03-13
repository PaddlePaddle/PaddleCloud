package updater

import (
	"fmt"
	padv1 "github.com/PaddlePaddle/cloud/go/apis/paddlepaddle/v1"
	trainingJobClient "github.com/PaddlePaddle/cloud/go/client/clientset/versioned"
	log "github.com/golang/glog"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"time"
)

const (
	retryTimes            = 5
	convertedTimerTicker  = 10 * time.Second
	confirmResourceTicker = 5 * time.Second
	eventChLength         = 1000
	factor                = 0.8
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

// TrainingJobUpdater is used to manage a specific TrainingJob
type TrainingJobUpdater struct {
	// Job is the job the TrainingJob manager.
	job *padv1.TrainingJob

	// kubeClient is standard kubernetes client.
	kubeClient kubernetes.Interface

	// TrainingJobClient is the client of TrainingJob.
	trainingJobClient trainingJobClient.Interface

	// Status is the status in memory, update when TrainingJob status changed and update the CRD resource status.
	status padv1.TrainingJobStatus

	// EventCh receives events from the controller, include Modify and Delete.
	// When trainingJobEvent is Delete it will delete all resources
	// The capacity is 1000.
	eventCh chan *trainingJobEvent
}

// NewUpdater creates a new TrainingJobUpdater and start a goroutine to control current job.
func NewUpdater(job *padv1.TrainingJob, kubeClient kubernetes.Interface, trainingJobClient trainingJobClient.Interface) (*TrainingJobUpdater,
	error) {
	log.Infof("NewJobber namespace=%v name=%v", job.Namespace, job.Name)
	updater := &TrainingJobUpdater{
		job:               job,
		kubeClient:        kubeClient,
		trainingJobClient: trainingJobClient,
		status:            job.Status,
		eventCh:           make(chan *trainingJobEvent, eventChLength),
	}
	go updater.start()
	return updater, nil
}

// Notify is used to receive event from controller. While controller receive a informer,
// it will notify updater to process the event. It send event to updater's eventCh.
func (updater *TrainingJobUpdater) notify(te *trainingJobEvent) {
	updater.eventCh <- te
	lene, cape := len(updater.eventCh), cap(updater.eventCh)
	if lene > int(float64(cape)*factor) {
		log.Warning("the len of updater eventCh ", updater.job.Name, " is near to full")
	}
}

// Delete send a delete event to updater, updater will kill the trainingjob and clear all the resource of the
// trainingjob.
func (updater *TrainingJobUpdater) Delete() {
	updater.notify(&trainingJobEvent{pet: trainingJobEventDelete})
}

// Modify send a modify event to updater. updater will processing according to the situation.
func (updater *TrainingJobUpdater) Modify(nj *padv1.TrainingJob) {
	updater.notify(&trainingJobEvent{pet: trainingJobEventModify, job: nj})
}

func (updater *TrainingJobUpdater) releaseResource(tp padv1.TrainingResourceType) error {
	resource := new(v1beta1.ReplicaSet)
	switch tp {
	case padv1.MASTER:
		resource = updater.job.Spec.Master.ReplicaSpec
	case padv1.PSERVER:
		resource = updater.job.Spec.Pserver.ReplicaSpec
	default:
		return fmt.Errorf("unknow resource")
	}
	var replica int32
	resource.Spec.Replicas = &replica
	_, err := updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Update(resource)
	if errors.IsNotFound(err) {
		return nil
	}
	key := "paddle-job-" + tp

	labels := Labels(map[string]string{
		string(key): updater.job.Name,
	})

	selector, _ := labels.LabelsParser()
	options := v1.ListOptions{
		LabelSelector: selector,
	}

	for j := 0; j <= retryTimes; j++ {
		time.Sleep(confirmResourceTicker)
		pl, err := updater.kubeClient.CoreV1().Pods(updater.job.Namespace).List(options)
		if err == nil && len(pl.Items) == 0 {
			return nil
		}
	}
	err = updater.kubeClient.CoreV1().Pods(updater.job.Namespace).DeleteCollection(&v1.DeleteOptions{}, options)
	return err
}

func (updater *TrainingJobUpdater) releaseMaster() error {
	return updater.releaseResource(padv1.MASTER)
}

func (updater *TrainingJobUpdater) releasePserver() error {
	return updater.releaseResource(padv1.PSERVER)
}

func (updater *TrainingJobUpdater) releaseTrainer() error {
	labels := Labels(map[string]string{
		"paddle-job": updater.job.Name,
	})
	selector, _ := labels.LabelsParser()
	options := v1.ListOptions{
		LabelSelector: selector,
	}

	return updater.kubeClient.CoreV1().Pods(updater.job.Namespace).DeleteCollection(&v1.DeleteOptions{}, options)
}

func (updater *TrainingJobUpdater) deleteTrainingJob() error {
	fault := false

	log.Infof("Start to delete TrainingJob namespace=%v name=%v", updater.job.Namespace, updater.job.Name)

	if updater.job.Spec.FaultTolerant {
		log.Infof("Release master, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
		if err := updater.releaseMaster(); err != nil {
			log.Error(err.Error())
			fault = true
		}
	}

	log.Infof("Release pserver, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
	if err := updater.releasePserver(); err != nil {
		log.Error(err.Error())
		fault = true
	}

	if updater.job.Spec.FaultTolerant {
		log.Infof("Deleting TrainingJob matadata, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Master.ReplicaSpec.Name)
		if err := updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Delete(updater.job.Spec.Master.ReplicaSpec.Name, &v1.DeleteOptions{}); err != nil {
			log.Error("delete master replicaset error: ", err.Error())
			fault = true
		}
	}

	log.Infof("Deleting TrainingJob matadata, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Pserver.ReplicaSpec.Name)
	if err := updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Delete(updater.job.Spec.Pserver.ReplicaSpec.Name, &v1.DeleteOptions{}); err != nil {
		log.Error("delete pserver replicaset error: ", err.Error())
		fault = true
	}

	log.Infof("Deleting TrainingJob matadata, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
	if err := updater.kubeClient.BatchV1().Jobs(updater.job.Namespace).Delete(updater.job.Spec.Trainer.ReplicaSpec.Name, &v1.DeleteOptions{}); err != nil {
		log.Error("delete trainer replicaset error: ", err.Error())
		fault = true
	}

	log.Infof("Release trainer, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
	if err := updater.releaseTrainer(); err != nil {
		log.Error("release trainer  error: ", err.Error())
		fault = true
	}

	log.Infof("End to delete TrainingJob namespace=%v name=%v", updater.job.Namespace, updater.job.Name)

	if fault {
		return fmt.Errorf("delete resource error namespace=%v name=%v", updater.job.Namespace, updater.job.Name)
	}
	return nil
}

func (updater *TrainingJobUpdater) createResource(tp padv1.TrainingResourceType) error {
	resource := new(v1beta1.ReplicaSet)
	switch tp {
	case padv1.MASTER:
		resource = updater.job.Spec.Master.ReplicaSpec
	case padv1.PSERVER:
		resource = updater.job.Spec.Pserver.ReplicaSpec
	default:
		return fmt.Errorf("unknow resource")
	}
	newResource, err := updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Get(resource.Name, v1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Infof("not found to create namespace=%v name=%v resourceName=%v", updater.job.Namespace, updater.job.Name, resource.Name)
		newResource, err = updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Create(resource)
		if err != nil {
			updater.status.Phase = padv1.TrainingJobPhaseFailed
			updater.status.Reason = "Internal error; create resource error:" + err.Error()
			return err
		}
	}
	if newResource != nil {
		ticker := time.NewTicker(confirmResourceTicker)
		defer ticker.Stop()
		for v := range ticker.C {
			rs, err := updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Get(resource.Name, v1.GetOptions{})
			log.Infof("Current time %v runing pod is %v, resourceName=%v", v.String(), rs.Status.ReadyReplicas, resource.Name)
			if err != nil && !errors.IsServerTimeout(err) && !errors.IsTooManyRequests(err) {
				updater.status.Reason = "Internal error; create resource error:" + err.Error()
				return err
			}
			if errors.IsServerTimeout(err) || errors.IsTooManyRequests(err) {
				log.Warningf("Connect to kubernetes failed for reasons=%v, retry next ticker", err.Error())
				continue
			}
			if *resource.Spec.Replicas == 0 {
				return fmt.Errorf(" trainingjob is deleting, namespace=%v name=%v ", updater.job.Namespace, updater.job.Name)

			}
			if rs.Status.ReadyReplicas == *resource.Spec.Replicas {
				log.Infof("Create resource done , namespace=%v name=%v resourceName=%v", updater.job.Namespace, updater.job.Name, resource.Name)
				return nil
			}
		}
	}
	return nil
}

func (updater *TrainingJobUpdater) createTrainer() error {
	newTrainer, err := updater.kubeClient.BatchV1().Jobs(updater.job.Namespace).Get(updater.job.Spec.Trainer.ReplicaSpec.Name, v1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Infof("not found to trainer pserver namespace=%v name=%v", updater.job.Namespace, updater.job.Name)
		newTrainer, err = updater.kubeClient.BatchV1().Jobs(updater.job.Namespace).Create(updater.job.Spec.Trainer.ReplicaSpec)
		if err != nil {
			updater.status.Phase = padv1.TrainingJobPhaseFailed
			updater.status.Reason = "Internal error; create trainer error:" + err.Error()
			return err
		}
	}
	if newTrainer != nil {
		updater.status.Phase = padv1.TrainingJobPhaseRunning
		updater.status.Reason = ""
		return nil
	}
	return nil
}

func (updater *TrainingJobUpdater) createTrainingJob() error {
	if updater.job.Spec.FaultTolerant {

		if err := updater.createResource(padv1.MASTER); err != nil {
			return err
		}
	}
	if err := updater.createResource(padv1.PSERVER); err != nil {
		return err
	}
	return updater.createTrainer()
}

func (updater *TrainingJobUpdater) updateCRDStatus() error {
	if reflect.DeepEqual(updater.status, updater.job.Status) {
		return nil
	}
	newTrainingJob := updater.job
	newTrainingJob.Status = updater.status
	newTrainingJob, err := updater.trainingJobClient.PaddlepaddleV1().TrainingJobs(updater.job.Namespace).Update(newTrainingJob)
	if err != nil {
		return err
	}
	updater.job = newTrainingJob
	return nil
}

// parseTrainingJob validates the fields and parses the TrainingJob
func (updater *TrainingJobUpdater) parseTrainingJob() {
	if updater.job == nil {
		updater.status.Phase = padv1.TrainingJobPhaseFailed
		updater.status.Reason = "Internal error; Setup error; job is missing TainingJob"
		return
	}

	err := func() error {
		// TODO(Zhengqi): Parse TrainingJob, this will be submitted in the next pr
		return nil
	}()

	if err != nil {
		updater.status.Phase = padv1.TrainingJobPhaseFailed
		updater.status.Reason = err.Error()
	} else {
		updater.status.Phase = padv1.TrainingJobPhaseCreating
		updater.status.Reason = ""
	}
}

func (updater *TrainingJobUpdater) getTrainerReplicaStatuses() ([]*padv1.TrainingResourceStatus, error) {
	var replicaStatuses []*padv1.TrainingResourceStatus
	trs := padv1.TrainingResourceStatus{
		TrainingResourceType: padv1.TRAINER,
		State:                padv1.ResourceStateNone,
		ResourceStates:       make(map[padv1.ResourceState]int),
	}
	// TODO(ZhengQi): get detail status in future
	replicaStatuses = append(replicaStatuses, &trs)
	return replicaStatuses, nil
}

// GetStatus get TrainingJob status from trainers.
func (updater *TrainingJobUpdater) GetStatus() (*padv1.TrainingJobStatus, error) {

	status := updater.status

	j, err := updater.kubeClient.BatchV1().Jobs(updater.job.Namespace).
		Get(updater.job.Spec.Trainer.ReplicaSpec.Name, v1.GetOptions{})
	if err != nil {
		log.Error("get trainer error:", err.Error())
		return &status, err
	}

	status.ReplicaStatuses, err = updater.getTrainerReplicaStatuses()
	if err != nil {
		log.Error("get trainer replica status error:", err.Error())
	}

	if updater.job.Spec.FaultTolerant {
		// TODO(ZhengQi): should to confirm when job done
		if j.Status.Failed == *updater.job.Spec.Trainer.ReplicaSpec.Spec.Parallelism {
			status.Phase = padv1.TrainingJobPhaseFailed
			status.Reason = "all trainer have failed!"
		} else {
			if j.Status.Succeeded != 0 && j.Status.Active == 0 {
				status.Phase = padv1.TrainingJobPhaseSucceeded
				status.Reason = "Success!"
			}
		}
	} else {
		if j.Status.Failed != 0 {
			status.Phase = padv1.TrainingJobPhaseFailed
			status.Reason = "at least one trainer failed!"
		} else {
			if j.Status.Succeeded == *updater.job.Spec.Trainer.ReplicaSpec.Spec.Parallelism && j.Status.Active == 0 {
				status.Phase = padv1.TrainingJobPhaseSucceeded
				status.Reason = "all trainer have succeeded!"
			}
		}
	}
	return &status, nil
}

// Convert is main process to convert TrainingJob to desire status.
func (updater *TrainingJobUpdater) Convert() {
	log.Infof("convert status, namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)

	if updater.status.Phase == padv1.TrainingJobPhaseRunning {
		status, err := updater.GetStatus()
		if err != nil {
			log.Error("get current status of trainer from k8s error:", err.Error())
			return
		}
		updater.status = *status.DeepCopy()
		log.Infof("Current status namespace=%v name=%v status=%v : ", updater.job.Namespace, updater.job.Name, status)
		err = updater.updateCRDStatus()
		if err != nil {
			log.Warning("get current status to update trainingJob status error: ", err.Error())
		}
		if updater.status.Phase == padv1.TrainingJobPhaseSucceeded || updater.status.Phase == padv1.TrainingJobPhaseFailed {
			log.Infof("Release Resource namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
			if updater.job.Spec.FaultTolerant {
				log.Infof("Release master, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
				if err := updater.releaseMaster(); err != nil {
					log.Error(err.Error())
				}
			}
			log.Infof("Release pserver, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
			if err := updater.releasePserver(); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

// InitResource is used to parse trainingJob and create trainingJob resources.
func (updater *TrainingJobUpdater) InitResource() {
	if updater.status.Phase == padv1.TrainingJobPhaseNone {
		log.Infof("set up trainingJob namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
		updater.parseTrainingJob()
		err := updater.updateCRDStatus()
		if err != nil {
			log.Warning("set up trainingJob to update trainingJob status error: ", err.Error())
		}
	}

	if updater.status.Phase == padv1.TrainingJobPhaseCreating {
		log.Infof("create trainingJob namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
		_ = updater.createTrainingJob()
		err := updater.updateCRDStatus()
		if err != nil {
			log.Warning("create trainingJob to update trainingJob status error: ", err.Error())
		}
		if updater.status.Phase == padv1.TrainingJobPhaseFailed {
			log.Infof("Release Resource for failed namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
			if updater.job.Spec.FaultTolerant {
				log.Infof("Release master, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
				if err := updater.releaseMaster(); err != nil {
					log.Error(err.Error())
				}
			}

			log.Infof("Release pserver, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
			if err := updater.releasePserver(); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

// Start is the main process of life cycle of a TrainingJob, including create resources, event process handle and
// status convert.
func (updater *TrainingJobUpdater) start() {
	log.Infof("start updater, namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
	go updater.InitResource()

	ticker := time.NewTicker(convertedTimerTicker)
	defer ticker.Stop()
	log.Infof("start ticker, namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
	for {
		select {
		case ev := <-updater.eventCh:
			switch ev.pet {
			case trainingJobEventDelete:
				log.Infof("Delete updater, namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
				if err := updater.deleteTrainingJob(); err != nil {
					log.Errorf(err.Error())
				}
				return
			}
		case <-ticker.C:
			updater.Convert()
			if updater.status.Phase == padv1.TrainingJobPhaseSucceeded || updater.status.Phase == padv1.TrainingJobPhaseFailed {
				if ticker != nil {
					log.Infof("stop ticker for job has done, namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
					ticker.Stop()
				}
			}
		}
	}
}
