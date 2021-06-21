/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package core

import (
	"fmt"
	"time"

	kubeutil "hostport-manager/pkg/utils/kubernetes"
	"hostport-manager/pkg/utils/portparse"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/registry/core/service/portallocator"
)

const controllerAgentName = "hostport-manager"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
)

// HostPortManagerOptions contain various options to customize how HostPortManager works
type HostPortManagerOptions struct {
	HostNodePortRange  utilnet.PortRange
	UseServiceNodePort bool
}

// HostPortManager is the controller implementation for hostport
type HostPortManager struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset               kubernetes.Interface
	restClient                  rest.Interface
	paddleSynced                cache.InformerSynced
	replicationControllerSynced cache.InformerSynced
	hostPortAllocator           portallocator.Interface
	// workqueue is a rate limited work queue. This is used to queue work to be processed instead of performing it as soon as a change happens.
	// This means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the Kubernetes API.
	recorder record.EventRecorder
}

// NewHostPortManager returns a new HostPortManager
func NewHostPortManager(opts HostPortManagerOptions,
	kubeclientset kubernetes.Interface, restClient *rest.RESTClient,
	kubeInformerFactory kubeinformers.SharedInformerFactory, stopCh <-chan struct{}) *HostPortManager {
	_, err := kubeclientset.Core().Nodes().List(metav1.ListOptions{})
	if err != nil {
		glog.Fatalf("Failed to get nodes from apiserver: %v", err)
	}
	glog.V(4).Info("Creating event broadcaster")
	recorder := kubeutil.CreateEventRecorder(kubeclientset)

	controller := &HostPortManager{
		kubeclientset:     kubeclientset,
		restClient:        restClient,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TrainingJob"),
		recorder:          recorder,
		hostPortAllocator: portallocator.NewPortAllocator(opts.HostNodePortRange),
	}

	glog.Info("Setting up event handlers")
	controller.addPaddleInformer(stopCh)
	// controller.addReplicationControllerInformer(kubeInformerFactory)
	return controller
}

func (c *HostPortManager) addPaddleInformer(stopCh <-chan struct{}) {
	source := cache.NewListWatchFromClient(c.restClient, paddlejob.TrainingJobs, "", fields.Everything())

	_, paddleInformer := cache.NewInformer(
		source,
		&paddlejob.TrainingJob{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.handleObject,
			UpdateFunc: func(old, new interface{}) {
				newPaddleTrain := new.(*paddlejob.TrainingJob)
				oldPaddleTrain := old.(*paddlejob.TrainingJob)
				if newPaddleTrain.ResourceVersion == oldPaddleTrain.ResourceVersion {
					return
				}
				// c.handleObject(new)
			},
			DeleteFunc: c.deleteObject,
		})
	go paddleInformer.Run(stopCh)
	glog.Info("PaddleInformer run")
	c.paddleSynced = paddleInformer.HasSynced

}

/*
func (c *HostPortManager) addReplicationControllerInformer(kubeInformerFactory kubeinformers.SharedInformerFactory) {
	rcInformer := kubeInformerFactory.Core().V1().ReplicationControllers()
	rcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newReplicationController := new.(*clientv1.ReplicationController)
			oldReplicationController := old.(*clientv1.ReplicationController)
			if newReplicationController.ResourceVersion == oldReplicationController.ResourceVersion {
				return
			}
			anno := newReplicationController.GetAnnotations()
			var exist bool
			var port string
			if port, exist = anno["hostport-manager/hostport"]; !exist {
				return
			}
			_, err := strconv.Atoi(port)
			if err != nil {
				c.handleObject(new)
			}
		},
		DeleteFunc: c.deleteObject,
	})

	c.replicationControllerSynced = rcInformer.Informer().HasSynced
}
*/

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *HostPortManager) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting HostPort controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.paddleSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *HostPortManager) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *HostPortManager) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
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

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *HostPortManager) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	glog.Infof("namespace:%s   name:%s", namespace, name)
	result := &paddlejob.TrainingJob{}
	err = c.restClient.Get().
		Namespace(namespace).
		Resource("trainingjobs").
		Name(name).
		Do().
		Into(result)

	if err != nil {
		// The paddlejob resource may no longer exist, in which case we stop
		// processing.
		glog.Errorf("Get trainingjobs error:%v", err)
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("paddlejob '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}
	anno := result.GetAnnotations()
	num, err := portparse.ParsePortNumAnno(anno)
	if err != nil {
		return err
	}
	ports, err := portparse.ParsePortsAllocAnno(anno)
	if err != nil {
		return err
	}
	if ports == nil {
		portsAnno := ""
		for index := 0; index < num; index++ {
			hostport, err := c.hostPortAllocator.AllocateNext()
			if err != nil {
				return err
			}
			if portsAnno == "" {
				portsAnno = fmt.Sprintf("%d", hostport)
			} else {
				portsAnno = fmt.Sprintf("%s,%d", portsAnno, hostport)
			}
		}
		if anno == nil {
			anno = make(map[string]string)
		}
		anno["hostport-manager/hostport"] = portsAnno
		result.SetAnnotations(anno)
		resultJob := &paddlejob.TrainingJob{}
		err = c.restClient.Put().
			Namespace(namespace).
			Resource("trainingjobs").
			Name(name).Body(result).
			Do().Into(resultJob)
		if err != nil {
			glog.Errorf("Put trainingjobs error:%v", err)
			return err
		}
		glog.Infof("paddlejob:%s, allocated hostport:%s", resultJob.Name, portsAnno)
		// c.recorder.Event(rc, clientv1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	}

	return nil
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *HostPortManager) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Foo resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Foo resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *HostPortManager) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	needNum, err := portparse.ParsePortNumAnno(object.GetAnnotations())
	if err != nil || needNum == 0 {
		return
	}

	glog.V(4).Infof("Processing object: %s", object.GetName())
	// need to allocate hostport
	portList, err := portparse.ParsePortsAllocAnno(object.GetAnnotations())
	if err != nil {
		glog.Error("%s", err.Error())
		// TODO: need to delete the invalid annotation
		return
	}
	if portList == nil {
		c.enqueue(object)
		return
	}
	if len(portList) != needNum {
		// TODO: need to delete the invalid annotation
	}

	for _, p := range portList {
		c.hostPortAllocator.Allocate(p)
	}
	return
}

// deleteObject will free allocated ports
func (c *HostPortManager) deleteObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Deleting object: %s", object.GetName())
	glog.V(4).Infof("Deleting object: %s", object.GetAnnotations())
	ports, err := portparse.ParsePortsAllocAnno(object.GetAnnotations())
	if err != nil {
		glog.Errorf("Deleting object  %s ParsePortsAllocAnno: %v", object.GetName(), err)
		return
	}
	for _, port := range ports {
		c.hostPortAllocator.Release(port)
		glog.V(4).Infof("Release port:%d", port)
	}

	return

}
