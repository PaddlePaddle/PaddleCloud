package updater

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	log "github.com/inconshreveable/log15"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"

	paddlev1 "github.com/paddleflow/paddle-operator/pkg/apis/paddlepaddle/v1alpha1"
	trainingJobClient "github.com/paddleflow/paddle-operator/pkg/client/clientset/versioned"
	"github.com/paddleflow/paddle-operator/pkg/client/clientset/versioned/scheme"
)

var (
	// ErrorUnkownResourceType not supported resource
	ErrorUnkownResourceType = errors.New("UnknownResourceType")
)

// JobUpdater controls the life circle of one TrainingJob instance
type JobUpdater struct {
	Job            *paddlev1.TrainingJob
	kubeCli        kubernetes.Interface
	trainingJobCli trainingJobClient.Interface
	status         paddlev1.TrainingJobStatus
	recorder       record.EventRecorder
	autoclean      bool
	Additional     int32
	restartLimit   int
	outter         bool
}

// NewJobUpdater returns JobUpdater instance
func NewJobUpdater(job *paddlev1.TrainingJob, kubeCli kubernetes.Interface, jobCli trainingJobClient.Interface,
	auto bool, restartLimit int, outter bool) *JobUpdater {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeCli.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "TrainingJobController"})

	return &JobUpdater{
		Job:            job,
		kubeCli:        kubeCli,
		trainingJobCli: jobCli,
		status:         *job.Status.DeepCopy(),
		recorder:       recorder,
		autoclean:      auto,
		restartLimit:   restartLimit,
		outter:         outter,
	}
}

// UID return uid of a job instance
func (j *JobUpdater) UID() types.UID {
	return j.Job.ObjectMeta.UID
}

// Update updates jobupdater's job instance
func (j *JobUpdater) Update(job *paddlev1.TrainingJob) {
	log.Debug("Updating", "job", j.FullName(), "statue", job.Status)
	j.Job = job
}

// GetJob returns trainingjob instance
func (j *JobUpdater) GetJob() *paddlev1.TrainingJob {
	return j.Job
}

// Delete deletes trainingjob instance
func (j *JobUpdater) Delete() error {
	return j.deleteTrainingJob()
}

// FullName returns job's namespace and name
func (j *JobUpdater) FullName() string {
	return fmt.Sprintf("%s/%s", j.Job.Namespace, j.Job.Name)
}

// IsReleased returns if tj has released resource
func (j *JobUpdater) IsReleased() bool {
	return j.Job.Status.Released
}

// SetReleased set resource of job released
func (j *JobUpdater) SetReleased(release bool) {
	j.Job.Status.Released = true
}

func (j *JobUpdater) masterName() string {
	return fmt.Sprintf("%s/%s", j.Job.Namespace, j.Job.Spec.Master.ReplicaSpec.Name)
}

func (j *JobUpdater) pserverName() string {
	return fmt.Sprintf("%s/%s", j.Job.Namespace, j.Job.Spec.Pserver.ReplicaSpec.Name)
}

func (j *JobUpdater) trainerName() string {
	return fmt.Sprintf("%s/%s", j.Job.Namespace, j.Job.Spec.Trainer.ReplicaSpec.Name)
}

