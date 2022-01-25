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
	"fmt"
	"sort"
	"strconv"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "github.com/paddleflow/paddle-operator/api/v1"
	"github.com/paddleflow/paddle-operator/api/v1alpha1"
	"github.com/paddleflow/paddle-operator/controllers/extensions/common"
	"github.com/paddleflow/paddle-operator/controllers/extensions/driver"
	"github.com/paddleflow/paddle-operator/controllers/extensions/utils"
)

var (
	PV           *Resource
	PVC          *Resource
	Node         *Resource
	Secret       *Resource
	Event        *Resource
	Service      *Resource
	SampleSet    *Resource
	SampleJob    *Resource
	PaddleJob    *Resource
	StatefulSet  *Resource
	RuntimePod   *Resource
	SourceSecret *Resource

	RmrJob    *JobType
	ClearJob  *JobType
	SyncJob   *JobType
	WarmupJob *JobType
	Terminate *JobType

	optionError *OptionError
	JobTypeMap  map[v1alpha1.SampleJobType]*JobType
)

func init() {
	optionError = &OptionError{}

	// get create delete exist
	PV = NewResource("PV")
	PV.WithLabel = true
	PV.Object = PVObject
	PV.ObjectKey = NameObjectKey
	PV.CreateObject = PVCreateObject

	// get create delete exists
	PVC = NewResource("PVC")
	PVC.WithLabel = true
	PVC.Object = PVCObject
	PVC.ObjectKey = NamespacedObjectKey
	PVC.CreateObject = PVCCreateObject

	// list
	Node = NewResource("Node")
	Node.WithLabel = true
	Node.ListOptions = NodeListOptions

	// get exists
	Secret = NewResource("Secret")
	Secret.Object = SecretObject
	Secret.ObjectKey = SecretObjectKey

	// get exists
	SourceSecret = NewResource("SourceSecret")
	SourceSecret.Object = SecretObject
	SourceSecret.ObjectKey = SourceSecretObjectKey

	// list
	Event = NewResource("Event")
	Event.ListOptions = EventListOptions

	// get create delete exists
	Service = NewResource("Service")
	Service.WithLabel = true
	Service.Object = ServiceObject
	Service.ObjectKey = ServiceObjectKey
	Service.CreateObject = ServiceCreateObject

	// get delete update exists
	SampleSet = NewResource("SampleSet")
	SampleSet.Object = SampleSetObject
	SampleSet.ObjectKey = NamespacedObjectKey

	// get delete update exists
	SampleJob = NewResource("SampleJob")
	SampleJob.Object = SampleJobObject
	SampleJob.ObjectKey = SampleJobObjectKey

	// get create delete update exists
	StatefulSet = NewResource("StatefulSet")
	StatefulSet.WithLabel = true
	StatefulSet.Object = StatefulSetObject
	StatefulSet.ObjectKey = StatefulSetObjectKey
	StatefulSet.CreateObject = StatefulSetCreateObject

	// list
	RuntimePod = NewResource("Pod")
	RuntimePod.WithLabel = true
	RuntimePod.ListOptions = RuntimePodListOptions

	// list
	PaddleJob = NewResource("PaddleJob")
	PaddleJob.ListOptions = PaddleJobListOptions

	// sync job options
	SyncJob = NewJobOptions("SyncJob")
	SyncJob.Options = SyncOptions
	SyncJob.BaseUris = FirstBaseUri
	SyncJob.CreateOptions = SyncCreateOptions
	SyncJob.OptionPath = common.PathSyncOptions
	SyncJob.ResultPath = common.PathSyncResult

	// warmup job options
	WarmupJob = NewJobOptions("WarmupJob")
	WarmupJob.Options = WarmupOptions
	WarmupJob.BaseUris = AllBaseUris
	WarmupJob.CreateOptions = WarmupCreateOptions
	WarmupJob.OptionPath = common.PathWarmupOptions
	WarmupJob.ResultPath = common.PathWarmupResult

	// clear job options
	ClearJob = NewJobOptions("ClearJob")
	ClearJob.Options = ClearOptions
	ClearJob.BaseUris = AllBaseUris
	ClearJob.CreateOptions = ClearCreateOptions
	ClearJob.OptionPath = common.PathClearOptions
	ClearJob.ResultPath = common.PathClearResult

	// rmr job option
	RmrJob = NewJobOptions("RmrJob")
	RmrJob.Options = RmrOptions
	RmrJob.BaseUris = FirstBaseUri
	RmrJob.CreateOptions = RmrCreateOptions
	RmrJob.OptionPath = common.PathRmrOptions
	RmrJob.ResultPath = common.PathRmrResult

	// Terminate
	Terminate = NewJobOptions("Terminate")
	Terminate.Options = TerminateOptions
	Terminate.BaseUris = AllBaseUris
	Terminate.CreateOptions = TerminateCreateOptions
	Terminate.OptionPath = common.PathClearOptions
	Terminate.ResultPath = common.PathClearResult

	// dependents
	PV.Dependents = []*Resource{Secret, SampleSet}
	ClearJob.Dependents = []*Resource{StatefulSet}
	SyncJob.Dependents = []*Resource{SourceSecret, SampleSet}
	StatefulSet.Dependents = []*Resource{PV, SampleSet}
	RmrJob.Dependents = []*Resource{SampleJob}
	WarmupJob.Dependents = []*Resource{SampleSet}

	JobTypeMap = map[v1alpha1.SampleJobType]*JobType{
		common.JobTypeSync:   SyncJob,
		common.JobTypeWarmup: WarmupJob,
		common.JobTypeClear:  ClearJob,
		common.JobTypeRmr:    RmrJob,
	}
}

