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
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "github.com/paddleflow/paddle-operator/api/v1"
	"github.com/paddleflow/paddle-operator/api/v1alpha1"
	"github.com/paddleflow/paddle-operator/controllers/extensions/common"
	"github.com/paddleflow/paddle-operator/controllers/extensions/driver"
	"github.com/paddleflow/paddle-operator/controllers/extensions/utils"
)

// SampleSetReconciler reconciles a SampleSet object
type SampleSetReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=samplesets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=samplesets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=samplesets/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=samplejobs,verbs=get;list;watch
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=paddlejobs,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the SampleSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SampleSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.V(1).Info("===============  Reconcile  ==============")

	// 1. Get SampleSet
	sampleSet := &v1alpha1.SampleSet{}
	if err := r.Get(ctx, req.NamespacedName, sampleSet); err != nil {
		return utils.RequeueWithError(client.IgnoreNotFound(err))
	}

	// 2. if SampleSet has not finalizer and has not deletion timestamp, then add finalizer and requeue
	if !utils.HasFinalizer(&sampleSet.ObjectMeta, GetSampleSetFinalizer(req.Name)) &&
		!utils.HasDeletionTimestamp(&sampleSet.ObjectMeta) {
		r.Log.V(1).Info("add finalizer successfully")
		return r.AddFinalizer(ctx, sampleSet)
	}

	// 3. Get driver and construct reconcile context
	var driverName v1alpha1.DriverName
	if sampleSet.Spec.CSI == nil {
		driverName = driver.DefaultDriver
	} else {
		driverName = sampleSet.Spec.CSI.Driver
	}
	CSIDriver, err := driver.GetDriver(driverName)
	if err != nil {
		r.Log.Error(err, "get driver error")
		r.Recorder.Event(sampleSet, v1.EventTypeWarning,
			common.ErrorDriverNotExist, err.Error())
		return utils.NoRequeue()
	}
	RCtx := &common.ReconcileContext{
		Ctx:      ctx,
		Client:   r.Client,
		Req:      &req,
		Log:      r.Log,
		Scheme:   r.Scheme,
		Recorder: r.Recorder,
	}
	// 4. construct SampleSet Controller
	ssc := NewSampleSetController(sampleSet, CSIDriver, RCtx)
	return ssc.reconcilePhase()
}

// AddFinalizer add finalizer to SampleSet
func (r *SampleSetReconciler) AddFinalizer(ctx context.Context, sampleSet *v1alpha1.SampleSet) (ctrl.Result, error) {
	sampleSetFinalizer := GetSampleSetFinalizer(sampleSet.Name)
	sampleSet.Finalizers = append(sampleSet.Finalizers, sampleSetFinalizer)
	if err := r.Update(ctx, sampleSet); err != nil {
		return utils.RequeueWithError(err)
	}
	return utils.NoRequeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *SampleSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1.Event{},
		common.IndexerKeyEvent,
		EventIndexerFunc); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1.Pod{},
		common.IndexerKeyRuntime,
		RuntimePodIndexerFunc); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&batchv1.PaddleJob{},
		common.IndexerKeyPaddleJob,
		PaddleJobIndexerFunc); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.SampleSet{}).
		Owns(&v1.Event{}).
		Owns(&v1.Pod{}).
		Complete(r)
}

type SampleSetController struct {
	Controller
	SampleSet *v1alpha1.SampleSet
}

func NewSampleSetController(
	sampleSet *v1alpha1.SampleSet,
	CSIDriver driver.Driver,
	ctx *common.ReconcileContext) *SampleSetController {
	return &SampleSetController{
		Controller: Controller{
			Driver:           CSIDriver,
			Sample:           sampleSet,
			ReconcileContext: ctx,
		},
		SampleSet: sampleSet,
	}
}

