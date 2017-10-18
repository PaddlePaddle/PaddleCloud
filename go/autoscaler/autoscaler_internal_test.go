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
package autoscaler

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"

	"github.com/PaddlePaddle/cloud/go/api"
	"github.com/stretchr/testify/assert"
)

var (
	q1  resource.Quantity
	q10 resource.Quantity
	q0  resource.Quantity
)

func init() {
	var err error
	q1, err = resource.ParseQuantity("1")
	if err != nil {
		panic(err)
	}
	q10, err = resource.ParseQuantity("10")
	if err != nil {
		panic(err)
	}
}

func makePtr(c int) *int32 {
	p := int32(c)
	return &p
}

func TestTrainerRequestLimit(t *testing.T) {
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q10
	j.Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	j.Config.Spec.Trainer.Resources.Requests["memory"] = q1
	assert.Equal(t, float64(1), j.TrainerCPURequest())
	assert.Equal(t, 10, j.TrainerGPULimit())
}

func TestScaleDryRunSatisfied(t *testing.T) {
	r := ClusterResource{CPUFree: 1000, CPUTotal: 1000, MemoryFreeGi: 100, MemoryTotalGi: 1000}
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 2
	j.Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q1
	j.Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	j.Config.Spec.Trainer.Resources.Requests["memory"] = q1
	j.TrainerJob.Spec.Parallelism = makePtr(2)
	assert.Equal(t, 0, scaleDryRun(&r, j, 0))
}

func TestScaleDryRunMoreCPU(t *testing.T) {
	r := ClusterResource{CPUFree: 1000, CPUTotal: 1000, MemoryFreeGi: 100, MemoryTotalGi: 1000}
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q0
	j.Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	j.Config.Spec.Trainer.Resources.Requests["memory"] = q1
	j.TrainerJob.Spec.Parallelism = makePtr(1)
	assert.Equal(t, 1, scaleDryRun(&r, j, 0))
}

func TestScaleDryRunNoMoreCPU(t *testing.T) {
	r := ClusterResource{CPUFree: 0, CPUTotal: 1000, MemoryFreeGi: 100, MemoryTotalGi: 1000}
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q0
	j.Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	j.Config.Spec.Trainer.Resources.Requests["memory"] = q1
	j.TrainerJob.Spec.Parallelism = makePtr(1)
	assert.Equal(t, 0, scaleDryRun(&r, j, 0))
}

func TestScaleDryRunMoreGPU(t *testing.T) {
	r := ClusterResource{CPUFree: 10, CPUTotal: 1000, GPUFree: 1, GPUTotal: 100, MemoryFreeGi: 100, MemoryTotalGi: 1000}
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q1
	j.Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	j.Config.Spec.Trainer.Resources.Requests["memory"] = q1
	j.TrainerJob.Spec.Parallelism = makePtr(1)
	assert.Equal(t, 1, scaleDryRun(&r, j, 0))
}

func TestScaleDryRunNoMoreGPU(t *testing.T) {
	r := ClusterResource{CPUFree: 10, CPUTotal: 1000, GPUFree: 0, GPUTotal: 100, MemoryFreeGi: 100, MemoryTotalGi: 1000}
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q1
	j.Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	j.Config.Spec.Trainer.Resources.Requests["memory"] = q1
	j.TrainerJob.Spec.Parallelism = makePtr(1)
	assert.Equal(t, 0, scaleDryRun(&r, j, 0))
}

func TestScaleDryRunScaleDown(t *testing.T) {
	r := ClusterResource{CPUFree: 1000, CPUTotal: 1000}
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.TrainerJob.Spec.Parallelism = makePtr(6)
	assert.Equal(t, -3, scaleDryRun(&r, j, 0))
}

func TestScaleAllDryRun(t *testing.T) {
	r := ClusterResource{CPUFree: 10, CPUTotal: 1000, GPUFree: 2, GPUTotal: 100, MemoryFreeGi: 100, MemoryTotalGi: 1000}
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q1
	j.Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	j.Config.Spec.Trainer.Resources.Requests["memory"] = q1
	j.TrainerJob.Spec.Parallelism = makePtr(1)
	scale := scaleAllDryRun([]job{j}, r)[""]
	assert.Equal(t, 2, scale)
}