// Reconcile tries to get the job into the desired state
func (j *JobUpdater) Reconcile() error {
	released := j.IsReleased()
	log.Info("Reconciling TrainingJob", "job", j.FullName(), "current status phase", j.Job.Status.Phase)

	if j.Job.ObjectMeta.DeletionTimestamp != nil {
		log.Info("Deleted timestamp", "job", j.FullName(), "timestamp", j.Job.ObjectMeta.DeletionTimestamp.String())
		return nil
	}

	if j.Job.Status.Phase == paddlev1.TrainingJobPhaseNone {
		log.Info("Setting up", "job", j.FullName())
		if err := j.setup(); err != nil {
			j.status.Phase = paddlev1.ResourceStateFailed
			j.status.Reason = err.Error()
			log.Error("Error setting up TrainingJob", "job", j.FullName(), "err", err.Error())
		} else {
			j.status.Phase = paddlev1.TrainingJobPhaseCreating
			log.Info("Finish setting up TrainingJob", "job", j.FullName())
		}
		if err := j.updateCRDStatus(released); err != nil {
			log.Error("Error updating TrainingJob", "job", j.FullName(), "err", err.Error())
			return err
		}
	}

	if j.Job.Status.Phase == paddlev1.TrainingJobPhaseCreating {
		log.Info("Creating TrainingJob", "job", j.FullName())
		if err := j.createTrainingJob(); err != nil {
			log.Error("Error creating TrainingJob", "job", j.FullName(), "err", err.Error())
			j.status.Phase = paddlev1.ResourceStateFailed
			j.status.Reason = err.Error()
		} else {
			log.Info("Finish creating TrainingJob", "job", j.FullName())
		}

		if err := j.updateCRDStatus(released); err != nil {
			log.Error("Error updating TrainingJob", "job", j.FullName(), "err", err.Error())
			return err
		}

		phase, reason, err := j.GetStatus()
		if err != nil {
			log.Error("Error get TrainingJob status", "job", j.FullName(), "err", err.Error())
			return err
		}
		log.Info("TrainingJob GetStatus", "job", j.FullName(), "current phase", phase, "reason", reason)

		j.status.Phase = phase
		j.status.Reason = reason

		if phase == paddlev1.TrainingJobPhaseRunning {
			j.status.StartTime = v1.NewTime(time.Now())
			log.Info("Job started", "job", j.FullName(), "time", j.status.StartTime)
		}

		if err := j.updateCRDStatus(released); err != nil {
			log.Error("Error updating TrainingJob", "job", j.FullName(), "err", err.Error())
			return err
		}
	}

	if j.Job.Status.Phase == paddlev1.TrainingJobPhaseRunning {

		j.initLabelOfPods()
		go j.traceLabelOfPods()

		phase, reason, err := j.GetStatus()
		if err != nil {
			log.Error("Get trainingjob error", "job", j.FullName(), "err", err.Error())
			return err
		}

		j.status.Phase = phase
		j.status.Reason = reason
		if err := j.updateCRDStatus(released); err != nil {
			log.Error("Update trainingjob error", "job", j.FullName(), "err", err.Error())
			return err
		}
	}

	if j.Job.Status.Phase == paddlev1.TrainingJobPhaseScaling {
		if j.Additional != 0 {
			if err := j.scale(); err != nil {
				//TODO
				return err
			}
			j.Additional = 0
		}

		phase, reason, err := j.GetStatus()
		if err != nil {
			log.Error("Error get TrainingJob", "job", j.FullName(), "err", err.Error())
			return err
		}

		j.status.Phase = phase
		j.status.Reason = reason
		if err := j.updateCRDStatus(released); err != nil {
			log.Error("Error updating TrainingJob", "job", j.FullName(), "err", err.Error())
			return err
		}
	}

	if j.Job.Status.Phase == paddlev1.TrainingJobPhaseSucceeded ||
		j.Job.Status.Phase == paddlev1.TrainingJobPhaseFailed ||
		j.Job.Status.Phase == paddlev1.TrainingJobPhaseTimeout {
		if j.autoclean {
			log.Info("Releasing TrainingJob resource", "job", j.FullName(), "current status phase", j.Job.Status.Phase)
			if err := j.releaseTrainer(); err != nil {
				log.Error("Error releasing TrainingJob trainer resource", "job", j.FullName(), "err", err.Error())
				return err
			}
			log.Info("Finish releasing TrainingJob trainer resource", "job", j.FullName())

			if err := j.releaseMasterRoles(); err != nil {
				log.Error("Error releasing TrainingJob master/pserver resource", "job", j.FullName(), "err", err.Error())
				return err
			}
			log.Info("Finish releasing TrainingJob master/pserver resource", "job", j.FullName())

			j.recorder.Event(j.Job, corev1.EventTypeNormal, "Terminated", "All pods cleaned")
			j.SetReleased(true)
		} else {
			j.recorder.Event(j.Job, corev1.EventTypeNormal, "Terminated", "All pods kept")
		}
	}

	podsList, err := j.kubeCli.CoreV1().Pods(j.Job.Namespace).List(context.TODO(),
		v1.ListOptions{LabelSelector: PserverLabel + "=" + j.Job.Name})
	if err != nil {
		return err
	}
	log.Info("Searching trainingJob", j.FullName())
findFailedPserver:
	for _, pod := range podsList.Items {
		if j.Job.Spec.Pserver.ReplicaSpec == nil ||
			!strings.Contains(pod.GetNamespace()+"/"+pod.GetName(), j.pserverName()) {
			continue
		}
		log.Info("Find trainingJob", "job", j.FullName())
		for _, pod := range pod.Status.ContainerStatuses {
			if pod.RestartCount < int32(j.restartLimit) {
				continue
			}

			j.status.Phase = paddlev1.TrainingJobPhaseFailed
			j.status.Reason = "Pserver reached to restart limit!"
			break findFailedPserver
		}
	}

	if err := j.updateCRDStatus(released); err != nil {
		log.Error("Error updating TrainingJob", "job", j.FullName(), "err", err.Error())
		return err
	}

	return nil
}

