package autoscaler

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/PaddlePaddle/cloud/go/api"
	"github.com/stretchr/testify/assert"
)

func TestFulfillment(t *testing.T) {
	j := job{}

	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 2
	j.CurInstance = 2
	assert.Equal(t, float64(1), j.Fulfillment())

	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 2
	j.CurInstance = 1
	assert.Equal(t, float64(0), j.Fulfillment())

	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.CurInstance = 2
	assert.Equal(t, float64(0.5), j.Fulfillment())
}

func TestSortedJobs(t *testing.T) {
	jobs := make([]job, 4)
	jobs[0].CurInstance = 2
	jobs[0].Config.Name = "a"
	jobs[0].Config.Spec.Trainer.MinInstance = 1
	jobs[0].Config.Spec.Trainer.MaxInstance = 2

	jobs[1].CurInstance = 2
	jobs[1].Config.Name = "b"
	jobs[1].Config.Spec.Trainer.MinInstance = 1
	jobs[1].Config.Spec.Trainer.MaxInstance = 20

	jobs[2].CurInstance = 2
	jobs[2].Config.Name = "c"
	jobs[2].Config.Spec.Trainer.MinInstance = 1
	jobs[2].Config.Spec.Trainer.MaxInstance = 10

	jobs[3].CurInstance = 1
	jobs[3].Config.Name = "d"
	jobs[3].Config.Spec.Trainer.MinInstance = 1
	jobs[3].Config.Spec.Trainer.MaxInstance = 1

	expected := []string{"b", "c", "a"}

	c := NewAutoscaler(nil)
	for _, j := range jobs {
		c.jobs[j.Config.Name] = j
	}

	assert.Equal(t, expected, c.sortedJobs(elastic))
}

func TestSortedJobsGPUOnly(t *testing.T) {
	jobs := make([]job, 4)
	jobs[0].CurInstance = 2
	jobs[0].Config.Name = "a"
	jobs[0].Config.Spec.Trainer.MinInstance = 1
	jobs[0].Config.Spec.Trainer.MaxInstance = 2

	q, err := resource.ParseQuantity("1")
	assert.Nil(t, err)
	jobs[0].Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	jobs[0].Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q

	jobs[1].CurInstance = 2
	jobs[1].Config.Name = "b"
	jobs[1].Config.Spec.Trainer.MinInstance = 1
	jobs[1].Config.Spec.Trainer.MaxInstance = 20

	jobs[2].CurInstance = 2
	jobs[2].Config.Name = "c"
	jobs[2].Config.Spec.Trainer.MinInstance = 1
	jobs[2].Config.Spec.Trainer.MaxInstance = 10

	jobs[3].CurInstance = 1
	jobs[3].Config.Name = "d"
	jobs[3].Config.Spec.Trainer.MinInstance = 1
	jobs[3].Config.Spec.Trainer.MaxInstance = 1

	expected := []string{"a"}

	c := NewAutoscaler(nil)
	for _, j := range jobs {
		c.jobs[j.Config.Name] = j
	}

	assert.Equal(t, expected, c.sortedJobs(elastic, gpu))
}

func TestSortedJobsWithTie(t *testing.T) {
	jobs := make([]job, 4)
	jobs[0].CurInstance = 1
	jobs[0].Config.Name = "a"
	jobs[0].Config.Spec.Trainer.MinInstance = 1
	jobs[0].Config.Spec.Trainer.MaxInstance = 2
	q, err := resource.ParseQuantity("1")
	assert.Nil(t, err)
	jobs[0].Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	jobs[0].Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q

	jobs[1].CurInstance = 1
	jobs[1].Config.Name = "b"
	jobs[1].Config.Spec.Trainer.MinInstance = 1
	jobs[1].Config.Spec.Trainer.MaxInstance = 2
	q, err = resource.ParseQuantity("1")
	assert.Nil(t, err)
	jobs[1].Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	jobs[1].Config.Spec.Trainer.Resources.Requests["cpu"] = q
	jobs[1].Config.Spec.Trainer.Resources.Requests["memory"] = q

	jobs[2].CurInstance = 1
	jobs[2].Config.Name = "c"
	jobs[2].Config.Spec.Trainer.MinInstance = 1
	jobs[2].Config.Spec.Trainer.MaxInstance = 2
	q, err = resource.ParseQuantity("10")
	assert.Nil(t, err)
	jobs[2].Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	jobs[2].Config.Spec.Trainer.Resources.Requests["cpu"] = q

	jobs[3].CurInstance = 1
	jobs[3].Config.Name = "d"
	jobs[3].Config.Spec.Trainer.MinInstance = 1
	jobs[3].Config.Spec.Trainer.MaxInstance = 2
	q, err = resource.ParseQuantity("1")
	assert.Nil(t, err)
	jobs[3].Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	jobs[3].Config.Spec.Trainer.Resources.Requests["cpu"] = q
	q, err = resource.ParseQuantity("2")
	assert.Nil(t, err)
	jobs[3].Config.Spec.Trainer.Resources.Requests["memory"] = q

	expected := []string{"b", "d", "c", "a"}

	c := NewAutoscaler(nil)
	for _, j := range jobs {
		c.jobs[j.Config.Name] = j
	}

	assert.Equal(t, expected, c.sortedJobs(elastic))
}
