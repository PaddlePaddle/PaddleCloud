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

package k8s

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// Controller for dispatching TrainingJob resource.
type Controller struct {
	client *rest.RESTClient
	scheme *runtime.Scheme
}

// Run start to watch kubernetes events and do handlers.
func (c *Controller) Run(ctx context.Context) error {
	err := c.startWatch(ctx)
	if err != nil {
		return err
	}
	<-ctx.Done()
	return ctx.Err()
}

func (c *Controller) startWatch(ctx context.Context) error {
	source := cache.NewListWatchFromClient(
		c.client,
		TrainingJobs,
		apiv1.NamespaceAll,
		fields.Everything())

	_, informer := cache.NewInformer(
		source,
		&TrainingJob{},

		// resyncPeriod
		// Every resyncPeriod, all resources in the cache will retrigger events.
		// Set to 0 to disable the resync.
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
	job := obj.(*TrainingJob)
	// call c.client.Put() to send REST call to api-server
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	oldjob := oldObj.(*TrainingJob)
	newjob := newObj.(*TrainingJob)
	// call c.client.Put() to update resource
}

func (c *Controller) onDelete(obj interface{}) {
	job := obj.(*TrainingJob)
	// call c.client.Delete()

}