type OptionError struct{}

func (e *OptionError) Error() string { return "option error" }

type Dependence interface {
	GetName() string
	GetDependents() []*Resource
}

type Resource struct {
	Name       string
	WithLabel  bool
	Dependents []*Resource

	Object       func() client.Object
	ObjectKey    func(c *Controller) client.ObjectKey
	ListOptions  func(c *Controller) []client.ListOption
	CreateObject func(c *Controller, object client.Object, ctx *common.RequestContext) error
}

func NewResource(name string) *Resource {
	return &Resource{Name: name}
}

func (r *Resource) GetName() string {
	return r.Name
}

func (r *Resource) GetDependents() []*Resource {
	if r.Dependents == nil {
		return []*Resource{}
	}
	return r.Dependents
}

func (r *Resource) ErrorAlreadyExist() string {
	return "Error" + r.Name + "AlreadyExist"
}

func (r *Resource) ErrorCreateObject() string {
	return "ErrorCreate" + r.Name
}

func (r *Resource) ErrorDeleteObject() string {
	return "ErrorDelete" + r.Name
}

func (r *Resource) CreateSuccessfully() string {
	return "Create" + r.Name + "Successfully"
}

func (r *Resource) UpdateSuccessfully() string {
	return "Update" + r.Name + "Successfully"
}

// Object func() client.Object

func PVObject() client.Object { return &v1.PersistentVolume{} }

func PVCObject() client.Object { return &v1.PersistentVolumeClaim{} }

func SecretObject() client.Object { return &v1.Secret{} }

func ServiceObject() client.Object { return &v1.Service{} }

func StatefulSetObject() client.Object { return &appv1.StatefulSet{} }

func SampleSetObject() client.Object { return &v1alpha1.SampleSet{} }

func SampleJobObject() client.Object { return &v1alpha1.SampleJob{} }

// ObjectKey func(c *Controller) client.ObjectKey

func NameObjectKey(c *Controller) client.ObjectKey {
	return client.ObjectKey{Name: c.Req.Name}
}

func NamespacedObjectKey(c *Controller) client.ObjectKey {
	return c.Req.NamespacedName
}

