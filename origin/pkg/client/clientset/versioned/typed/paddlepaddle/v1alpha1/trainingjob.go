/*
Copyright (c) 2016 PaddlePaddle Authors All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/package v1alpha1

import (
	"context"

	v1alpha1 "github.com/paddleflow/paddle-operator/pkg/apis/paddlepaddle/v1alpha1"
	scheme "github.com/paddleflow/paddle-operator/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// TrainingJobsGetter has a method to return a TrainingJobInterface.
// A group's client should implement this interface.
type TrainingJobsGetter interface {
	TrainingJobs(namespace string) TrainingJobInterface
}

// TrainingJobInterface has methods to work with TrainingJob resources.
type TrainingJobInterface interface {
	Create(context.Context, *v1alpha1.TrainingJob) (*v1alpha1.TrainingJob, error)
	Update(context.Context, *v1alpha1.TrainingJob) (*v1alpha1.TrainingJob, error)
	Delete(ctx context.Context, name string, options *v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(ctx context.Context, name string, options v1.GetOptions) (*v1alpha1.TrainingJob, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.TrainingJobList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.TrainingJob, err error)
	TrainingJobExpansion
}

// trainingJobs implements TrainingJobInterface
type trainingJobs struct {
	client rest.Interface
	ns     string
}

// newTrainingJobs returns a TrainingJobs
func newTrainingJobs(c *PaddlepaddleV1alpha1Client, namespace string) *trainingJobs {
	return &trainingJobs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the trainingJob, and returns the corresponding trainingJob object, and an error if there is any.
func (c *trainingJobs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.TrainingJob, err error) {
	result = &v1alpha1.TrainingJob{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of TrainingJobs that match those selectors.
func (c *trainingJobs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.TrainingJobList, err error) {
	result = &v1alpha1.TrainingJobList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested trainingJobs.
func (c *trainingJobs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}

// Create takes the representation of a trainingJob and creates it.  Returns the server's representation of the trainingJob, and an error, if there is any.
func (c *trainingJobs) Create(ctx context.Context, trainingJob *v1alpha1.TrainingJob) (result *v1alpha1.TrainingJob, err error) {
	result = &v1alpha1.TrainingJob{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("trainingjobs").
		Body(trainingJob).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a trainingJob and updates it. Returns the server's representation of the trainingJob, and an error, if there is any.
func (c *trainingJobs) Update(ctx context.Context, trainingJob *v1alpha1.TrainingJob) (result *v1alpha1.TrainingJob, err error) {
	result = &v1alpha1.TrainingJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(trainingJob.Name).
		Body(trainingJob).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the trainingJob and deletes it. Returns an error if one occurs.
func (c *trainingJobs) Delete(ctx context.Context, name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(name).
		Body(options).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *trainingJobs) DeleteCollection(ctx context.Context, options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched trainingJob.
func (c *trainingJobs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.TrainingJob, err error) {
	result = &v1alpha1.TrainingJob{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("trainingjobs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