func (s *SampleSetController) reconcilePhase() (ctrl.Result, error) {
	// if SampleSet has deletion timestamp the delete all resource create by this controller
	if utils.HasDeletionTimestamp(&s.SampleSet.ObjectMeta) {
		return s.deleteSampleSet()
	}

	// Reconcile the phase of SampleSet from None to Ready
	switch s.SampleSet.Status.Phase {
	case common.SampleSetNone:
		return s.reconcileNone()
	case common.SampleSetBound:
		return s.reconcileBound()
	case common.SampleSetMount:
		return s.reconcileMount()
	case common.SampleSetSyncing:
		return s.reconcileSyncing()
	case common.SampleSetSyncFailed:
		return s.reconcileSyncFailed()
	case common.SampleSetPartialReady:
		return s.reconcilePartialReady()
	case common.SampleSetReady:
		return s.reconcileReady()
	}
	s.Log.Error(fmt.Errorf("phase %s not support", s.SampleSet.Status.Phase), "")
	return utils.NoRequeue()
}

// reconcileNone After user create SampleSet CR then create PV and PVC automatically
func (s *SampleSetController) reconcileNone() (ctrl.Result, error) {
	s.Log.Info("==== reconcileNone ====")

	// 1. check if pv and pvc is already exist
	pv := &v1.PersistentVolume{}
	pvExists, err := s.ResourcesExistWithObject(pv, PV)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	pvcExists, err := s.ResourcesExist(PVC)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	// if pv and pvc is already create by paddle-operator,
	if pvcExists && pvExists {
		//  wait util pv phase is Bound
		if pv.Status.Phase != v1.VolumeBound {
			s.Log.Info("requeue and wait pv bound")
			return utils.RequeueAfter(1 * time.Second)
		}
		// if pv is bounded then update the phase of SampleSet to Bound
		s.SampleSet.Status.Phase = common.SampleSetBound
		err = s.UpdateResourceStatus(s.SampleSet, SampleSet)
		if err != nil {
			return utils.RequeueWithError(err)
		}
		return utils.NoRequeue()
	}

	// 2. check if secret name is none or secret not exist
	if s.SampleSet.Spec.SecretRef == nil {
		e := errors.New("spec.secretRef should not be empty")
		s.Log.Error(e, "spec.secretRef is nil")
		s.Recorder.Event(s.SampleSet, v1.EventTypeWarning, common.ErrorSecretNotExist, e.Error())
		return utils.NoRequeue()
	}
	if s.SampleSet.Spec.SecretRef.Name == "" {
		err := errors.New("spec.secretRef.name is not set")
		s.Log.Error(err, "spec.secretRef.name is empty string")
		s.Recorder.Event(s.SampleSet, v1.EventTypeWarning, common.ErrorSecretNotExist, err.Error())
		return utils.NoRequeue()
	}
	exist, err := s.ResourcesExist(Secret)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	if !exist {
		err := fmt.Errorf("secret %s is not exist", s.SampleSet.Spec.SecretRef.Name)
		s.Log.Error(err, "please create secret object and named it as spec.secretRef.name")
		s.Recorder.Event(s.SampleSet, v1.EventTypeWarning, common.ErrorSecretNotExist, err.Error())
		return utils.NoRequeue()
	}
	// 3. create persistent volume and set its name as the SampleSet
	if err := s.CreateResource(PV); err != nil {
		return utils.RequeueWithError(err)
	}
	// 4. create persistent volume claim and set its name as the SampleSet
	if err := s.CreateResource(PVC); err != nil {
		return utils.RequeueWithError(err)
	}

	return utils.RequeueAfter(1 * time.Second)
}