func SecretObjectKey(c *Controller) client.ObjectKey {
	if sampleSet, ok := c.Sample.(*v1alpha1.SampleSet); ok {
		name := sampleSet.Spec.SecretRef.Name
		namespace := c.Req.Namespace
		if sampleSet.Spec.SecretRef.Namespace != "" {
			namespace = sampleSet.Spec.SecretRef.Namespace
		}
		return client.ObjectKey{Name: name, Namespace: namespace}
	}
	panic(fmt.Errorf("%s is not register in SecretObjectKey", c.Sample.GetObjectKind().GroupVersionKind().String()))
}

func SourceSecretObjectKey(c *Controller) client.ObjectKey {
	// under SampleSet controller
	if sampleSet, ok := c.Sample.(*v1alpha1.SampleSet); ok {
		name := sampleSet.Spec.SecretRef.Name
		namespace := c.Req.Namespace
		if sampleSet.Spec.SecretRef.Namespace != "" {
			namespace = sampleSet.Spec.SecretRef.Namespace
		}
		// if source secret is not nil then use it
		if sampleSet.Spec.Source != nil && sampleSet.Spec.Source.SecretRef != nil {
			if sampleSet.Spec.Source.SecretRef.Name != "" {
				name = sampleSet.Spec.Source.SecretRef.Name
			}
			if sampleSet.Spec.Source.SecretRef.Namespace != "" {
				namespace = sampleSet.Spec.Source.SecretRef.Namespace
			}
		}
		return client.ObjectKey{Name: name, Namespace: namespace}
	}
	// under SampleJob Controller
	if sampleJob, ok := c.Sample.(*v1alpha1.SampleJob); ok {
		name := sampleJob.Spec.SecretRef.Name
		namespace := c.Req.Namespace
		if sampleJob.Spec.SecretRef.Namespace != "" {
			namespace = sampleJob.Spec.SecretRef.Namespace
		}
		return client.ObjectKey{Name: name, Namespace: namespace}
	}
	panic(fmt.Errorf("%s is not register in SourceSecretObjectKey", c.Sample.GetObjectKind().GroupVersionKind().String()))
}

func ServiceObjectKey(c *Controller) client.ObjectKey {
	name := c.GetServiceName(c.Req.Name)
	return client.ObjectKey{Name: name, Namespace: c.Req.Namespace}
}

func StatefulSetObjectKey(c *Controller) client.ObjectKey {
	name := c.GetRuntimeName(c.Req.Name)
	return client.ObjectKey{Name: name, Namespace: c.Req.Namespace}
}

func SampleJobObjectKey(c *Controller) client.ObjectKey {
	sampleJob := c.Sample.(*v1alpha1.SampleJob)
	return client.ObjectKey{Name: sampleJob.Name, Namespace: sampleJob.Namespace}
}

// func(c *Controller, object client.Object, ctx *common.RequestContext) error

func PVCreateObject(c *Controller, object client.Object, ctx *common.RequestContext) error {
	pv := object.(*v1.PersistentVolume)
	if err := c.CreatePV(pv, ctx); err != nil {
		return err
	}
	return nil
}

func PVCCreateObject(c *Controller, object client.Object, ctx *common.RequestContext) error {
	pvc := object.(*v1.PersistentVolumeClaim)
	if err := c.CreatePVC(pvc, ctx); err != nil {
		return err
	}
	return nil
}

func ServiceCreateObject(c *Controller, object client.Object, ctx *common.RequestContext) error {
	service := object.(*v1.Service)
	if err := c.CreateService(service, ctx); err != nil {
		return err
	}
	return nil
}

func StatefulSetCreateObject(c *Controller, object client.Object, ctx *common.RequestContext) error {
	statefulSet := object.(*appv1.StatefulSet)
	if err := c.CreateRuntime(statefulSet, ctx); err != nil {
		return err
	}
	return nil
}

// ListOptions func(c *Controller) []client.ListOption

