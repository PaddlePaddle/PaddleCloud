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
	q1    resource.Quantity
	q10   resource.Quantity
	q0    resource.Quantity
	q100M resource.Quantity
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
	q100M, err = resource.ParseQuantity("100Mi")
	if err != nil {
		panic(err)
	}
}

func makePtr(c int) *int32 {
	p := int32(c)
	return &p
}

func makeJob(name string, cpuReq, cpuLim, memReq, memLim, gpuLim string, min, max, p int) job {
	cr, err := resource.ParseQuantity(cpuReq)
	if err != nil {
		panic(err)
	}
	cl, err := resource.ParseQuantity(cpuLim)
	if err != nil {
		panic(err)
	}
	mr, err := resource.ParseQuantity(memReq)
	if err != nil {
		panic(err)
	}
	ml, err := resource.ParseQuantity(memLim)
	if err != nil {
		panic(err)
	}
	gl, err := resource.ParseQuantity(gpuLim)
	if err != nil {
		panic(err)
	}

	j := job{
		Config:     &api.TrainingJob{},
		TrainerJob: &batchv1.Job{},
	}
	j.Config.Name = name
	j.Config.Spec.Trainer.Resources.Requests = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Config.Spec.Trainer.Resources.Requests["cpu"] = cr
	j.Config.Spec.Trainer.Resources.Limits["cpu"] = cl
	j.Config.Spec.Trainer.Resources.Requests["memory"] = mr
	j.Config.Spec.Trainer.Resources.Limits["memory"] = ml
	j.Config.Spec.Trainer.Resources.Limits[v1.ResourceNvidiaGPU] = gl
	j.TrainerJob.Spec.Parallelism = makePtr(p)
	j.Config.Spec.Trainer.MinInstance = min
	j.Config.Spec.Trainer.MaxInstance = max
	return j
}

func TestTrainerRequestLimit(t *testing.T) {
	j := makeJob("name", "1k", "1k", "100Mi", "100Mi", "10", 1, 1, 1)
	assert.Equal(t, int64(1000000), j.TrainerCPURequestMilli())
	assert.Equal(t, int64(105), j.TrainerMemRequestMega())
	assert.Equal(t, 10, j.TrainerGPULimit())
}

func TestScaleDryRunSatisfied(t *testing.T) {
	r := ClusterResource{CPUTotalMilli: 2000, MemoryTotalMega: 1000}
	j := makeJob("name", "1000Mi", "1000Mi", "100Mi", "100Mi", "0", 1, 2, 2)
	assert.Equal(t, 0, scaleDryRun(&r, j, 0, false))
}

func TestScaleDryRunMoreCPU(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     100,
		CPURequestMilli:   100,
		CPUTotalMilli:     3000,
		MemoryRequestMega: 100,
		MemoryLimitMega:   100,
		MemoryTotalMega:   1000,
	}
	j := makeJob("name", "1", "1", "100Mi", "100Mi", "0", 1, 3, 1)
	assert.Equal(t, 1, scaleDryRun(&r, j, 0, false))
}

func TestScaleDryRunNoMoreCPU(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     1000,
		CPURequestMilli:   1000,
		CPUTotalMilli:     1000,
		MemoryRequestMega: 100,
		MemoryLimitMega:   100,
		MemoryTotalMega:   1000,
	}

	j := makeJob("name", "1", "1", "100Mi", "100Mi", "0", 1, 3, 1)
	assert.Equal(t, 0, scaleDryRun(&r, j, 0, false))
}

func TestScaleDryRunMoreGPU(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     0,
		CPURequestMilli:   0,
		CPUTotalMilli:     2000,
		MemoryRequestMega: 100,
		MemoryLimitMega:   100,
		MemoryTotalMega:   1000,
		GPULimit:          0,
		GPURequest:        0,
		GPUTotal:          10,
	}
	j := makeJob("name", "1", "1", "10Mi", "10Mi", "1", 1, 3, 1)
	assert.Equal(t, 1, scaleDryRun(&r, j, 0, false))
}

func TestScaleDryRunNoMoreGPU(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     0,
		CPURequestMilli:   0,
		CPUTotalMilli:     2000,
		MemoryRequestMega: 100,
		MemoryLimitMega:   100,
		MemoryTotalMega:   1000,
		GPULimit:          10,
		GPURequest:        10,
		GPUTotal:          10,
	}

	j := makeJob("name", "1", "1", "10Mi", "10Mi", "1", 1, 3, 1)
	assert.Equal(t, 0, scaleDryRun(&r, j, 0, false))
}

func TestScaleDryRunScaleDownMoreThanExpected(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     1000,
		CPURequestMilli:   1000,
		CPUTotalMilli:     1000,
		MemoryRequestMega: 1000,
		MemoryLimitMega:   1000,
		MemoryTotalMega:   1000,
		GPULimit:          10,
		GPURequest:        10,
		GPUTotal:          10,
	}

	j := makeJob("name", "1", "1", "10Mi", "10Mi", "0", 1, 3, 6)
	assert.Equal(t, -3, scaleDryRun(&r, j, 0, true))
}

func TestScaleDryRunScaleDownToMin(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     5000,
		CPURequestMilli:   5000,
		CPUTotalMilli:     3000,
		MemoryRequestMega: 1000,
		MemoryLimitMega:   1000,
		MemoryTotalMega:   1000,
		GPULimit:          10,
		GPURequest:        10,
		GPUTotal:          10,
	}

	j := makeJob("name", "1", "1", "10Mi", "10Mi", "0", 1, 3, 3)
	assert.Equal(t, -1, scaleDryRun(&r, j, 0, true))
	assert.Equal(t, -1, scaleDryRun(&r, j, -1, true))
	assert.Equal(t, 0, scaleDryRun(&r, j, -2, true))
}