// reconcileBound After create PV/PVC then create runtime StatefulSet
func (s *SampleSetController) reconcileBound() (ctrl.Result, error) {
	s.Log.Info("==== reconcileBound ====")

	// 1. check if Service and StatefulSet is already exist
	serviceExists, err := s.ResourcesExist(Service)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	statefulSet := &appv1.StatefulSet{}
	runtimeExists, err := s.ResourcesExistWithObject(statefulSet, StatefulSet)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	// if service and runtime StatefulSet is created by paddle-operator
	if serviceExists && runtimeExists {
		// wait util at least one replica ready and update the phase of SampleSet to Mount
		if statefulSet.Status.ReadyReplicas > 0 {
			s.SampleSet.Status.Phase = common.SampleSetMount
			runtimeStatus := &v1alpha1.RuntimeStatus{
				SpecReplicas:  s.SampleSet.Spec.Partitions,
				ReadyReplicas: statefulSet.Status.ReadyReplicas,
				RuntimeReady: fmt.Sprintf("%d/%d",
					statefulSet.Status.ReadyReplicas,
					s.SampleSet.Spec.Partitions),
			}
			s.SampleSet.Status.RuntimeStatus = runtimeStatus
			err = s.UpdateResourceStatus(s.SampleSet, SampleSet)
			if err != nil {
				return utils.RequeueWithError(err)
			}
			return utils.NoRequeue()
		}

		// check if the pod created by StatefulSet is work properly
		eventList := &v1.EventList{}
		err = s.ListResources(eventList, Event)
		if err != nil {
			return utils.RequeueWithError(err)
		}
		eventLen := len(eventList.Items)
		if eventLen == 0 {
			s.Log.V(1).Info("wait at least one replica ready")
			return utils.RequeueAfter(5 * time.Second)
		}

		// if the pod created by StatefulSet is not work properly then recorder event and requeue
		event := eventList.Items[0]
		s.Recorder.Event(s.SampleSet, v1.EventTypeWarning, event.Reason, event.Message)
		return utils.RequeueWithError(errors.New(event.Message))
	}
	// 2. create service
	if err := s.CreateResource(Service); err != nil {
		return utils.RequeueWithError(err)
	}
	// 3. create runtime StatefulSet
	if err := s.CreateResource(StatefulSet); err != nil {
		return utils.RequeueWithError(err)
	}
	return utils.RequeueAfter(10 * time.Second)
}

// reconcileMount After create runtime daemon set and mounted, before sync data job done
func (s *SampleSetController) reconcileMount() (ctrl.Result, error) {
	s.Log.Info("==== reconcileMount ====")

	// 1. upload syncJobOptions to runtime server and trigger it to do sync data job
	syncJobName := s.SampleSet.Status.JobsName.SyncJobName
	if !s.SampleSet.Spec.NoSync && syncJobName == "" {
		syncJobName = uuid.NewUUID()
		// post sync job options to runtime server 0
		err := s.PostJobOptions(syncJobName, SyncJob)
		if err != nil {
			if errors.Is(err, optionError) {
				return utils.NoRequeue()
			}
			return utils.RequeueWithError(err)
		}
		// Add jobsName to SampleSet
		s.SampleSet.Status.JobsName.SyncJobName = syncJobName
		s.SampleSet.Status.Phase = common.SampleSetSyncing
		// update SampleSet Status
		if err := s.UpdateResourceStatus(s.SampleSet, SampleSet); err != nil {
			return utils.RequeueWithError(err)
		}
		return utils.NoRequeue()
	}

	// 2. wait util the first runtime server produce cache info file
	status, err := s.CollectCacheStatusByIndex(0)
	if err != nil {
		return utils.RequeueAfter(10 * time.Second)
	}
	if status != nil {
		s.SampleSet.Status.CacheStatus = status
	}

	// 3. update SampleSet phase to partial ready
	s.SampleSet.Status.Phase = common.SampleSetPartialReady
	if err := s.UpdateResourceStatus(s.SampleSet, SampleSet); err != nil {
		return utils.RequeueWithError(err)
	}
	return utils.NoRequeue()
}