func (j *JobUpdater) setup() error {

	if j.outter {
		var parser DefaultJobParser
		var err error
		j.Job, err = parser.NewTrainingJob(j.Job)
		if err != nil {
			log.Error("Error settting up", "job", j.FullName(), "err", err.Error())
		}

		return err
	}
	err := func() error {
		var parser DefaultJobParser
		err := parser.Validate(j.Job)
		if err != nil {
			log.Error("Validating error", "error", err.Error())
			return err
		}

		extEnv, err := parser.GetExtraEnv(j.Job, j.kubeCli)
		if err != nil {
			log.Error("Getting extra env failed", "error", err.Error())
			return err
		}

		j.Job = parser.ParseToTrainingJob(j.Job, extEnv)
		return nil
	}()

	if err != nil {
		log.Error("Error settting up", "name", j.FullName(), "error", err.Error())
	}

	return err
}

func (j *JobUpdater) updateCRDStatus(released bool) error {
	log.Debug("Updating TrainingJob status", "job", j.FullName(), "former status", j.Job.Status, "current status",
		j.status)
	if reflect.DeepEqual(j.status, j.Job.Status) && released == j.IsReleased() {
		log.Debug("Update TrainingJob skipped", "job", j.FullName(), "status", j.status)
		return nil
	}

	newJob := j.Job
	newJob.Status = j.status
	// sync trainingjob to apiserver
	newJob, err := j.trainingJobCli.PaddlepaddleV1alpha1().TrainingJobs(j.Job.Namespace).Update(context.TODO(), newJob)
	if err != nil {
		return err
	}

	j.Job = newJob
	return nil
}

// GetStatus get current status phase and reasion of job
func (j *JobUpdater) GetStatus() (paddlev1.TrainingJobPhase, string, error) {
	phase := j.status.Phase
	reason := ""

	trainers, err := j.kubeCli.BatchV1().Jobs(j.Job.Namespace).Get(context.TODO(),
		j.Job.Spec.Trainer.ReplicaSpec.Name, v1.GetOptions{})
	if err != nil {
		log.Error("error getting trainers", "name", j.trainerName(), "err", err.Error())
		return phase, reason, err
	}

	// total running
	totalRunning, err := j.jobTotalRunning()
	if err != nil {
		return phase, reason, err
	} else if totalRunning {
		phase = paddlev1.TrainingJobPhaseRunning
		reason = "all pods are running"
	}

	// the parallelism of batch/job trainer will be modified after success/failure
	total := *j.Job.Spec.Trainer.ReplicaSpec.Spec.Parallelism
	if j.Job.Spec.FaultTolerant {
		if trainers.Status.Failed == total {
			phase = paddlev1.TrainingJobPhaseFailed
			reason = "all trainer instances have failed"
			return phase, reason, nil
		} else if trainers.Status.Succeeded == total && trainers.Status.Active == 0 {
			phase = paddlev1.TrainingJobPhaseSucceeded
			reason = "all trainer instances have done"
		}
	} else {
		if trainers.Status.Failed != 0 {
			failedPods, err := j.findFailedTrainerPods()
			if err != nil {
				return phase, reason, err
			}

			podNameList := make([]string, 0)
			podNodeList := make([]string, 0)
			podReasonList := make([]string, 0)
			for _, pod := range failedPods {
				podNameList = append(podNameList, pod.Name)
				podNodeList = append(podNodeList, pod.Status.HostIP)

				status := pod.Status.ContainerStatuses
				if len(status) > 0 {
					podReasonList = append(podReasonList, fmt.Sprint(status[0].State))
				}
			}

			phase = paddlev1.TrainingJobPhaseFailed
			reason = fmt.Sprintf("trainer instances %s on %s have failed", podNameList, podNodeList)
			podFailReason := fmt.Sprintf("trainer instances %s on %s have failed, detailed reasons: %s", podNameList,
				podNodeList, podReasonList)
			j.recorder.Event(j.Job, corev1.EventTypeWarning, "Pods Failed", podFailReason)
		} else if trainers.Status.Succeeded == total && trainers.Status.Active == 0 {
			phase = paddlev1.TrainingJobPhaseSucceeded
			reason = "all trainer instances have done"
		}
	}

	timeLimit := int64(j.Job.Spec.Annotations.Walltime)
	currentTime := time.Now()
	if timeLimit != 0 && trainers.Status.Active != 0 && !j.Job.Status.StartTime.IsZero() &&
		int64((currentTime.Sub(j.Job.Status.StartTime.Time)).Seconds()) > timeLimit {
		phase = paddlev1.TrainingJobPhaseTimeout
		reason = "timeout!"
		log.Warn("Job started", "job", j.Job.Name, "start time",
			j.Job.Status.StartTime.Format("2018-01-01 23:00:00"), "current time",
			currentTime.Format("2018-01-01 23:00:00"), "timeLimit", timeLimit)
	}

	if j.Additional != 0 {
		phase = paddlev1.TrainingJobPhaseScaling
		reason = fmt.Sprintf("need scale")
	}

	return phase, reason, nil
}