func TestScaleDryRunScaleDownFullCluster(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     2000,
		CPURequestMilli:   2000,
		CPUTotalMilli:     1000,
		MemoryRequestMega: 1000,
		MemoryLimitMega:   1000,
		MemoryTotalMega:   1000,
		GPULimit:          10,
		GPURequest:        10,
		GPUTotal:          10,
	}

	j := makeJob("name", "1", "1", "10Mi", "10Mi", "0", 1, 3, 3)
	assert.Equal(t, -1, scaleDryRun(&r, j, 0, true))
}

func TestScaleDryRunNoMem(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     1000,
		CPURequestMilli:   1000,
		CPUTotalMilli:     1000,
		MemoryRequestMega: 1000,
		MemoryLimitMega:   1000,
		MemoryTotalMega:   1000,
		GPULimit:          10,
		GPURequest:        10,
		GPUTotal:          10,
	}

	j := makeJob("name", "1", "1", "100Mi", "100Mi", "0", 1, 3, 1)
	assert.Equal(t, 0, scaleDryRun(&r, j, 0, false))
}

func TestScaleAllDryRunNoMem(t *testing.T) {
	r := ClusterResource{
		CPUTotalMilli:     1000,
		MemoryRequestMega: 1000,
		MemoryLimitMega:   1000,
		MemoryTotalMega:   1000,
		GPUTotal:          10,
	}

	j := makeJob("name", "1", "1", "1", "1", "1", 1, 3, 1)
	scale := scaleAllDryRun([]job{j}, r)["name"]
	assert.Equal(t, 0, scale)
}

func TestScaleAllDryRun(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     1000,
		CPURequestMilli:   1000,
		CPUTotalMilli:     4000,
		MemoryRequestMega: 100,
		MemoryLimitMega:   100,
		MemoryTotalMega:   1000,
		GPULimit:          8,
		GPURequest:        8,
		GPUTotal:          10,
	}

	j := makeJob("name", "1", "1", "100Mi", "100Mi", "0", 1, 3, 1)
	scale := scaleAllDryRun([]job{j}, r)["name"]
	assert.Equal(t, 2, scale)
}

func TestScaleAllDryRunLessCPU(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     2000,
		CPURequestMilli:   2000,
		CPUTotalMilli:     3000,
		MemoryRequestMega: 100,
		MemoryLimitMega:   100,
		MemoryTotalMega:   1000,
		GPULimit:          8,
		GPURequest:        8,
		GPUTotal:          10,
	}

	j := makeJob("name", "1", "1", "1", "1", "1", 1, 3, 1)
	scale := scaleAllDryRun([]job{j}, r)["name"]
	assert.Equal(t, 1, scale)
}

func TestScaleAllDryRunLessGPU(t *testing.T) {
	r := ClusterResource{
		CPULimitMilli:     990,
		CPURequestMilli:   990,
		CPUTotalMilli:     2000,
		MemoryRequestMega: 100,
		MemoryLimitMega:   100,
		MemoryTotalMega:   1000,
		GPULimit:          9,
		GPURequest:        9,
		GPUTotal:          10,
	}

	j := makeJob("name", "1", "1", "1", "1", "1", 1, 3, 1)
	scale := scaleAllDryRun([]job{j}, r)["name"]
	assert.Equal(t, 1, scale)
}

func TestFulfillment(t *testing.T) {
	j := makeJob("name", "1", "1", "1", "1", "1", 1, 2, 2)
	assert.Equal(t, float64(1), j.Fulfillment())

	j = makeJob("name", "1", "1", "1", "1", "1", 1, 2, 1)
	assert.Equal(t, float64(0), j.Fulfillment())

	j = makeJob("name", "1", "1", "1", "1", "1", 1, 3, 2)
	assert.Equal(t, float64(0.5), j.Fulfillment())
}

func TestSortedJobs(t *testing.T) {
	jobs := []job{
		makeJob("a", "1", "1", "1", "1", "1", 1, 2, 2),
		makeJob("b", "1", "1", "1", "1", "1", 1, 20, 2),
		makeJob("c", "1", "1", "1", "1", "1", 1, 10, 2),
		makeJob("d", "1", "1", "1", "1", "1", 1, 1, 2),
	}

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
	jobs := []job{
		makeJob("a", "1", "1", "1", "1", "1", 1, 2, 2),
		makeJob("b", "1", "1", "1", "1", "0", 1, 20, 2),
		makeJob("c", "1", "1", "1", "1", "0", 1, 10, 2),
		makeJob("d", "1", "1", "1", "1", "0", 1, 1, 2),
	}

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
	jobs := []job{
		makeJob("a", "1", "0", "1", "1", "1", 1, 2, 1),
		makeJob("b", "1", "1", "1", "1", "0", 1, 2, 1),
		makeJob("c", "10", "10", "1", "1", "0", 1, 2, 1),
		makeJob("d", "1", "1", "2", "2", "0", 1, 2, 1),
	}
	expected := []string{"b", "d", "c", "a"}

	sorted := sortedJobs(jobs, elastic)
	result := make([]string, len(sorted))
	for i, j := range sorted {
		result[i] = j.Config.Name
	}
	assert.Equal(t, expected, result)

}