// reconcileSyncing wait the sync data job done and return to mount phase
func (s *SampleSetController) reconcileSyncing() (ctrl.Result, error) {
	s.Log.Info("==== reconcileSyncing ====")

	time.Sleep(5 * time.Second)
	// 1. get cache status from the first runtime server
	newStatus, err := s.CollectCacheStatusByIndex(0)
	if err != nil {
		return utils.RequeueAfter(5 * time.Second)
	}
	if newStatus != nil && !reflect.DeepEqual(newStatus, s.SampleSet.Status.CacheStatus) {
		s.SampleSet.Status.CacheStatus = newStatus
		err = s.UpdateResourceStatus(s.SampleSet, SampleSet)
		if err != nil {
			return utils.RequeueWithError(err)
		}
		return utils.NoRequeue()
	}

	// 2. get the result of sync data job
	filename := s.SampleSet.Status.JobsName.SyncJobName
	result, err := s.GetJobResult(filename, SyncJob)
	if err != nil {
		return utils.RequeueWithError(err)
	}

	// 3. if sync job status is running then wait seconds and requeue
	if result != nil && result.Status == common.JobStatusRunning {
		s.Log.Info("wait util sync job done")
		return utils.RequeueAfter(5 * time.Second)
	}
	// 4. if sync job status is failed, then update phase to SyncFailed
	if result != nil && result.Status == common.JobStatusFail {
		e := errors.New(result.Message)
		s.SampleSet.Status.Phase = common.SampleSetSyncFailed
		s.Log.Error(e, "sync job error", "jobName", filename)
		s.Recorder.Event(s.SampleSet, v1.EventTypeWarning, SyncJob.ErrorDoJob(), e.Error())
	}
	// 5. if sync job status is success, then return phase to mount
	if result == nil {
		s.SampleSet.Status.Phase = common.SampleSetMount
	}

	// 6. update SampleSet phase
	if err := s.UpdateResourceStatus(s.SampleSet, SampleSet); err != nil {
		return utils.RequeueWithError(err)
	}
	s.Log.Info("return SampleSet phase to mount")
	return utils.NoRequeue()
}

// reconcileReady reconcile
func (s *SampleSetController) reconcileReady() (ctrl.Result, error) {
	s.Log.Info("==== reconcileReady ====")

	// 1. check whether spec.partitions is changed, if partitions is changed,
	// update the replicas of StatefulSet and update SampleSet phase to partial ready.
	statefulSet := &appv1.StatefulSet{}
	err := s.GetResource(statefulSet, StatefulSet)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	specReplicas := *statefulSet.Spec.Replicas
	readyReplicas := statefulSet.Status.ReadyReplicas
	partitions := s.SampleSet.Spec.Partitions

	// if runtime server ready replicas not equal SampleSet spec partitions
	if specReplicas != partitions || readyReplicas != partitions {
		// update phase to partial ready
		s.SampleSet.Status.Phase = common.SampleSetPartialReady
		err = s.UpdateResourceStatus(s.SampleSet, SampleSet)
		if err != nil {
			return utils.RequeueWithError(err)
		}
		s.Log.Info("update sampleset phase to partial ready")
		return utils.NoRequeue()
	}

	// 2. get the cache status from runtime servers
	newStatus, err := s.CollectCacheStatusByPartitions(int(partitions))
	if err != nil || newStatus == nil {
		return utils.RequeueAfter(common.RuntimeCacheInterval*time.Second + 1)
	}
	if newStatus != nil && !reflect.DeepEqual(newStatus, s.SampleSet.Status.CacheStatus) {
		s.SampleSet.Status.CacheStatus = newStatus
		err = s.UpdateResourceStatus(s.SampleSet, SampleSet)
		if err != nil {
			return utils.RequeueWithError(err)
		}
		s.Log.Info("updated sampleset cache status")
		return utils.RequeueAfter(common.RuntimeCacheInterval*time.Second + 1)
	}

	// 3. if all sample data has been cached in local disk
	if newStatus.TotalSize == newStatus.CachedSize {
		return utils.NoRequeue()
	}

	// 4. if there are some PaddleJob is in running phase, requeue and keep update cache status
	pdjList := &batchv1.PaddleJobList{}
	err = s.ListResources(pdjList, PaddleJob)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	if len(pdjList.Items) > 0 {
		return utils.RequeueAfter(2*common.RuntimeCacheInterval*time.Second + 2)
	}

	return utils.NoRequeue()
}

