package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	paddlev1alpha1 "github.com/PaddlePaddle/cloud/go/pkg/apis/paddlepaddle/v1alpha1"
	paddleclientset "github.com/PaddlePaddle/cloud/go/pkg/client/clientset/versioned"
	paddlescheme "github.com/PaddlePaddle/cloud/go/pkg/client/clientset/versioned/scheme"
	paddleinformers "github.com/PaddlePaddle/cloud/go/pkg/client/informers/externalversions"
	paddlelisters "github.com/PaddlePaddle/cloud/go/pkg/client/listers/paddlepaddle/v1alpha1"
)

type Controller struct {
	// KubeCli is a standard kubernetes clientset
	KubeCli kubernetes.Interface
	// ApiCli is the extension kubernetes clientset
	ApiCli apiextensionsclient.Interface
	// PaddleCli is a clientset for our own API group
	PaddleCli paddleclientset.Interface

	trainingjobLister paddlelisters.TrainingJobLister
	trainingjobSynced cache.InformerSynced

	// TODO jobtracker keep track of every training job
	jobtracker map[string]interface{}

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func New(
	kubeCli kubernetes.Interface,
	apiCli apiextensionsclient.Interface,
	paddleCli paddleclientset.Interface,
	tjInformer paddleinformers.SharedInformerFactory) *Controller {

	traingingjobInformer := tjInformer.Paddlepaddle().V1alpha1().TrainingJobs()

	paddlescheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating trainingjob event broadcaster")
	eventtBroadcaster := record.NewBroadcaster()
	eventtBroadcaster.StartLogging(glog.Infof)
	eventtBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeCli.CoreV1().Events("")})
	recorder := eventtBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "TrainingJobController"})

	controller := &Controller{
		KubeCli:           kubeCli,
		ApiCli:            apiCli,
		PaddleCli:         paddleCli,
		trainingjobLister: traingingjobInformer.Lister(),
		trainingjobSynced: traingingjobInformer.Informer().HasSynced,
		jobtracker:        make(map[string]interface{}, 0),
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TrainingJob"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers")
	traingingjobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueue,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldTj := oldObj.(*paddlev1alpha1.TrainingJob)
			newTj := newObj.(*paddlev1alpha1.TrainingJob)
			if oldTj.ResourceVersion == newTj.ResourceVersion {
				glog.V(4).Infof("same resourceversion for training job %s/%s, skipped", oldTj.Namespace, oldTj.Name)
				return
			}
			glog.V(4).Infof("resourceversion for training job %s/%s updated", oldTj.Namespace, oldTj.Name)
			controller.enqueue(newObj)
		},
		DeleteFunc: controller.dequeue,
	})

	return controller
}

// Run will set up the event handlers for trainingjob, as well as syncing
// informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	// TODO add a lock to ensure there is only one controller in the cluster
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	glog.Info("Starting trainingjob controller")
	glog.Info("Starting to create custom resource definition")

	if err := c.createCRD(); err != nil {
		return fmt.Errorf("failed to create kind TrainingJob: %v", err)
	}

	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.trainingjobSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

func (c *Controller) createCRD() error {
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: paddlev1alpha1.CRDName(),
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   paddlev1alpha1.CRDGroup,
			Version: paddlev1alpha1.CRDVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:       paddlev1alpha1.CRDKind,
				Plural:     paddlev1alpha1.CRDKindPlural,
				ShortNames: []string{paddlev1alpha1.CRDShortName},
			},
		},
	}

	_, err := c.ApiCli.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	glog.Infof("CRD %s created", paddlev1alpha1.CRDName())

	return nil
}

// enqueue takes a TrainingJob resource and converts it into a namespace/name
// string which is then put onto the work queue.
func (c *Controller) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	glog.Infof("enqueue key: %v", key)
	c.workqueue.AddRateLimited(key)
}

func (c *Controller) dequeue(obj interface{}) {
	job := obj.(*paddlev1alpha1.TrainingJob)
	key := job.Namespace + "/" + job.Name
	glog.Infof("dequeue key: %v", key)
	if _, ok := c.jobtracker[key]; !ok {
		glog.Warningf("unsafe state. %s was never created but we received delete event", key)
	}
	//c.jobtracker[key].Delete()
	delete(c.jobtracker, key)
}

func (c *Controller) runWorker() {
	for c.processNestWorkItem() {
	}
}

func (c *Controller) processNestWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}

		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	glog.Infof("syncHandler, key: %s", key)
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	job, createErr := c.trainingjobLister.TrainingJobs(ns).Get(name)
	if createErr != nil {
		glog.Errorf("get trainingjob error: %v", err)
		if apierrors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("trainingjob '%s' in the work queue no longer exists", key))
			return nil
		}

		return err
	}

	_, ok := c.jobtracker[key]
	if !ok {
		glog.Infof("create a new job tracker, key: '%s'", key)
		glog.Infof("received job: %+v", job)
		// TODO create a tracker for the job, just record its name here
		c.jobtracker[key] = job.Name
	}

	return nil
}
