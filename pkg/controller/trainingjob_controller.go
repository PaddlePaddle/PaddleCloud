package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/inconshreveable/log15"
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

	paddlev1 "github.com/paddleflow/paddle-operator/pkg/apis/paddlepaddle/v1alpha1"
	"github.com/paddleflow/paddle-operator/pkg/autoscaler"
	paddleclientset "github.com/paddleflow/paddle-operator/pkg/client/clientset/versioned"
	paddlescheme "github.com/paddleflow/paddle-operator/pkg/client/clientset/versioned/scheme"
	paddleinformers "github.com/paddleflow/paddle-operator/pkg/client/informers/externalversions"
	paddlelisters "github.com/paddleflow/paddle-operator/pkg/client/listers/paddlepaddle/v1alpha1"
	"github.com/paddleflow/paddle-operator/pkg/updater"
)

// TrainingJobController defines the structure to manage TrainingJob resource
type TrainingJobController struct {
	// kubeCli is a standard kubernetes clientset
	kubeCli kubernetes.Interface
	// apiCli is the extension kubernetes clientset
	apiCli apiextensionsclient.Interface
	// paddleCli is a clientset for our own API group
	paddleCli paddleclientset.Interface

	trainingjobLister paddlelisters.TrainingJobLister
	trainingjobSynced cache.InformerSynced

	// jobtracker keeps a map from job full name to its updater
	jobtracker *sync.Map

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// autoclean means whether or not cleaning pods after termination.
	autoclean bool

	//restartLimit means pserver restart count reach to limit
	restartLimit int

	//outter meas if operator runs out of baidu
	outter bool
}

// New returns a TrainingJobController object
func New(
	kubeCli kubernetes.Interface,
	apiCli apiextensionsclient.Interface,
	paddleCli paddleclientset.Interface,
	tjInformer paddleinformers.SharedInformerFactory,
	auto bool, restartLimit int, outter bool) *TrainingJobController {

	traingingjobInformer := tjInformer.Paddlepaddle().V1alpha1().TrainingJobs()

	paddlescheme.AddToScheme(scheme.Scheme)
	log.Debug("Creating trainingjob event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(log.Info)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeCli.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "TrainingJobController"})

	controller := &TrainingJobController{
		kubeCli:           kubeCli,
		apiCli:            apiCli,
		paddleCli:         paddleCli,
		trainingjobLister: traingingjobInformer.Lister(),
		trainingjobSynced: traingingjobInformer.Informer().HasSynced,
		jobtracker:        new(sync.Map),
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TrainingJob"),
		recorder:          recorder,
		autoclean:         auto,
		restartLimit:      restartLimit,
		outter:            outter,
	}

	log.Info("Setting up event handlers")
	traingingjobInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *paddlev1.TrainingJob:
					log.Debug("filter trainingjob", "namespace", t.Namespace, "name", t.Name)
					return true
				default:
					return false
				}
			},

			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					log.Debug("AddFunc called")
					controller.enqueue(obj)
				},
				UpdateFunc: func(oldObj, newObj interface{}) {
					oldTj := oldObj.(*paddlev1.TrainingJob)
					newTj := newObj.(*paddlev1.TrainingJob)
					if oldTj.ResourceVersion == newTj.ResourceVersion {
						log.Debug("same resourceversion skipped", "namespace", oldTj.Namespace, "name", oldTj.Name)
						return
					}
					log.Debug("resourceversion updated", "namespace", oldTj.Namespace, "name", oldTj.Name)
					controller.enqueue(newObj)
				},
				DeleteFunc: func(obj interface{}) {
					log.Debug("DeleteFunc called")
					controller.enqueue(obj)
				},
			},
		})

	return controller
}

// Run will set up the event handlers for trainingjob, as well as syncing
// informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *TrainingJobController) Run(threadiness int, maxLoadDesired float64, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	log.Info("Starting trainingjob controller")
	log.Info("Starting to create custom resource definition")

	if err := c.createCRD(); err != nil {
		return fmt.Errorf("failed to create kind TrainingJob: %v", err)
	}

	log.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.trainingjobSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	gc := NewGarbageCollector(c.kubeCli, c.trainingjobLister)
	go gc.CleanOrphans(10 * time.Minute)

	log.Info("Started workers")

	as := autoscaler.NewAutoscaler(c.kubeCli, c.jobtracker, autoscaler.WithMaxLoadDesired(maxLoadDesired))
	as.Run()

	// FIXME: as.Run() will unconditionally loop forever. The following lines will not be reached.
	<-stopCh
	log.Info("Shutting down workers")

	return nil
}

