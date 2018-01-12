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

package resource_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/PaddlePaddle/cloud/go/api"
	"github.com/stretchr/testify/assert"
)

func TestNeedGPU(t *testing.T) {
	var j api.TrainingJob
	assert.False(t, j.NeedGPU())

	q, err := resource.ParseQuantity("1")
	assert.Nil(t, err)

	j.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Spec.Trainer.Resources.Limits[v1.ResourceNvidiaGPU] = q
	assert.True(t, j.NeedGPU())
}

func TestElastic(t *testing.T) {
	var j api.TrainingJob
	assert.False(t, j.Elastic())

	j.Spec.Trainer.MinInstance = 1
	j.Spec.Trainer.MaxInstance = 2
	assert.True(t, j.Elastic())
}