func TestScaleAllDryRunLessCPU(t *testing.T) {
	r := ClusterResource{CPUFree: 1, CPUTotal: 1000, GPUFree: 2, GPUTotal: 100, MemoryFreeGi: 100, MemoryTotalGi: 1000}
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q1
	j.Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	j.Config.Spec.Trainer.Resources.Requests["memory"] = q1
	j.TrainerJob.Spec.Parallelism = makePtr(1)
	scale := scaleAllDryRun([]job{j}, r)[""]
	assert.Equal(t, 1, scale)
}

func TestScaleAllDryRunLessGPU(t *testing.T) {
	r := ClusterResource{CPUFree: 10, CPUTotal: 1000, GPUFree: 1, GPUTotal: 100, MemoryFreeGi: 100, MemoryTotalGi: 1000}
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q1
	j.Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	j.Config.Spec.Trainer.Resources.Requests["memory"] = q1
	j.TrainerJob.Spec.Parallelism = makePtr(1)
	scale := scaleAllDryRun([]job{j}, r)[""]
	assert.Equal(t, 1, scale)
}

func TestFulfillment(t *testing.T) {
	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}

	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 2
	j.TrainerJob.Spec.Parallelism = makePtr(2)
	assert.Equal(t, float64(1), j.Fulfillment())

	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 2
	j.TrainerJob.Spec.Parallelism = makePtr(1)
	assert.Equal(t, float64(0), j.Fulfillment())

	j.Config.Spec.Trainer.MinInstance = 1
	j.Config.Spec.Trainer.MaxInstance = 3
	j.TrainerJob.Spec.Parallelism = makePtr(2)
	assert.Equal(t, float64(0.5), j.Fulfillment())
}

func TestSortedJobs(t *testing.T) {
	jobs := make([]job, 4)

	jobs[0].Config = &api.TrainingJob{}
	jobs[0].TrainerJob = &batchv1.Job{}
	jobs[0].Config.Name = "a"
	jobs[0].Config.Spec.Trainer.MinInstance = 1
	jobs[0].Config.Spec.Trainer.MaxInstance = 2
	jobs[0].TrainerJob.Spec.Parallelism = makePtr(2)

	jobs[1].Config = &api.TrainingJob{}
	jobs[1].TrainerJob = &batchv1.Job{}
	jobs[1].TrainerJob.Spec.Parallelism = makePtr(2)
	jobs[1].Config.Name = "b"
	jobs[1].Config.Spec.Trainer.MinInstance = 1
	jobs[1].Config.Spec.Trainer.MaxInstance = 20

	jobs[2].Config = &api.TrainingJob{}
	jobs[2].TrainerJob = &batchv1.Job{}
	jobs[2].TrainerJob.Spec.Parallelism = makePtr(2)
	jobs[2].Config.Name = "c"
	jobs[2].Config.Spec.Trainer.MinInstance = 1
	jobs[2].Config.Spec.Trainer.MaxInstance = 10

	jobs[3].Config = &api.TrainingJob{}
	jobs[3].TrainerJob = &batchv1.Job{}
	jobs[3].TrainerJob.Spec.Parallelism = makePtr(2)
	jobs[3].Config.Name = "d"
	jobs[3].Config.Spec.Trainer.MinInstance = 1
	jobs[3].Config.Spec.Trainer.MaxInstance = 1

	expected := []string{"b", "c", "a"}

	c := New(nil)
	for _, j := range jobs {
		c.jobs[j.Config.Name] = j
	}

	sorted := sortedJobs(jobs, elastic)
	result := make([]string, len(sorted))
	for i, j := range sorted {
		result[i] = j.Config.Name
	}
	assert.Equal(t, expected, result)
}