func EventListOptions(c *Controller) []client.ListOption {
	lOpt := client.Limit(1)
	runtimeName := c.GetRuntimeName(c.Req.Name)
	nOpt := client.InNamespace(c.Req.Namespace)
	values := []string{"Pod", c.Req.Namespace, runtimeName + "-0", "Warning"}
	fOpt := client.MatchingFields{
		common.IndexerKeyEvent: strings.Join(values, "-"),
	}
	return []client.ListOption{fOpt, nOpt, lOpt}
}

func NodeListOptions(c *Controller) []client.ListOption {
	label := c.GetLabel(c.Req.Name)
	return []client.ListOption{client.HasLabels{label}}
}

func RuntimePodListOptions(c *Controller) []client.ListOption {
	label := c.GetLabel(c.Req.Name)
	runtimeName := c.GetRuntimeName(c.Req.Name)
	nOpt := client.InNamespace(c.Req.Namespace)
	lOpt := client.MatchingLabels(map[string]string{
		label: "true", "name": runtimeName,
	})
	fOpt := client.MatchingFields{
		common.IndexerKeyRuntime: string(v1.PodRunning),
	}
	return []client.ListOption{nOpt, lOpt, fOpt}
}

func PaddleJobListOptions(c *Controller) []client.ListOption {
	nOpt := client.InNamespace(c.Req.Namespace)
	values := []string{c.Req.Name, string(batchv1.Running)}
	fOpt := client.MatchingFields{
		common.IndexerKeyPaddleJob: strings.Join(values, "-"),
	}
	return []client.ListOption{nOpt, fOpt}
}

type JobType struct {
	Name       string
	OptionPath string
	ResultPath string
	Dependents []*Resource

	Options       func() interface{}
	BaseUris      func(c *Controller) ([]string, error)
	CreateOptions func(c *Controller, opt interface{}, ctx *common.RequestContext) error
}

func NewJobOptions(name string) *JobType {
	return &JobType{Name: name}
}

func (j *JobType) GetName() string {
	return j.Name
}

func (j *JobType) GetDependents() []*Resource {
	if j.Dependents == nil {
		return []*Resource{}
	}
	return j.Dependents
}

func (j *JobType) ErrorCreateJob() string {
	return "ErrorCreate" + string(j.Name)
}

func (j *JobType) ErrorDoJob() string {
	return "ErrorDo" + string(j.Name)
}

func (j *JobType) CreateSuccessfully() string {
	return "Create" + j.Name + "Successfully"
}

func (j *JobType) DoJobSuccessfully() string {
	return "Do" + j.Name + "Successfully"
}

// Options func() interface{}

func SyncOptions() interface{} { return &v1alpha1.SyncJobOptions{} }

func WarmupOptions() interface{} { return &v1alpha1.WarmupJobOptions{} }

func ClearOptions() interface{} { return &v1alpha1.ClearJobOptions{} }

func RmrOptions() interface{} { return &v1alpha1.RmrJobOptions{} }

func TerminateOptions() interface{} { return nil }

// CreateOptions func(c *Controller, opt *JobOptions, ctx *common.RequestContext) error

func SyncCreateOptions(c *Controller, opt interface{}, ctx *common.RequestContext) error {
	options := opt.(*v1alpha1.SyncJobOptions)
	err := c.CreateSyncJobOptions(options, ctx)
	if err != nil {
		return err
	}
	return nil
}

func RmrCreateOptions(c *Controller, opt interface{}, ctx *common.RequestContext) error {
	options := opt.(*v1alpha1.RmrJobOptions)
	err := c.CreateRmrJobOptions(options, ctx)
	if err != nil {
		return err
	}
	return nil
}

func WarmupCreateOptions(c *Controller, opt interface{}, ctx *common.RequestContext) error {
	options := opt.(*v1alpha1.WarmupJobOptions)
	err := c.CreateWarmupJobOptions(options, ctx)
	if err != nil {
		return err
	}
	return nil
}

func ClearCreateOptions(c *Controller, opt interface{}, ctx *common.RequestContext) error {
	options := opt.(*v1alpha1.ClearJobOptions)
	err := c.CreateClearJobOptions(options, ctx)
	if err != nil {
		return err
	}
	return nil
}