func (j *JobUpdater) createTrainingJob() error {

	frameWork := j.Job.Spec.FrameWork

	if j.Job.Spec.FaultTolerant {
		log.Debug("Creating master", "name", j.masterName())
		if err := j.createResource(paddlev1.MASTER); err != nil {
			return err
		}
	}

	if frameWork == nil && !j.Job.Spec.LocalJob && !j.Job.Spec.IsNccl ||
		frameWork != nil && frameWork.Type == paddlev1.Multi {
		log.Debug("Creating pserver", "name", j.pserverName())
		if err := j.createResource(paddlev1.PSERVER); err != nil {
			return err
		}
	}

	log.Debug("creating trainer", "name", j.trainerName())
	if err := j.createTrainer(); err != nil {
		return err
	}

	return nil
}

func (j *JobUpdater) createResource(rt paddlev1.TrainingResourceType) error {
	resource := new(appsv1.ReplicaSet)
	switch rt {
	case paddlev1.MASTER:
		resource = j.Job.Spec.Master.ReplicaSpec
	case paddlev1.PSERVER:
		resource = j.Job.Spec.Pserver.ReplicaSpec
	default:
		return ErrorUnkownResourceType
	}

	if _, err := j.kubeCli.AppsV1().ReplicaSets(resource.Namespace).Get(context.TODO(), resource.Name, v1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			if _, err := j.kubeCli.AppsV1().ReplicaSets(resource.Namespace).Create(context.TODO(), resource, v1.CreateOptions{}); err != nil {
				log.Error("Error creating resource", "namespace", resource.Namespace, "name", resource.Name, "err",
					err.Error())
				return err
			}
			log.Debug("Finish creating resource", "namespace", resource.Namespace, "name", resource.Name)
			return nil
		}
		log.Error("Drror getting resource", "namespace", resource.Namespace, "name", resource.Name, "err", err.Error())
		return err
	}

	log.Debug("Resource already existing, skipping", "namespace", resource.Namespace, "name", resource.Name)
	return nil
}

func (j *JobUpdater) createTrainer() error {
	if _, err := j.kubeCli.BatchV1().Jobs(j.Job.Namespace).Get(context.TODO(), j.Job.Spec.Trainer.ReplicaSpec.Name, v1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			if _, err = j.kubeCli.BatchV1().Jobs(j.Job.Namespace).Create(context.TODO(), j.Job.Spec.Trainer.ReplicaSpec, v1.CreateOptions{}); err != nil {
				log.Error("Error creating trainer", "name", j.trainerName(), "err", err.Error())
				return err
			}
			log.Debug("Finishing creating trainer", "name", j.trainerName())
			return nil
		}
		log.Error("Error getting trainer", "name", j.trainerName(), "err", err.Error())
		return err
	}

	log.Debug("Trainer already existing skipping", "name", j.trainerName())
	return nil
}

