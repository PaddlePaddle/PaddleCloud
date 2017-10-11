/* Copyright (c) 2016 PaddlePaddle Authors All Rights Reserve.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
	 limitations under the License. */

// Controller is responsible to watch resource type "TrainingJob"
// event and parse "TrainingJob" into several other resources like
// "Job" and "ReplicaSet".

// Controller will manage "TrainingJob" creation and destruction while
// AutoScaler will scale the job to maximize the cluster resource usage.

// When controller starts, both event watching routine and resource
// monitoring and scaling routine should be started.

package k8s

import (
	"context"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	v1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/kubernetes/pkg/api"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	paddlejob "github.com/PaddlePaddle/cloud/go/controller"
)

// Controller for dispatching TrainingJob resource.
type Controller struct {
	client     *rest.RESTClient
	clientset  *kubernetes.Clientset
	autoscaler *paddlejob.Autoscaler
}

// NewController construct a new Controller struct
func NewController(config *rest.Config) (*Controller, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}
	as := paddlejob.NewAutoscaler(nil)
	return &Controller{
		client:     client,
		clientset:  clientset,
		autoscaler: as,
	}, nil
}

// Run start to watch kubernetes events and do handlers.
func (c *Controller) Run(ctx context.Context) error {
	err := c.startWatch(ctx)
	if err != nil {
		return err
	}
	// TODO(helin): start autoscaler

	<-ctx.Done()
	return ctx.Err()
}

func (c *Controller) startWatch(ctx context.Context) error {
	source := cache.NewListWatchFromClient(
		c.client,
		paddlejob.TrainingJobs,
		api.NamespaceAll,
		fields.Everything())

	_, informer := cache.NewInformer(
		source,
		&paddlejob.TrainingJob{},

		// resyncPeriod: Every resyncPeriod, all resources in
		// the cache will retrigger events. Set to 0 to
		// disable the resync.
		0,

		// TrainingJob custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		})

	go informer.Run(ctx.Done())
	return nil
}

func (c *Controller) onAdd(obj interface{}) {
	job := obj.(*paddlejob.TrainingJob)
	log.Debugln("onAdd: %v", *job)
	// TODO: call c.client.Put() to send REST call to api-server
	namespace := job.ObjectMeta.Namespace
	jobname := job.ObjectMeta.Name
	rslist, err := c.clientset.ExtensionsV1beta1().ReplicaSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorln(err)
	}
	log.Debugln("Existing replicasets: %v", *rslist)
	exist := false
	for _, item := range rslist.Items {
		if item.Name == jobname {
			exist = true
			break
		}
	}

	if exist {
		log.Errorln("Job name already exists:", jobname)
		return
	}

	// generate a pserver replicaset resource according to "TrainingJob" resource specs.
	pserverRS := v1beta1.ReplicaSet{}
	c.clientset.ExtensionsV1beta1().ReplicaSets(namespace).Create(&pserverRS)
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	oldjob := oldObj.(*paddlejob.TrainingJob)
	newjob := newObj.(*paddlejob.TrainingJob)
	log.Debugln(oldjob)
	log.Debugln(newjob)
	log.Debugln("onUpdate.")
	// TODO: call c.client.Put() to update resource
}

func (c *Controller) onDelete(obj interface{}) {
	job := obj.(*paddlejob.TrainingJob)
	log.Debugln("onDelete.")
	log.Debugln(job)
	// TODO: call c.client.Delete()
}
