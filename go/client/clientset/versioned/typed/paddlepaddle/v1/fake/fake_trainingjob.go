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
*/package fake

import (
	paddlepaddle_v1 "github.com/PaddlePaddle/cloud/go/apis/paddlepaddle/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTrainingJobs implements TrainingJobInterface
type FakeTrainingJobs struct {
	Fake *FakePaddlepaddleV1
	ns   string
}

var trainingjobsResource = schema.GroupVersionResource{Group: "paddlepaddle.org", Version: "v1", Resource: "trainingjobs"}

var trainingjobsKind = schema.GroupVersionKind{Group: "paddlepaddle.org", Version: "v1", Kind: "TrainingJob"}

// Get takes name of the trainingJob, and returns the corresponding trainingJob object, and an error if there is any.
func (c *FakeTrainingJobs) Get(name string, options v1.GetOptions) (result *paddlepaddle_v1.TrainingJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(trainingjobsResource, c.ns, name), &paddlepaddle_v1.TrainingJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*paddlepaddle_v1.TrainingJob), err
}

// List takes label and field selectors, and returns the list of TrainingJobs that match those selectors.
func (c *FakeTrainingJobs) List(opts v1.ListOptions) (result *paddlepaddle_v1.TrainingJobList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(trainingjobsResource, trainingjobsKind, c.ns, opts), &paddlepaddle_v1.TrainingJobList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &paddlepaddle_v1.TrainingJobList{}
	for _, item := range obj.(*paddlepaddle_v1.TrainingJobList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested trainingJobs.
func (c *FakeTrainingJobs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(trainingjobsResource, c.ns, opts))

}

// Create takes the representation of a trainingJob and creates it.  Returns the server's representation of the trainingJob, and an error, if there is any.
func (c *FakeTrainingJobs) Create(trainingJob *paddlepaddle_v1.TrainingJob) (result *paddlepaddle_v1.TrainingJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(trainingjobsResource, c.ns, trainingJob), &paddlepaddle_v1.TrainingJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*paddlepaddle_v1.TrainingJob), err
}

// Update takes the representation of a trainingJob and updates it. Returns the server's representation of the trainingJob, and an error, if there is any.
func (c *FakeTrainingJobs) Update(trainingJob *paddlepaddle_v1.TrainingJob) (result *paddlepaddle_v1.TrainingJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(trainingjobsResource, c.ns, trainingJob), &paddlepaddle_v1.TrainingJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*paddlepaddle_v1.TrainingJob), err
}

// Delete takes name of the trainingJob and deletes it. Returns an error if one occurs.
func (c *FakeTrainingJobs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(trainingjobsResource, c.ns, name), &paddlepaddle_v1.TrainingJob{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTrainingJobs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(trainingjobsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &paddlepaddle_v1.TrainingJobList{})
	return err
}

// Patch applies the patch and returns the patched trainingJob.
func (c *FakeTrainingJobs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *paddlepaddle_v1.TrainingJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(trainingjobsResource, c.ns, name, data, subresources...), &paddlepaddle_v1.TrainingJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*paddlepaddle_v1.TrainingJob), err
}