func TerminateCreateOptions(c *Controller, opt interface{}, ctx *common.RequestContext) error {
	return nil
}

// BaseUris func(c *Controller) ([]string, error)

func FirstBaseUri(c *Controller) ([]string, error) {
	runtimeName := c.GetRuntimeName(c.Req.Name)
	serviceName := c.GetServiceName(c.Req.Name)
	baseUri := utils.GetBaseUriByIndex(runtimeName, serviceName, 0)
	return []string{baseUri}, nil
}

func AllBaseUris(c *Controller) ([]string, error) {
	podList := &v1.PodList{}
	err := c.ListResources(podList, RuntimePod)
	if err != nil {
		return nil, err
	}
	// sort pod name with index
	var podSlice []common.PodNameIndex
	for _, pod := range podList.Items {
		nameSplit := strings.Split(pod.Name, "-")
		if len(nameSplit) == 0 {
			continue
		}
		indexStr := nameSplit[len(nameSplit)-1]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			continue
		}
		nameIndex := common.PodNameIndex{
			Name:  pod.Name,
			Index: index,
		}
		podSlice = append(podSlice, nameIndex)
	}
	sort.Slice(podSlice, func(i, j int) bool {
		return podSlice[i].Index < podSlice[j].Index
	})

	var baseUris []string
	serviceName := c.GetServiceName(c.Req.Name)
	for _, pod := range podSlice {
		baseUri := utils.GetBaseUriByName(pod.Name, serviceName)
		baseUris = append(baseUris, baseUri)
	}
	return baseUris, nil
}

type Controller struct {
	driver.Driver
	Sample client.Object
	*common.ReconcileContext
}

func (c *Controller) GetResource(object client.Object, r *Resource) error {
	if r.ObjectKey == nil {
		panic(fmt.Errorf("%s ObjectKey function not implement", r.Name))
	}
	key := r.ObjectKey(c)
	// get object by key
	if err := c.Get(c.Ctx, key, object); err != nil {
		// if error is not found, return it without wrapper
		if errors.IsNotFound(err) {
			return err
		}
		return fmt.Errorf("get %s %s error: %s", r.Name, key.String(), err.Error())
	}
	// if is not labeled resource, then return without check label
	if !r.WithLabel {
		c.Log.Info("get resource successful", "resource", r.Name, "name", key.String())
		return nil
	}
	// check if label in object, if not return already exist error
	label := c.GetLabel(c.Req.Name)
	if _, exits := object.GetLabels()[label]; !exits {
		err := fmt.Errorf("%s %s is is already exist, delete it and try again", r.Name, key.String())
		c.Recorder.Event(c.Sample, v1.EventTypeWarning, r.ErrorAlreadyExist(), err.Error())
	}
	c.Log.Info("get resource successfully", "name", key.String(), "resource", r.Name)
	return nil
}

func (c *Controller) ResourcesExist(r *Resource) (bool, error) {
	if r.Object == nil {
		panic(fmt.Errorf("%s Object function not implement", r.Name))
	}
	if err := c.GetResource(r.Object(), r); err != nil {
		if errors.IsNotFound(err) {
			c.Log.Info("resource not exist", "resource", r.Name, "name", r.ObjectKey(c).String())
			return false, nil
		}
		return false, err
	}
	c.Log.Info("resource exist", "resource", r.Name, "name", r.ObjectKey(c).String())
	return true, nil
}

func (c *Controller) ResourcesExistWithObject(object client.Object, r *Resource) (bool, error) {
	if r.Object == nil {
		panic(fmt.Errorf("%s Object function not implement", r.Name))
	}

	if err := c.GetResource(object, r); err != nil {
		if errors.IsNotFound(err) {
			c.Log.Info("resource not exist", "resource", r.Name,
				"name", r.ObjectKey(c).String())
			return false, nil
		}
		return false, err
	}
	c.Log.Info("resource exist", "resource", r.Name,
		"name", r.ObjectKey(c).String())
	return true, nil
}

