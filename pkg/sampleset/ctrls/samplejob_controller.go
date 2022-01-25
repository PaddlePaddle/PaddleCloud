// Copyright 2021 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctrls

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/paddleflow/paddle-operator/api/v1alpha1"
	"github.com/paddleflow/paddle-operator/controllers/extensions/common"
	"github.com/paddleflow/paddle-operator/controllers/extensions/driver"
	"github.com/paddleflow/paddle-operator/controllers/extensions/utils"
)

// SampleJobReconciler reconciles a SampleJob object
type SampleJobReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=samplejobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=samplejobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=samplejobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=samplesets,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=samplesets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=statefulsets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the SampleJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SampleJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("===============  Reconcile  ==============")

	// 1. get SampleJob
	sampleJob := &v1alpha1.SampleJob{}
	if err := r.Get(ctx, req.NamespacedName, sampleJob); err != nil {
		return utils.RequeueWithError(client.IgnoreNotFound(err))
	}

	// 2. if SampleJob has not finalizer and has not deletion timestamp, then add finalizer and requeue
	if !utils.HasFinalizer(&sampleJob.ObjectMeta, GetSampleJobFinalizer(req.Name)) &&
		!utils.HasDeletionTimestamp(&sampleJob.ObjectMeta) {
		r.Log.Info("add finalizer successfully")
		return r.AddFinalizer(ctx, sampleJob)
	}

	// 3. if SampleJob has deletion timestamp, then delete finalizer and resources create by this controller
	if utils.HasDeletionTimestamp(&sampleJob.ObjectMeta) {
		return r.deleteSampleJob(ctx, sampleJob)
	}

	// 4. check if job type is exists or registered in JobTypeMap
	if _, exists := JobTypeMap[sampleJob.Spec.Type]; !exists {
		err := fmt.Errorf("job type %s is not supported", sampleJob.Spec.Type)
		r.Log.Error(err, "please set field spec.type correctly and try again")
		r.Recorder.Event(sampleJob, v1.EventTypeWarning, common.ErrorJobTypeNotSupport, err.Error())
		return utils.NoRequeue()
	}

	// 5. check if SampleSet is exists and get SampleSet
	if sampleJob.Spec.SampleSetRef == nil || sampleJob.Spec.SampleSetRef.Name == "" {
		err := fmt.Errorf("samplejob %s spec.sampleSetRef is empty", req.Name)
		r.Log.Error(err, "please set field spec.sampleSetRef and try again")
		r.Recorder.Event(sampleJob, v1.EventTypeWarning, common.ErrorSampleSetNotExist, err.Error())
		return utils.NoRequeue()
	}
	sampleSetKey := types.NamespacedName{
		Name:      sampleJob.Spec.SampleSetRef.Name,
		Namespace: req.Namespace,
	}
	sampleSet := &v1alpha1.SampleSet{}
	if err := r.Get(ctx, sampleSetKey, sampleSet); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Error(err, "sampleset not found", "sampleset", sampleSetKey.String())
			r.Recorder.Event(sampleJob, v1.EventTypeWarning, common.ErrorSampleSetNotExist, err.Error())
			return utils.NoRequeue()
		}
		return utils.RequeueWithError(err)
	}

	// 6. update SampleSet cache status to trigger SampleSet controller collect new cache info
	if sampleSet.Status.CacheStatus != nil {
		sampleSet.Status.CacheStatus.DiskUsageRate = ""
		if err := r.Status().Update(ctx, sampleSet); err != nil {
			return utils.RequeueWithError(err)
		}
	}

	// 7. Get driver and construct reconcile context
	var driverName v1alpha1.DriverName
	if sampleSet.Spec.CSI == nil {
		driverName = driver.DefaultDriver
	} else {
		driverName = sampleSet.Spec.CSI.Driver
	}
	CSIDriver, err := driver.GetDriver(driverName)
	if err != nil {
		r.Log.Error(err, "get driver error")
		r.Recorder.Event(sampleJob, v1.EventTypeWarning,
			common.ErrorDriverNotExist, err.Error())
		return utils.NoRequeue()
	}
	request := &ctrl.Request{NamespacedName: sampleSetKey}
	RCtx := &common.ReconcileContext{
		Ctx:      ctx,
		Client:   r.Client,
		Req:      request,
		Log:      r.Log,
		Scheme:   r.Scheme,
		Recorder: r.Recorder,
	}
	// 8. construct SampleJob Controller
	sjc := NewSampleJobController(sampleJob, CSIDriver, RCtx)

	return sjc.reconcilePhase()
}

// AddFinalizer add finalizer to SampleJob
func (r *SampleJobReconciler) AddFinalizer(ctx context.Context, sampleJob *v1alpha1.SampleJob) (ctrl.Result, error) {
	sampleJobFinalizer := GetSampleJobFinalizer(sampleJob.Name)
	sampleJob.Finalizers = append(sampleJob.Finalizers, sampleJobFinalizer)
	if err := r.Update(ctx, sampleJob); err != nil {
		return utils.RequeueWithError(err)
	}
	return utils.NoRequeue()
}