func (j *JobUpdater) deleteTrainingJob() error {

	frameWork := j.Job.Spec.FrameWork

	if j.Job.Spec.FaultTolerant {
		log.Debug("deleting master", "name", j.masterName())
		if err := j.deleteResource(paddlev1.MASTER); err != nil {
			log.Error("error deleting master", "name", j.masterName(), "err", err.Error())
			return err
		}
	}
	if frameWork == nil && !j.Job.Spec.LocalJob && !j.Job.Spec.IsNccl ||
		frameWork != nil && frameWork.Type == paddlev1.Multi {
		log.Debug("Deleting pserver", "name", j.pserverName())
		if err := j.deleteResource(paddlev1.PSERVER); err != nil {
			log.Error("Deleting pserver error", "name", j.pserverName(), "reason", err.Error())
			return err
		}
	}

	log.Debug("deleting trainer", "name", j.trainerName())
	if err := j.deleteTrainer(); err != nil {
		log.Error("Deleting trainer error", "name", j.trainerName(), "reason", err.Error())
		return err
	}

	return nil
}

func (j *JobUpdater) deleteResource(rt paddlev1.TrainingResourceType) error {
	if err := j.releaseResource(rt); err != nil {
		return err
	}

	var gracePeriodSeconds *int64 = nil
	switch rt {
	case paddlev1.MASTER:
		gracePeriodSeconds = j.Job.Spec.Master.GracePeriodSeconds
	case paddlev1.PSERVER:
		gracePeriodSeconds = j.Job.Spec.Pserver.GracePeriodSeconds
	default:
		return ErrorUnkownResourceType
	}

	resourceName := j.Job.Name + "-" + string(rt)
	if err := j.kubeCli.AppsV1().ReplicaSets(j.Job.Namespace).Delete(context.TODO(), resourceName, v1.DeleteOptions{GracePeriodSeconds: gracePeriodSeconds}); err != nil {
		if apierrors.IsNotFound(err) {
			log.Debug("Resource not found, skipped", "namespace", j.Job.Namespace, "name", resourceName)
			return nil
		}
		return err
	}
	log.Debug("Finishing releasing", "namespace", j.Job.Namespace, "name", resourceName)
	return nil
}

func (j *JobUpdater) deleteTrainer() error {
	if err := j.releaseTrainer(); err != nil {
		return err
	}

	if err := j.kubeCli.BatchV1().Jobs(j.Job.Namespace).Delete(context.TODO(), j.Job.Spec.Trainer.ReplicaSpec.Name, v1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			log.Debug("trainer not exist skipped", "name", j.trainerName())
			return nil
		}
		return err
	}
	log.Debug("Finishing deleting trainer", "name", j.trainerName())
	return nil
}

func (j *JobUpdater) releaseMasterRoles() error {

	frameWork := j.Job.Spec.FrameWork

	if j.Job.Spec.FaultTolerant {
		if err := j.releaseResource(paddlev1.MASTER); err != nil {
			log.Error("Error releasing master", "name", j.masterName(), "err", err)
			return err
		}
	}

	if frameWork == nil && !j.Job.Spec.LocalJob && !j.Job.Spec.IsNccl ||
		frameWork != nil && frameWork.Type == paddlev1.Multi {
		if err := j.releaseResource(paddlev1.PSERVER); err != nil {
			log.Error("Error releasing pserver", "name", j.pserverName(), "err", err)
			return err
		}

	}
	return nil
}