func (s *SampleSetController) deleteSampleSet() (ctrl.Result, error) {
	s.Log.Info("==== deleteSampleSet ====")

	sampleSetName := s.SampleSet.Name
	label := s.GetLabel(sampleSetName)

	// 1. wait util all PaddleJob is finish

	// 2. delete label of cache nodes
	nodeList := &v1.NodeList{}
	if err := s.ListResources(nodeList, Node); err != nil {
		return utils.RequeueWithError(err)
	}
	for _, node := range nodeList.Items {
		delete(node.Labels, label)
		if err := s.UpdateResource(&node, Node); err != nil {
			return utils.RequeueWithError(err)
		}
		s.Log.Info("remove label from node success", "node", node.Name)
	}
	// 3. delete runtime StatefulSet
	if err := s.DeleteResource(StatefulSet); err != nil {
		return utils.RequeueWithError(err)
	}
	// 4. wait all StatefulSet replicas deleted
	podList := &v1.PodList{}
	if err := s.ListResources(podList, RuntimePod); err != nil {
		return utils.RequeueWithError(err)
	}
	if len(podList.Items) > 0 {
		s.Log.Info("wait all statefulset replicas deleted")
		return utils.RequeueAfter(5 * time.Second)
	}
	// 5. delete runtime service
	if err := s.DeleteResource(Service); err != nil {
		return utils.RequeueWithError(err)
	}
	// 6. delete pvc
	if err := s.DeleteResource(PVC); err != nil {
		return utils.RequeueWithError(err)
	}
	// 7. delete pv
	if err := s.DeleteResource(PV); err != nil {
		return utils.RequeueWithError(err)
	}
	// 8. remove SampleSet finalizer and update SampleSet
	sampleSetFinalizer := GetSampleSetFinalizer(sampleSetName)
	if utils.HasFinalizer(&s.SampleSet.ObjectMeta, sampleSetFinalizer) {
		utils.RemoveFinalizer(&s.SampleSet.ObjectMeta, sampleSetFinalizer)
	}
	if err := s.UpdateResource(s.SampleSet, SampleSet); err != nil {
		return utils.RequeueWithError(err)
	}

	s.Log.Info("==== deleted all resource ====")
	return utils.NoRequeue()
}

func (s *SampleSetController) reconcileSyncFailed() (ctrl.Result, error) {
	s.Log.Info("==== reconcileSyncFailed ====")

	// 1. if spec.noSync update to true, then return phase to mount
	if s.SampleSet.Spec.NoSync {
		s.SampleSet.Status.Phase = common.SampleSetMount
		err := s.UpdateResourceStatus(s.SampleSet, SampleSet)
		if err != nil {
			return utils.RequeueWithError(err)
		}
		s.Log.Info("noSync is true, return SampleSet phase to mount")
		return utils.NoRequeue()
	}
	runtimeName := s.GetRuntimeName(s.Req.Name)
	serviceName := s.GetServiceName(s.Req.Name)
	filename := s.SampleSet.Status.JobsName.SyncJobName
	baseUri := utils.GetBaseUriByIndex(runtimeName, serviceName, 0)

	// 2. get syncJobOptions from runtime server
	oldOptions := &v1alpha1.SyncJobOptions{}
	err := utils.GetJobOption(oldOptions, filename, baseUri, common.PathSyncOptions)
	if err != nil {
		e := fmt.Errorf("get sync job option error: %s", err.Error())
		return utils.RequeueWithError(e)
	}

	// 3. create syncJobOptions for check if SampleSet is update by user
	newOptions := &v1alpha1.SyncJobOptions{}
	err = s.CreateJobOptions(newOptions, SyncJob)
	if err != nil {
		if errors.Is(err, optionError) {
			return utils.NoRequeue()
		}
		return utils.RequeueWithError(err)
	}

	// 4. if syncJobOptions have not updated, then no requeue
	if reflect.DeepEqual(oldOptions, newOptions) {
		s.Log.Info("syncJobOptions has not update")
		return utils.NoRequeue()
	}

	// 5. if SampleSet updated by user and syncJobOptions has changed,
	// then delete old syncJobName and return SampleSet phase to mount.
	s.SampleSet.Status.JobsName.SyncJobName = ""
	s.SampleSet.Status.Phase = common.SampleSetMount
	err = s.UpdateResourceStatus(s.SampleSet, SampleSet)
	if err != nil {
		return utils.RequeueWithError(err)
	}
	s.Log.Info("syncJobOptions changed, return SampleSet phase to mount")
	return utils.NoRequeue()
}