func (c *TrainingJobController) createCRD() error {
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: paddlev1.CRDName(),
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   paddlev1.CRDGroup,
			Version: paddlev1.CRDVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:       paddlev1.CRDKind,
				Plural:     paddlev1.CRDKindPlural,
				ShortNames: []string{paddlev1.CRDShortName},
			},
		},
	}

	_, err := c.apiCli.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		log.Error("Failed to create crd", "err", err.Error())
		return err
	}

	return nil
}

// enqueue takes a TrainingJob resource and converts it into a namespace/name
// string which is then put onto the work queue.
func (c *TrainingJobController) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	log.Info("enqueue", "key", key)
	c.workqueue.AddRateLimited(key)
}

func (c *TrainingJobController) runWorker() {
	log.Debug("Run worker again")
	for c.processNextWorkItem() {
	}
}

func (c *TrainingJobController) processNextWorkItem() bool {
	key, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	defer c.workqueue.Done(key)

	forget, err := c.syncHandler(key.(string))
	if err == nil {
		if forget {
			c.workqueue.Forget(key)
			log.Info("Successfully synced", "key", key.(string))
		}
		return true
	}

	runtime.HandleError(fmt.Errorf("Error syncing job: %v", err))
	c.workqueue.AddRateLimited(key)

	return true
}

func (c *TrainingJobController) syncHandler(key string) (bool, error) {
	log.Info("syncHandler", "key", key)
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return false, nil
	}

	jobIsDeleted := false
	job, getErr := c.trainingjobLister.TrainingJobs(ns).Get(name)
	if getErr != nil {
		log.Debug("Error fetching TrainingJob", "key", key, "err", getErr.Error())
		if apierrors.IsNotFound(getErr) {
			jobIsDeleted = true
		} else {
			return false, nil
		}
	} else {
		log.Debug("TrainingJob fetching status", "namespace", job.Namespace, "name", job.Name, "status", job.Status)
	}

	var jobUpdater *updater.JobUpdater
	jobUpdaterObj, exists := c.jobtracker.Load(key)

	if !exists {
		if jobIsDeleted {
			log.Debug("key not exist", "key", key)
			return true, fmt.Errorf("JobNotExists")
		}
		log.Debug("TrainingJob new", "namespace", job.Namespace, "name", job.Name)
		nj := updater.NewJobUpdater(job, c.kubeCli, c.paddleCli, c.autoclean, c.restartLimit, c.outter)
		c.jobtracker.Store(key, nj)
		jobUpdater = nj
	} else {
		var ok bool
		jobUpdater, ok = jobUpdaterObj.(*updater.JobUpdater)
		if !ok {
			log.Debug("Conversion object error", "object", jobUpdaterObj)
			return true, fmt.Errorf("ConversionError")
		}

		if jobIsDeleted {
			// clean job
			log.Info("Deleting TrainingJob", "name", jobUpdater.FullName())
			if err := jobUpdater.Delete(); err != nil {
				log.Error("Error deleting TrainingJob", "name", jobUpdater.FullName(), "err", err.Error())
				return false, nil
			}
			log.Info("Finishing deleting TrainingJob", "name", jobUpdater.FullName())
			c.jobtracker.Delete(key)
			return true, nil
		}

		if jobUpdater.UID() != job.ObjectMeta.UID {
			// update job
			log.Debug("TrainingJob UID changed", "namespace", job.Namespace, "name", job.Name)
			jobUpdater.Update(job)
		}
	}

	if jobUpdater.IsReleased() {
		log.Info("Ignore reconciling", "namespace", job.Namespace, "job", job.Name,
			"since whose resource has been released")
		return false, nil
	}

	if err := jobUpdater.Reconcile(); err != nil {
		log.Error("Error reconciling", "namespace", job.Namespace, "name", job.Name, "err", err.Error())
		return false, err
	}

	currentPhase := jobUpdater.GetJob().Status.Phase

	if currentPhase == paddlev1.TrainingJobPhaseCreating ||
		currentPhase == paddlev1.TrainingJobPhaseRunning ||
		currentPhase == paddlev1.TrainingJobPhaseScaling {
		c.workqueue.AddAfter(key, 3*time.Second)
		log.Debug("TrainingJob put into workqueue again", "key", key, "current statue phase", currentPhase)
	}

	return false, nil
}