func (j *JobUpdater) releaseResource(rt paddlev1.TrainingResourceType) error {
	resourceName := ""
	switch rt {
	case paddlev1.MASTER:
		resourceName = j.Job.Spec.Master.ReplicaSpec.Name
	case paddlev1.PSERVER:
		resourceName = j.Job.Spec.Pserver.ReplicaSpec.Name
	default:
		return ErrorUnkownResourceType
	}

	resource, getErr := j.kubeCli.AppsV1().ReplicaSets(j.Job.Namespace).Get(context.TODO(), resourceName, v1.GetOptions{})
	if getErr != nil {
		if apierrors.IsNotFound(getErr) {
			log.Debug("Resouce instance not exist, skipped", "namespace", j.Job.Namespace, "name", resourceName)
			return nil
		}
		log.Error("Error getting instance", "namespace", j.Job.Namespace, "name", resourceName, "err", getErr)
		return getErr
	}

	if *resource.Spec.Replicas != 0 {
		var replicas int32
		replicas = 0
		resource.Spec.Replicas = &replicas
		if _, err := j.kubeCli.AppsV1().ReplicaSets(j.Job.Namespace).Update(context.TODO(), resource, v1.UpdateOptions{}); err != nil {
			log.Error("error setting replicas to 0", "namespace", j.Job.Namespace, "name", resourceName, "err", err.Error())
			return err
		}
	}

	if resource.Status.FullyLabeledReplicas != 0 {
		key := "paddle-job-" + rt
		labels := Labels(map[string]string{
			string(key): j.Job.Name,
		})

		selector, _ := labels.LabelsParser()
		options := v1.ListOptions{
			LabelSelector: selector,
		}

		if err := j.kubeCli.CoreV1().Pods(j.Job.Namespace).DeleteCollection(context.TODO(), v1.DeleteOptions{}, options); err != nil {
			log.Error("error deleting resource pods", "namespace", j.Job.Namespace, "name", resourceName, "err", err.Error())
			return err
		}
	}

	return nil
}

func (j *JobUpdater) releaseTrainer() error {
	jobNs := j.Job.Namespace
	jobName := j.Job.Spec.Trainer.ReplicaSpec.Name

	jobSpec, getErr := j.kubeCli.BatchV1().Jobs(jobNs).Get(context.TODO(), jobName, v1.GetOptions{})
	if getErr != nil {
		if apierrors.IsNotFound(getErr) {
			return nil
		}
		log.Error("Error getting job spec for TrainingJob trainer", "name", j.trainerName())
		return getErr
	}

	if *jobSpec.Spec.Parallelism != 0 {
		log.Debug("Reset parallelism to zero for TrainingJob trainer", "name", j.trainerName())
		var parallism int32
		parallism = 0
		jobSpec.Spec.Parallelism = &parallism
		if _, err := j.kubeCli.BatchV1().Jobs(jobNs).Update(context.TODO(), jobSpec, v1.UpdateOptions{}); err != nil {
			log.Error("Error resetting parallelism for TrainingJob trainer", "name", j.trainerName())
			return err
		}
	}

	labels := Labels(map[string]string{
		TrainerLabel: j.Job.Name,
	})
	selector, _ := labels.LabelsParser()
	options := v1.ListOptions{
		LabelSelector: selector,
	}

	if err := j.kubeCli.CoreV1().Pods(jobNs).DeleteCollection(context.TODO(), v1.DeleteOptions{}, options); err != nil {
		log.Error("Error deleting pods of TrainingJob trainer", "name", j.trainerName())
		return err
	}

	return nil
}

func (j *JobUpdater) jobTotalRunning() (bool, error) {
	frameWork := j.Job.Spec.FrameWork

	if j.Job.Spec.FaultTolerant {
		masterRunning, err := j.masterRoleTotalRunning(paddlev1.MASTER)
		if err != nil || !masterRunning {
			return false, err
		}
	}

	if frameWork == nil && !j.Job.Spec.LocalJob && !j.Job.Spec.IsNccl ||
		frameWork != nil && frameWork.Type == paddlev1.Multi {
		pserverRunning, err := j.masterRoleTotalRunning(paddlev1.PSERVER)
		if err != nil || !pserverRunning {
			return false, err
		}
	}

	return j.trainerTotalRunning()
}

func (j *JobUpdater) masterRoleTotalRunning(rt paddlev1.TrainingResourceType) (bool, error) {
	var resourceName string
	switch rt {
	case paddlev1.MASTER:
		resourceName = j.Job.Spec.Master.ReplicaSpec.Name
	case paddlev1.PSERVER:
		resourceName = j.Job.Spec.Pserver.ReplicaSpec.Name
	default:
		return false, ErrorUnkownResourceType
	}
	resource, err := j.kubeCli.AppsV1().ReplicaSets(j.Job.Namespace).Get(context.TODO(), resourceName, v1.GetOptions{})
	if err != nil {
		return false, err
	}

	log.Debug("Resource status", "namespace", j.Job.Namespace, "name", resourceName, "status", resource.Status)

	return resource.Status.ReadyReplicas >= *resource.Spec.Replicas, nil
}