func (c *Controller) ListResources(list client.ObjectList, r *Resource) error {
	if r.ListOptions == nil {
		panic(fmt.Errorf("%s ListOptions function not implement", r.Name))
	}
	opts := r.ListOptions(c)
	if err := c.List(c.Ctx, list, opts...); err != nil {
		return fmt.Errorf("list %s error: %s", r.Name, err.Error())
	}
	c.Log.Info("list successfully", "resource", r.Name)
	return nil
}

func (c *Controller) CreateResource(r *Resource) error {
	if r.CreateObject == nil {
		panic(fmt.Errorf("%s CreateObject function not implement", r.Name))
	}

	// if resource is already exist return with nil
	exist, err := c.ResourcesExist(r)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	ctx, err := c.GetRequestContext(r)
	if err != nil {
		return fmt.Errorf("create %s resources get request context error: %s", r.Name, err.Error())
	}

	object := r.Object()
	key := r.ObjectKey(c).String()
	if err := r.CreateObject(c, object, ctx); err != nil {
		e := fmt.Errorf("create %s %s error: %s", r.Name, key, err.Error())
		c.Recorder.Event(c.Sample, v1.EventTypeWarning, r.ErrorCreateObject(), e.Error())
		return e
	}
	if err := c.Create(c.Ctx, object); err != nil {
		e := fmt.Errorf("create %s %s error: %s", r.Name, key, err.Error())
		c.Recorder.Event(c.Sample, v1.EventTypeWarning, r.ErrorCreateObject(), e.Error())
		return e
	}
	c.Log.Info("create resource successfully", "name", key, "resource", r.Name)
	c.Recorder.Eventf(c.Sample, v1.EventTypeNormal, r.CreateSuccessfully(),
		"create %s %s successfully", r.Name, key)
	return nil
}

func (c *Controller) CreateJobOptions(opt interface{}, j *JobType) error {
	if j.CreateOptions == nil {
		panic(fmt.Errorf("%s CreateOptions function not implement", j.Name))
	}
	ctx, err := c.GetRequestContext(j)
	if err != nil {
		return fmt.Errorf("create %s options get request context error: %s", j.Name, err.Error())
	}

	if err := j.CreateOptions(c, opt, ctx); err != nil {
		e := fmt.Errorf("create %s error: %s", j.Name, err.Error())
		c.Log.Error(e, "please fix the job options and try again")
		c.Recorder.Event(c.Sample, v1.EventTypeWarning, j.ErrorCreateJob(), e.Error())
		return optionError
	}
	c.Log.Info("create job successfully", "jobName", j.Name,
		"options", fmt.Sprintf("%+v", opt))
	c.Recorder.Eventf(c.Sample, v1.EventTypeNormal, j.CreateSuccessfully(),
		"create %s successfully", j.Name)
	return nil
}

func (c *Controller) DeleteResource(r *Resource) error {
	if r.Object == nil {
		panic(fmt.Errorf("%s Object function not implement", r.Name))
	}

	object := r.Object()
	if err := c.GetResource(object, r); err != nil {
		return client.IgnoreNotFound(err)
	}

	key := r.ObjectKey(c).String()
	if err := c.Delete(c.Ctx, object); err != nil {
		e := fmt.Errorf("delete resource %s %s error: %s", r.Name, key, err.Error())
		c.Recorder.Event(c.Sample, v1.EventTypeWarning, r.ErrorDeleteObject(), e.Error())
		return e
	}
	c.Log.Info("delete resource successfully", "name", key, "resource", r.Name)
	return nil
}

func (c *Controller) UpdateResource(object client.Object, r *Resource) error {
	if err := c.Update(c.Ctx, object); err != nil {
		return fmt.Errorf("update %s %s error %s", r.Name, object.GetName(), err.Error())
	}
	c.Log.Info("update resource successfully", "name", object.GetName(), "resource", r.Name)
	return nil
}