func (r *SampleJobReconciler) deleteSampleJob(ctx context.Context, sampleJob *v1alpha1.SampleJob) (ctrl.Result, error) {
	// TODO: clean cronjob

	sampleJobFinalizer := GetSampleJobFinalizer(sampleJob.Name)
	if utils.HasFinalizer(&sampleJob.ObjectMeta, sampleJobFinalizer) {
		utils.RemoveFinalizer(&sampleJob.ObjectMeta, sampleJobFinalizer)
	}
	if err := r.Update(ctx, sampleJob); err != nil {
		return utils.RequeueWithError(err)
	}

	r.Log.Info("==== deleted all resource ====")
	return utils.NoRequeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *SampleJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1.Pod{},
		common.IndexerKeyRuntime,
		RuntimePodIndexerFunc); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.SampleJob{}).
		Owns(&v1.Pod{}).
		Complete(r)
}

type SampleJobController struct {
	Controller
	SampleJob *v1alpha1.SampleJob
}

func NewSampleJobController(
	sampleJob *v1alpha1.SampleJob,
	CSIDriver driver.Driver,
	ctx *common.ReconcileContext) *SampleJobController {
	return &SampleJobController{
		Controller: Controller{
			Driver:           CSIDriver,
			Sample:           sampleJob,
			ReconcileContext: ctx,
		},
		SampleJob: sampleJob,
	}
}

func (s *SampleJobController) reconcilePhase() (ctrl.Result, error) {
	// Reconcile the phase of SampleSet from None to Ready
	switch s.SampleJob.Status.Phase {
	case common.SampleJobNone:
		return s.reconcileNone()
	case common.SampleJobPending:
		return s.reconcilePending()
	case common.SampleJobRunning:
		return s.reconcileRunning()
	case common.SampleJobFailed:
		return s.reconcileFailed()
	case common.SampleJobSucceeded:
		return s.reconcileSucceeded()
	}
	s.Log.Error(fmt.Errorf("phase %s not support", s.SampleJob.Status.Phase), "")
	return utils.NoRequeue()
}

// reconcileNone job: post
func (s *SampleJobController) reconcileNone() (ctrl.Result, error) {
	s.Log.Info("==== reconcileNone ====")
	// 1. if terminate is true then terminate SampleJobs that in processing
	if s.SampleJob.Spec.Terminate {
		if err := s.PostTerminateSignal(); err != nil {
			return utils.RequeueWithError(err)
		}
		time.Sleep(10 * time.Second)
	}

	// 2. if job type is sync then check if secrets is exists
	if s.SampleJob.Spec.Type == common.JobTypeSync {
		if s.SampleJob.Spec.SecretRef == nil {
			e := errors.New("spec.secretRef should not be empty")
			s.Log.Error(e, "spec.secretRef is nil")
			s.Recorder.Event(s.SampleJob, v1.EventTypeWarning, common.ErrorSecretNotExist, e.Error())
			return utils.NoRequeue()
		}
		if s.SampleJob.Spec.SecretRef.Name == "" {
			err := errors.New("spec.secretRef.name is not set")
			s.Log.Error(err, "spec.secretRef.name is empty string")
			s.Recorder.Event(s.SampleJob, v1.EventTypeWarning, common.ErrorSecretNotExist, err.Error())
			return utils.NoRequeue()
		}
		exist, err := s.ResourcesExist(SourceSecret)
		if err != nil {
			return utils.RequeueWithError(err)
		}
		if !exist {
			err := fmt.Errorf("secret %s is not exist", s.SampleJob.Spec.SecretRef.Name)
			s.Log.Error(err, "please create secret object and named it as spec.secretRef.name")
			s.Recorder.Event(s.SampleJob, v1.EventTypeWarning, common.ErrorSecretNotExist, err.Error())
			return utils.NoRequeue()
		}
	}

	// 3. if job name is none then generate job name and post options to server
	if s.SampleJob.Status.JobName == "" {
		jobName := uuid.NewUUID()
		jobType := JobTypeMap[s.SampleJob.Spec.Type]
		err := s.PostJobOptions(jobName, jobType)
		if err != nil {
			if errors.Is(err, optionError) {
				return utils.NoRequeue()
			}
			return utils.RequeueWithError(err)
		}
		s.SampleJob.Status.JobName = jobName
	}

	// 4. update SampleSet phase to pending
	s.SampleJob.Status.Phase = common.SampleJobPending
	err := s.UpdateResourceStatus(s.SampleJob, SampleJob)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	s.Log.Info("update samplejob phase to pending")
	return utils.NoRequeue()
}