func (j *JobUpdater) trainerTotalRunning() (bool, error) {
	trainerName := j.Job.Spec.Trainer.ReplicaSpec.Name
	trainers, err := j.kubeCli.BatchV1().Jobs(j.Job.Namespace).Get(context.TODO(), trainerName, v1.GetOptions{})
	if err != nil {
		return false, err
	}

	log.Debug("Trainer status", "namespace", j.Job.Namespace, "name", trainerName, "status", trainers.Status)
	podsList, err := j.kubeCli.CoreV1().Pods(j.Job.Namespace).List(context.TODO(), v1.ListOptions{LabelSelector: TrainerLabel + "=" + j.Job.Name})
	var runningPodCount int32
	for _, pod := range podsList.Items {
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded {
			runningPodCount++
		}
	}

	return runningPodCount >= *trainers.Spec.Parallelism, nil
}

func (j *JobUpdater) findFailedTrainerPods() ([]*corev1.Pod, error) {
	failedPods := make([]*corev1.Pod, 0)

	podsList, err := j.kubeCli.CoreV1().Pods(j.Job.Namespace).List(context.TODO(), v1.ListOptions{LabelSelector: TrainerLabel + "=" + j.Job.Name})
	if err != nil {
		return failedPods, err
	}
	for _, pod := range podsList.Items {
		if pod.Status.Phase == corev1.PodFailed {
			failedPods = append(failedPods, &pod)
		}
	}

	return failedPods, nil
}

func (j *JobUpdater) scale() (err error) {
	jobNs := j.Job.Namespace
	jobName := j.Job.Spec.Trainer.ReplicaSpec.Name
	jobSpec, err := j.kubeCli.BatchV1().Jobs(jobNs).Get(context.TODO(), jobName, v1.GetOptions{})
	if err != nil {
		return err
	}

	newParallelism := *jobSpec.Spec.Parallelism + j.Additional
	newBackoffLimit := *jobSpec.Spec.BackoffLimit
	if j.Additional < 0 {
		newBackoffLimit -= j.Additional
	}
	jobSpec.Spec.Parallelism = &newParallelism
	jobSpec.Spec.BackoffLimit = &newBackoffLimit
	j.Job.Spec.Trainer.ReplicaSpec.Spec.Parallelism = &newParallelism
	log.Debug("Scaling job", "namespace", jobNs, "name", jobName, "new instance num", newParallelism)
	if _, err := j.kubeCli.BatchV1().Jobs(jobNs).Update(context.TODO(), jobSpec, v1.UpdateOptions{}); err != nil {
		log.Debug("Failed to scale job", "namespace", jobNs, "name", jobName, "error", err.Error())
		return err
	}

	return nil
}

func (j *JobUpdater) initLabelOfPods() {

	if !j.Job.Spec.Trainer.IndexSucceed {
		success, err := j.addLabelToPods(TRAINER)
		if err == nil && success {
			j.Job.Spec.Trainer.IndexSucceed = true
		} else {
			log.Error("Add label to trainer failed", "name", j.trainerName())
		}

	}

	frameWork := j.Job.Spec.FrameWork
	if frameWork != nil && frameWork.Type == paddlev1.Multi &&
		!j.Job.Spec.Pserver.IndexSucceed {
		success, err := j.addLabelToPods(PSERVER)
		if err == nil && success {
			j.Job.Spec.Pserver.IndexSucceed = true
		} else {
			log.Error("Add label to pserver failed", "name", j.pserverName())
		}

	}
}

