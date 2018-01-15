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

package edl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestAddResourceList(t *testing.T) {
	cpu, _ := resource.ParseQuantity("200m")
	mem, _ := resource.ParseQuantity("500Mi")
	gpu, _ := resource.ParseQuantity("8")
	a := v1.ResourceList{
		"cpu":    cpu,
		"memory": mem,
	}

	b := v1.ResourceList{
		"cpu":                cpu,
		"memory":             mem,
		v1.ResourceNvidiaGPU: gpu,
	}

	cpuExpected, _ := resource.ParseQuantity("400m")
	memExpected, _ := resource.ParseQuantity("1000Mi")
	gpuExpected, _ := resource.ParseQuantity("8")

	AddResourceList(a, b)
	assert.Equal(t, a.Cpu().Value(), cpuExpected.Value())
	assert.Equal(t, a.Memory().Value(), memExpected.Value())
	assert.Equal(t, a.NvidiaGPU().Value(), gpuExpected.Value())
}