func (s *SampleSetController) reconcilePartialReady() (ctrl.Result, error) {
	s.Log.Info("==== reconcilePartialReady ====")

	label := s.GetLabel(s.Req.Name)

	// 1. get runtime StatefulSet
	statefulSet := &appv1.StatefulSet{}
	if err := s.GetResource(statefulSet, StatefulSet); err != nil {
		return utils.RequeueWithError(err)
	}

	// 2. list runtime server pods
	podList := &v1.PodList{}
	if err := s.ListResources(podList, RuntimePod); err != nil {
		return utils.RequeueWithError(err)
	}
	var runtimePodNames []string
	nodePodMap := make(map[string]string)
	for _, pod := range podList.Items {
		nodeName := pod.Spec.NodeName
		nodePodMap[nodeName] = pod.Name
		runtimePodNames = append(runtimePodNames, pod.Name)
	}

	// 3. update nodes label
	for nodeName, value := range nodePodMap {
		node := &v1.Node{}
		key := client.ObjectKey{Name: nodeName}
		if err := s.Get(s.Ctx, key, node); err != nil {
			e := fmt.Errorf("get node %s error: %s", nodeName, err.Error())
			return utils.RequeueWithError(e)
		}
		// if label is exist and the value is equal to the name of pod, continue
		if v, exist := node.Labels[label]; exist && v == value {
			continue
		}
		// update the label of node
		node.Labels[label] = value
		if err := s.UpdateResource(node, Node); err != nil {
			return utils.RequeueWithError(err)
		}
		s.Log.Info("label node successful", "name", nodeName, "label", value)
	}

	// 4. list nodes with label
	nodeList := &v1.NodeList{}
	if err := s.ListResources(nodeList, Node); err != nil {
		return utils.RequeueWithError(err)
	}

	// 5. remove the label of nodes with terminate runtime pod
	for _, node := range nodeList.Items {
		labelValue := node.Labels[label]
		podName, exist := nodePodMap[node.Name]
		if exist && labelValue == podName {
			continue
		}
		delete(node.Labels, label)
		if err := s.UpdateResource(&node, Node); err != nil {
			return utils.RequeueWithError(err)
		}
		s.Log.Info("remove label from node success", "node", node.Name)
	}

	needUpdateStatefulSet := false

	// 7. collect all cache status from running runtime server
	status, err := s.CollectCacheStatus(runtimePodNames)
	if err != nil {
		return utils.RequeueAfter(5 * time.Second)
	}
	if status != nil && !reflect.DeepEqual(status, s.SampleSet.Status.CacheStatus) {
		needUpdateStatefulSet = true
		s.SampleSet.Status.CacheStatus = status
		s.Log.Info("cache status has changed")
	}

	// 8. update StatefulSet if spec replicas is not equal to the partitions of SampleSet
	specReplicas := *statefulSet.Spec.Replicas
	partitions := s.SampleSet.Spec.Partitions
	if specReplicas != partitions {
		statefulSet.Spec.Replicas = &partitions
		if err := s.UpdateResource(statefulSet, StatefulSet); err != nil {
			return utils.RequeueWithError(err)
		}
	}

	// 9. update SampleSet RuntimeStatus and phase
	newReadyReplicas := statefulSet.Status.ReadyReplicas
	oldSpecReplicas := s.SampleSet.Status.RuntimeStatus.SpecReplicas
	oldReadyReplicas := s.SampleSet.Status.RuntimeStatus.ReadyReplicas
	if oldSpecReplicas != partitions || oldReadyReplicas != newReadyReplicas {
		needUpdateStatefulSet = true
		s.SampleSet.Status.RuntimeStatus.SpecReplicas = partitions
		s.SampleSet.Status.RuntimeStatus.ReadyReplicas = newReadyReplicas
		s.SampleSet.Status.RuntimeStatus.RuntimeReady = fmt.Sprintf(
			"%d/%d", newReadyReplicas, partitions)
	}

	allReady := partitions == specReplicas
	allReady = allReady && specReplicas == newReadyReplicas
	allReady = allReady && len(nodePodMap) == int(newReadyReplicas)
	if allReady {
		s.SampleSet.Status.Phase = common.SampleSetReady
	}

	if needUpdateStatefulSet || allReady {
		if err := s.UpdateResourceStatus(s.SampleSet, SampleSet); err != nil {
			return utils.RequeueWithError(err)
		}
		s.Log.Info("update sampleset status")
		return utils.NoRequeue()
	}
	s.Log.Info("wait all runtime server ready")
	return utils.RequeueAfter(10 * time.Second)
}