func (j *JobUpdater) addLabelToPods(podType PodType) (bool, error) {

	var labelOptions v1.ListOptions
	var podName string
	var desiredPodNum int
	switch podType {
	case TRAINER:
		labelOptions = v1.ListOptions{LabelSelector: TrainerLabel + "=" + j.Job.Name}
		podName = j.trainerName()
		desiredPodNum = int(*j.Job.Spec.Trainer.ReplicaSpec.Spec.Parallelism)

	case PSERVER:
		labelOptions = v1.ListOptions{LabelSelector: PserverLabel + "=" + j.Job.Name}
		podName = j.pserverName()
		desiredPodNum = int(*j.Job.Spec.Pserver.ReplicaSpec.Spec.Replicas)

	}

	log.Info("Start to add label", "podKind", podName)

	podsList, err := j.kubeCli.CoreV1().Pods(j.Job.Namespace).List(context.TODO(), labelOptions)
	if err != nil {
		return false, err
	}

	for idx, pod := range podsList.Items {
		oldPod := pod
		if oldPod.Status.Phase != corev1.PodRunning || oldPod.DeletionTimestamp != nil {
			continue
		}

		labels := oldPod.GetLabels()
		labels[podName+"-idx"] = strconv.Itoa(idx)
		oldPod.SetLabels(labels)

		if _, err := j.kubeCli.CoreV1().Pods(j.Job.Namespace).Update(context.TODO(), &oldPod, v1.UpdateOptions{}); err != nil {
			log.Error("Resource status updated failed", "namespace", j.Job.Namespace, "pod", oldPod.Name)
			return false, err
		}

		// FIXME: the loop will now index up to desiredPodNum+1 pods instead of desiredPodNum. Is that the desired behavior?
		if idx >= desiredPodNum {
			log.Warn("Idx exceededs desired pod number", "current id", idx)
			break
		}
	}

	return true, nil
}

func (j *JobUpdater) traceLabelOfPods() {

	if err := j.traceAddLabelToPods(TRAINER); err != nil {
		log.Error("Trace pod label of pods failed", "podKind", TRAINER)
	}

	frameWork := j.Job.Spec.FrameWork
	if frameWork != nil && frameWork.Type == paddlev1.Multi {
		if err := j.traceAddLabelToPods(PSERVER); err != nil {
			log.Error("Trace pod label of pods failed", "podKind", PSERVER)

		}
	}
}

func (j *JobUpdater) traceAddLabelToPods(podType PodType) error {

	indexMap := make(map[int]string)
	unIndexedPod := make([]*corev1.Pod, 0)
	desiredPodNum := 0
	var labelOptions v1.ListOptions
	var podKind string
	switch podType {
	case TRAINER:
		labelOptions = v1.ListOptions{LabelSelector: TrainerLabel + "=" + j.Job.Name}
		podKind = j.trainerName()
		desiredPodNum = int(*j.Job.Spec.Trainer.ReplicaSpec.Spec.Parallelism)

	case PSERVER:
		labelOptions = v1.ListOptions{LabelSelector: PserverLabel + "=" + j.Job.Name}
		podKind = j.pserverName()
		desiredPodNum = int(*j.Job.Spec.Pserver.ReplicaSpec.Spec.Replicas)

	}
	log.Info("Start trace pod", "podKind", podKind)

	podsList, err := j.kubeCli.CoreV1().Pods(j.Job.Namespace).List(context.TODO(), labelOptions)
	if err != nil {
		return err
	}

	for _, pod := range podsList.Items {
		oldPod := pod
		if oldPod.Status.Phase != corev1.PodRunning && oldPod.Status.Phase != corev1.PodSucceeded {
			continue
		}

		if oldPod.DeletionTimestamp != nil {
			continue
		}
		labels := oldPod.GetLabels()

		v, exist := labels[podKind+"-idx"]
		if exist {
			if id, err := strconv.Atoi(v); err == nil {
				indexMap[id] = oldPod.Name
			}
		} else {
			unIndexedPod = append(unIndexedPod, &oldPod)
		}

	}

	log.Info("UnIndexed pod", "info", unIndexedPod)

	for id := 0; id < desiredPodNum; id++ {
		podLen := len(unIndexedPod)
		if _, exist := indexMap[id]; !exist && podLen > 0 {

			oldPod := *unIndexedPod[0]
			indexMap[id] = oldPod.Name
			labels := oldPod.GetLabels()
			labels[podKind+"-idx"] = strconv.Itoa(id)
			oldPod.SetLabels(labels)

			if _, err := j.kubeCli.CoreV1().Pods(j.Job.Namespace).Update(context.TODO(), &oldPod, v1.UpdateOptions{}); err != nil {
				log.Error("Resource status updated failed", "namespace", j.Job.Namespace, "name", oldPod.Name)
				return err
			}

			if podLen > 1 {
				unIndexedPod = unIndexedPod[1:]
			} else {
				log.Info("Finished trace label of job", "podKind", podKind)
				break
			}

		}
	}

	log.Info("Traced indexMap", "index info", indexMap)

	return nil
}
