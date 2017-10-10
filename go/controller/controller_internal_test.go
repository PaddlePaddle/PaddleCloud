package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynamicScaling(t *testing.T) {

}

func TestFulfillment(t *testing.T) {
	j := job{}

	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 2
	j.CurInstance = 2
	assert.Equal(t, float64(1), j.Fulfullment())

	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 2
	j.CurInstance = 1
	assert.Equal(t, float64(0), j.Fulfullment())

	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.CurInstance = 2
	assert.Equal(t, float64(0.5), j.Fulfullment())
}

func TestSortedElasticJobs(t *testing.T) {
	jobs := make([]job, 4)
	jobs[0].CurInstance = 2
	jobs[0].Config.MetaData.Name = "a"
	jobs[0].Config.Spec.Trainer.MinInstance = 1
	jobs[0].Config.Spec.Trainer.MaxInstance = 2

	jobs[1].CurInstance = 2
	jobs[1].Config.MetaData.Name = "b"
	jobs[1].Config.Spec.Trainer.MinInstance = 1
	jobs[1].Config.Spec.Trainer.MaxInstance = 20

	jobs[2].CurInstance = 2
	jobs[2].Config.MetaData.Name = "c"
	jobs[2].Config.Spec.Trainer.MinInstance = 1
	jobs[2].Config.Spec.Trainer.MaxInstance = 10

	jobs[3].CurInstance = 1
	jobs[3].Config.MetaData.Name = "d"
	jobs[3].Config.Spec.Trainer.MinInstance = 1
	jobs[3].Config.Spec.Trainer.MaxInstance = 1

	expected := []string{"b", "c", "a"}

	c := New(nil)
	for _, j := range jobs {
		c.jobs[j.Config.MetaData.Name] = j
	}

	assert.Equal(t, expected, c.sortedElasticJobs())
}

func TestSortedElasticJobsWithTie(t *testing.T) {
	jobs := make([]job, 4)
	jobs[0].CurInstance = 1
	jobs[0].Config.MetaData.Name = "a"
	jobs[0].Config.Spec.Trainer.MinInstance = 1
	jobs[0].Config.Spec.Trainer.MaxInstance = 2
	jobs[0].Config.Spec.Trainer.Resources.Limits.GPU = 1

	jobs[1].CurInstance = 1
	jobs[1].Config.MetaData.Name = "b"
	jobs[1].Config.Spec.Trainer.MinInstance = 1
	jobs[1].Config.Spec.Trainer.MaxInstance = 2
	jobs[1].Config.Spec.Trainer.Resources.Requests.CPU = 1
	jobs[1].Config.Spec.Trainer.Resources.Requests.Mem = 1

	jobs[2].CurInstance = 1
	jobs[2].Config.MetaData.Name = "c"
	jobs[2].Config.Spec.Trainer.MinInstance = 1
	jobs[2].Config.Spec.Trainer.MaxInstance = 2
	jobs[2].Config.Spec.Trainer.Resources.Requests.CPU = 10

	jobs[3].CurInstance = 1
	jobs[3].Config.MetaData.Name = "d"
	jobs[3].Config.Spec.Trainer.MinInstance = 1
	jobs[3].Config.Spec.Trainer.MaxInstance = 2
	jobs[3].Config.Spec.Trainer.Resources.Requests.CPU = 1
	jobs[3].Config.Spec.Trainer.Resources.Requests.Mem = 2

	expected := []string{"b", "d", "c", "a"}

	c := New(nil)
	for _, j := range jobs {
		c.jobs[j.Config.MetaData.Name] = j
	}

	assert.Equal(t, expected, c.sortedElasticJobs())
}
