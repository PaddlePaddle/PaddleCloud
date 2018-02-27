/*
Copyright (c) 2016 PaddlePaddle Authors All Rights Reserve.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/package v1

import (
	v1 "github.com/PaddlePaddle/cloud/go/apis/paddlepaddle/v1"
	scheme "github.com/PaddlePaddle/cloud/go/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Create(*v1.TrainingJob) (*v1.TrainingJob, error)
	Update(*v1.TrainingJob) (*v1.TrainingJob, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.TrainingJob, error)
	List(opts meta_v1.ListOptions) (*v1.TrainingJobList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.TrainingJob, err error)
	TrainingJobExpansion
}

// trainingJobs implements TrainingJobInterface
type trainingJobs struct {
	client rest.Interface
	ns     string
}

// newTrainingJobs returns a TrainingJobs
func newTrainingJobs(c *PaddlepaddleV1Client, namespace string) *trainingJobs {
	return &trainingJobs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the trainingJob, and returns the corresponding trainingJob object, and an error if there is any.
func (c *trainingJobs) Get(name string, options meta_v1.GetOptions) (result *v1.TrainingJob, err error) {
	result = &v1.TrainingJob{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of TrainingJobs that match those selectors.
func (c *trainingJobs) List(opts meta_v1.ListOptions) (result *v1.TrainingJobList, err error) {
	result = &v1.TrainingJobList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested trainingJobs.
func (c *trainingJobs) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a trainingJob and creates it.  Returns the server's representation of the trainingJob, and an error, if there is any.
func (c *trainingJobs) Create(trainingJob *v1.TrainingJob) (result *v1.TrainingJob, err error) {
	result = &v1.TrainingJob{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("trainingjobs").
		Body(trainingJob).
		Do().
		Into(result)
	return
}

// Update takes the representation of a trainingJob and updates it. Returns the server's representation of the trainingJob, and an error, if there is any.
func (c *trainingJobs) Update(trainingJob *v1.TrainingJob) (result *v1.TrainingJob, err error) {
	result = &v1.TrainingJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(trainingJob.Name).
		Body(trainingJob).
		Do().
		Into(result)
	return
}

// Delete takes name of the trainingJob and deletes it. Returns an error if one occurs.
func (c *trainingJobs) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *trainingJobs) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched trainingJob.
func (c *trainingJobs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.TrainingJob, err error) {
	result = &v1.TrainingJob{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("trainingjobs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