func (c *Controller) UpdateResourceStatus(object client.Object, r *Resource) error {
	if err := c.Status().Update(c.Ctx, object); err != nil {
		return fmt.Errorf("update %s %s status error %s", r.Name, object.GetName(), err.Error())
	}
	c.Log.Info("update resource status successfully", "name", object.GetName(), "resource", r.Name)
	return nil
}

func (c *Controller) GetRequestContext(d Dependence) (*common.RequestContext, error) {
	ctx := &common.RequestContext{Req: c.Req}
	// add Sample Resource to request context
	switch c.Sample.(type) {
	case *v1alpha1.SampleSet:
		ctx.SampleSet = c.Sample.(*v1alpha1.SampleSet)
	case *v1alpha1.SampleJob:
		ctx.SampleJob = c.Sample.(*v1alpha1.SampleJob)
	default:
		panic(fmt.Errorf("%s is not register in sample", d.GetName()))
	}
	// add resource dependents in request context
	for _, resource := range d.GetDependents() {
		if resource.Object == nil {
			panic(fmt.Errorf("%s Object function not implement", resource.Name))
		}
		object := resource.Object()
		// SampleSet or SampleJob has been set then continue
		switch object.(type) {
		case *v1alpha1.SampleSet:
			if ctx.SampleSet != nil {
				continue
			}
		case *v1alpha1.SampleJob:
			if ctx.SampleSet != nil {
				continue
			}
		}
		// get object resource
		err := c.GetResource(object, resource)
		if err != nil {
			return nil, err
		}
		switch object.(type) {
		case *v1.PersistentVolume:
			ctx.PV = object.(*v1.PersistentVolume)
		case *v1.Secret:
			ctx.Secret = object.(*v1.Secret)
		case *v1.Service:
			ctx.Service = object.(*v1.Service)
		case *v1alpha1.SampleSet:
			ctx.SampleSet = object.(*v1alpha1.SampleSet)
		case *v1alpha1.SampleJob:
			ctx.SampleJob = object.(*v1alpha1.SampleJob)
		case *appv1.StatefulSet:
			ctx.StatefulSet = object.(*appv1.StatefulSet)
		default:
			panic(fmt.Errorf("%s is not register in request context", d.GetName()))
		}
	}
	c.Log.Info("get request context successful")
	return ctx, nil
}

func (c *Controller) PostJobOptionsWithParam(filename types.UID, j *JobType, param string) error {
	// 1. check if functions is implement in jobType
	if j.Options == nil {
		panic(fmt.Errorf("%s CreateOptions function not implement", j.Name))
	}
	if j.BaseUris == nil {
		panic(fmt.Errorf("%s BaseUris function not implement", j.Name))
	}
	if j.OptionPath == "" {
		panic(fmt.Errorf("%s OptionPath attribute has not been set", j.Name))
	}
	if j.ResultPath == "" {
		panic(fmt.Errorf("%s ResultPath attribute has not been set", j.Name))
	}
	// 2. create options of jobType
	opt := j.Options()
	err := c.CreateJobOptions(opt, j)
	if err != nil {
		return err
	}
	// 3. get base uris that need to post options
	baseUris, err := j.BaseUris(c)
	if err != nil {
		return err
	}
	// 4. upload options to runtime server after check job is not exist
	for _, baseUri := range baseUris {
		// if options has already upload to runtime
		err = utils.GetJobOption(j.Options(), filename, baseUri, j.OptionPath)
		if err == nil {
			c.Log.Info("options has already upload to server", "filename", filename,
				"baseUri", baseUri, "OptionPath", j.OptionPath)
			continue
		}
		err = utils.PostJobOption(opt, filename, baseUri, j.OptionPath, param)
		if err != nil {
			return err
		}
		c.Log.Info("upload options to server successfully", "filename", filename,
			"baseUri", baseUri, "OptionPath", j.OptionPath)
	}
	return nil
}

func (c *Controller) PostJobOptions(filename types.UID, j *JobType) error {
	return c.PostJobOptionsWithParam(filename, j, "")
}