func TestSortedJobsGPUOnly(t *testing.T) {
	jobs := make([]job, 4)
	jobs[0].Config = &api.TrainingJob{}
	jobs[0].TrainerJob = &batchv1.Job{}
	jobs[0].TrainerJob.Spec.Parallelism = makePtr(2)
	jobs[0].Config.Name = "a"
	jobs[0].Config.Spec.Trainer.MinInstance = 1
	jobs[0].Config.Spec.Trainer.MaxInstance = 2

	jobs[0].Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	jobs[0].Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q1

	jobs[1].Config = &api.TrainingJob{}
	jobs[1].TrainerJob = &batchv1.Job{}
	jobs[1].TrainerJob.Spec.Parallelism = makePtr(2)
	jobs[1].Config.Name = "b"
	jobs[1].Config.Spec.Trainer.MinInstance = 1
	jobs[1].Config.Spec.Trainer.MaxInstance = 20

	jobs[2].Config = &api.TrainingJob{}
	jobs[2].TrainerJob = &batchv1.Job{}
	jobs[2].TrainerJob.Spec.Parallelism = makePtr(2)
	jobs[2].Config.Name = "c"
	jobs[2].Config.Spec.Trainer.MinInstance = 1
	jobs[2].Config.Spec.Trainer.MaxInstance = 10

	jobs[3].Config = &api.TrainingJob{}
	jobs[3].TrainerJob = &batchv1.Job{}
	jobs[3].TrainerJob.Spec.Parallelism = makePtr(2)
	jobs[3].Config.Name = "d"
	jobs[3].Config.Spec.Trainer.MinInstance = 1
	jobs[3].Config.Spec.Trainer.MaxInstance = 1

	expected := []string{"a"}

	c := New(nil)
	for _, j := range jobs {
		c.jobs[j.Config.Name] = j
	}

	sorted := sortedJobs(jobs, gpu)
	result := make([]string, len(sorted))
	for i, j := range sorted {
		result[i] = j.Config.Name
	}
	assert.Equal(t, expected, result)
}

func TestSortedJobsWithTie(t *testing.T) {
	jobs := make([]job, 4)
	jobs[0].Config = &api.TrainingJob{}
	jobs[0].TrainerJob = &batchv1.Job{}
	jobs[0].TrainerJob.Spec.Parallelism = makePtr(1)
	jobs[0].Config.Name = "a"
	jobs[0].Config.Spec.Trainer.MinInstance = 1
	jobs[0].Config.Spec.Trainer.MaxInstance = 2
	jobs[0].Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	jobs[0].Config.Spec.Trainer.Resources.Limits[api.GPUResourceName] = q1

	jobs[1].Config = &api.TrainingJob{}
	jobs[1].TrainerJob = &batchv1.Job{}
	jobs[1].TrainerJob.Spec.Parallelism = makePtr(1)
	jobs[1].Config.Name = "b"
	jobs[1].Config.Spec.Trainer.MinInstance = 1
	jobs[1].Config.Spec.Trainer.MaxInstance = 2
	jobs[1].Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	jobs[1].Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	jobs[1].Config.Spec.Trainer.Resources.Requests["memory"] = q1

	jobs[2].Config = &api.TrainingJob{}
	jobs[2].TrainerJob = &batchv1.Job{}
	jobs[2].TrainerJob.Spec.Parallelism = makePtr(1)
	jobs[2].Config.Name = "c"
	jobs[2].Config.Spec.Trainer.MinInstance = 1
	jobs[2].Config.Spec.Trainer.MaxInstance = 2
	q, err := resource.ParseQuantity("10")
	assert.Nil(t, err)
	jobs[2].Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	jobs[2].Config.Spec.Trainer.Resources.Requests["cpu"] = q

	jobs[3].Config = &api.TrainingJob{}
	jobs[3].TrainerJob = &batchv1.Job{}
	jobs[3].TrainerJob.Spec.Parallelism = makePtr(1)
	jobs[3].Config.Name = "d"
	jobs[3].Config.Spec.Trainer.MinInstance = 1
	jobs[3].Config.Spec.Trainer.MaxInstance = 2
	jobs[3].Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	jobs[3].Config.Spec.Trainer.Resources.Requests["cpu"] = q1
	q, err = resource.ParseQuantity("2")
	assert.Nil(t, err)
	jobs[3].Config.Spec.Trainer.Resources.Requests["memory"] = q

	expected := []string{"b", "d", "c", "a"}

	sorted := sortedJobs(jobs, elastic)
	result := make([]string, len(sorted))
	for i, j := range sorted {
		result[i] = j.Config.Name
	}
	assert.Equal(t, expected, result)

}