func (s *SampleJobController) reconcilePending() (ctrl.Result, error) {
	s.Log.Info("==== reconcilePending ====")

	// 1. get job result from runtime server
	jobType := JobTypeMap[s.SampleJob.Spec.Type]
	jobName := s.SampleJob.Status.JobName
	result, err := s.GetJobResult(jobName, jobType)
	if err != nil {
		s.Log.Info("wait util job to be processed", "error", err.Error())
		return utils.RequeueAfter(10 * time.Second)
	}
	// 2 .if sync job status is running, then update SampleJob phase to Running
	if result != nil && result.Status == common.JobStatusRunning {
		s.SampleJob.Status.Phase = common.SampleJobRunning
	}
	// 3. if sync job status is failed, then update SampleJob phase to Failed
	if result != nil && result.Status == common.JobStatusFail {
		s.SampleJob.Status.Phase = common.SampleJobFailed
		e := errors.New(result.Message)
		s.Log.Error(e, "do job error", "jobName", jobName, "jobType", jobType.Name)
		s.Recorder.Event(s.SampleJob, v1.EventTypeWarning, jobType.ErrorDoJob(), e.Error())
	}
	// 3. if sync job status is success, then update SampleJob phase to Succeeded
	if result == nil {
		s.SampleJob.Status.Phase = common.SampleJobSucceeded
	}

	err = s.UpdateResourceStatus(s.SampleJob, SampleJob)
	if err != nil {
		return utils.RequeueWithError(err)
	}

	s.Log.Info("update samplejob phase", "phase", s.SampleJob.Status.Phase)
	return utils.NoRequeue()
}

func (s *SampleJobController) reconcileRunning() (ctrl.Result, error) {
	s.Log.Info("==== reconcileRunning ====")

	// 1. get job result from runtime server
	jobType := JobTypeMap[s.SampleJob.Spec.Type]
	jobName := s.SampleJob.Status.JobName
	result, err := s.GetJobResult(jobName, jobType)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	// 2 .if sync job status is running, then update SampleJob phase to Running
	if result != nil && result.Status == common.JobStatusRunning {
		s.Log.Info("wait util job is done")
		return utils.RequeueAfter(10 * time.Second)
	}
	// 3. if sync job status is failed, then update SampleJob phase to Failed
	if result != nil && result.Status == common.JobStatusFail {
		s.SampleJob.Status.Phase = common.SampleJobFailed
		e := errors.New(result.Message)
		s.Log.Error(e, "do job error", "jobName", jobName, "jobType", jobType.Name)
		s.Recorder.Event(s.SampleJob, v1.EventTypeWarning, jobType.ErrorDoJob(), e.Error())
	}
	// 3. if sync job status is success, then update SampleJob phase to Succeeded
	if result == nil {
		s.SampleJob.Status.Phase = common.SampleJobSucceeded
	}

	err = s.UpdateResourceStatus(s.SampleJob, SampleJob)
	if err != nil {
		return utils.RequeueWithError(err)
	}

	s.Log.Info("update samplejob phase", "phase", s.SampleJob.Status.Phase)
	return utils.NoRequeue()
}

func (s *SampleJobController) reconcileFailed() (ctrl.Result, error) {
	s.Log.Info("==== reconcileFailed ====")

	jobName := s.SampleJob.Status.JobName
	jobType := JobTypeMap[s.SampleJob.Spec.Type]
	runtimeName := s.GetRuntimeName(s.Req.Name)
	serviceName := s.GetServiceName(s.Req.Name)
	baseUri := utils.GetBaseUriByIndex(runtimeName, serviceName, 0)

	// 1. get the options from runtime server by jobName
	oldOptions := jobType.Options()
	err := utils.GetJobOption(oldOptions, jobName, baseUri, jobType.OptionPath)
	if err != nil {
		e := fmt.Errorf("get %s option error: %s", jobType.Name, err.Error())
		return utils.RequeueWithError(e)
	}

	// 3. create options for check if SampleJob is updated by user
	newOptions := jobType.Options()
	err = s.CreateJobOptions(newOptions, jobType)
	if err != nil {
		if errors.Is(err, optionError) {
			return utils.NoRequeue()
		}
		return utils.RequeueWithError(err)
	}

	// 4. if options have not updated, then no requeue
	if reflect.DeepEqual(oldOptions, newOptions) {
		s.Log.Info("options has not update", "jobName", jobName, "jobType", jobType.Name)
		return utils.NoRequeue()
	}

	// 5. if SampleJob updated by user and options has changed,
	// then delete old jobName and return SampleJob phase to None.
	s.SampleJob.Status.JobName = ""
	s.SampleJob.Status.Phase = common.SampleJobNone
	err = s.UpdateResourceStatus(s.SampleJob, SampleJob)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	s.Log.Info("options changed, return SampleJob phase to None",
		"jobName", jobName, "jobType", jobType.Name)
	return utils.NoRequeue()
}

func (s *SampleJobController) reconcileSucceeded() (ctrl.Result, error) {
	s.Log.Info("==== reconcileSucceeded ====")

	jobType := JobTypeMap[s.SampleJob.Spec.Type]
	s.Log.Info("job succeeded and it cannot be updated")
	s.Recorder.Event(s.SampleJob, v1.EventTypeNormal, jobType.DoJobSuccessfully(),
		"samplejob cannot be updated after succeeded")
	return utils.NoRequeue()
}