func (c *Controller) PostTerminateSignal() error {
	param := "?" + common.TerminateSignal + "=true"
	return c.PostJobOptionsWithParam(common.TerminateSignal, Terminate, param)
}

func (c *Controller) GetJobResult(filename types.UID, j *JobType) (*common.JobResult, error) {
	// 1. check if functions is implement in jobType
	if j.BaseUris == nil {
		panic(fmt.Errorf("%s BaseUris function not implement", j.Name))
	}
	if j.ResultPath == "" {
		panic(fmt.Errorf("%s ResultPath attribute has not been set", j.Name))
	}
	// 2. get base uris that need request job result from
	baseUris, err := j.BaseUris(c)
	if err != nil {
		return nil, err
	}
	// 3. get job result from runtime server
	for _, baseUri := range baseUris {
		result, err := utils.GetJobResult(filename, baseUri, j.ResultPath)
		if err != nil {
			return nil, err
		}
		// if sync job status is success, job result nned not return
		if result.Status == common.JobStatusSuccess {
			continue
		}
		// if result status is running or fail return the result
		c.Log.Info("get job result successfully", "filename", filename,
			"baseUri", baseUri, "ResultPath", j.ResultPath)
		return result, nil
	}
	c.Log.Info("job is done successfully")
	c.Recorder.Eventf(c.Sample, v1.EventTypeNormal, j.DoJobSuccessfully(),
		"do %s successfully", j.Name)
	return nil, nil
}

func (c *Controller) CollectCacheStatus(podNames []string) (*v1alpha1.CacheStatus, error) {
	serviceName := c.GetServiceName(c.Req.Name)
	status, err := utils.CollectAllCacheStatus(podNames, serviceName)
	if err != nil {
		c.Log.Error(err, "get cache status error", "podNames", podNames)
		return nil, err
	}
	if status.ErrorMassage != "" {
		c.Log.Error(fmt.Errorf(status.ErrorMassage), "error massage from cache status")
	}
	// There some error when obtain the total size of mounted sample data
	if status.TotalSize == "0 B" {
		return nil, nil
	}
	c.Log.Info("collect cache status successfully", "podNames", podNames)
	return status, nil
}

func (c *Controller) CollectCacheStatusByIndex(index int) (*v1alpha1.CacheStatus, error) {
	runtimeName := c.GetRuntimeName(c.Req.Name)
	podName := fmt.Sprintf("%s-%d", runtimeName, index)
	return c.CollectCacheStatus([]string{podName})
}

func (c *Controller) CollectCacheStatusByPartitions(partitions int) (*v1alpha1.CacheStatus, error) {
	runtimeName := c.GetRuntimeName(c.Req.Name)
	var podNames []string
	for i := 0; i < partitions; i++ {
		podNames = append(podNames, fmt.Sprintf("%s-%d", runtimeName, i))
	}
	return c.CollectCacheStatus(podNames)
}

func GetSampleSetFinalizer(name string) string {
	return common.PaddleLabel + "/" + "sampleset-" + name
}

func GetSampleJobFinalizer(name string) string {
	return common.PaddleLabel + "/" + "samplejob-" + name
}

func EventIndexerFunc(obj client.Object) []string {
	event := obj.(*v1.Event)
	keys := []string{
		event.InvolvedObject.Kind,
		event.InvolvedObject.Namespace,
		event.InvolvedObject.Name,
		event.Type,
	}
	keyStr := strings.Join(keys, "-")
	return []string{keyStr}
}

func RuntimePodIndexerFunc(obj client.Object) []string {
	pod := obj.(*v1.Pod)
	return []string{string(pod.Status.Phase)}
}

func PaddleJobIndexerFunc(obj client.Object) []string {
	pdj := obj.(*batchv1.PaddleJob)
	sampleSetName := ""
	if pdj.Spec.SampleSetRef != nil {
		sampleSetName = pdj.Spec.SampleSetRef.Name
	}
	keys := []string{
		sampleSetName,
		string(pdj.Status.Phase),
	}
	keyStr := strings.Join(keys, "-")
	return []string{keyStr}
}
