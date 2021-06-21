// Copyright 2021 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	paddlev1 "github.com/paddleflow/paddle-operator/api/v1"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("PaddleJob controller", func() {

	const (
		jobNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	jobTypeMeta := metav1.TypeMeta{
		APIVersion: "batch.paddlepaddle.org/v1",
		Kind:       "PaddleJob",
	}

	podSpec := v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:  "paddle",
					Image: "registry.baidubce.com/kuizhiqing/demo-wide-and-deep:v1",
				},
			},
		},
	}

	wideAndDeepService := paddlev1.PaddleJob{
		TypeMeta: jobTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wide-and-deep-service",
			Namespace: jobNamespace,
		},
		Spec: paddlev1.PaddleJobSpec{
			CleanPodPolicy: paddlev1.CleanNever,
			Intranet:       paddlev1.Service,
			PS: paddlev1.ResourceSpec{
				Replicas: 3,
				Template: podSpec,
			},
			Worker: paddlev1.ResourceSpec{
				Replicas: 2,
				Template: podSpec,
			},
		},
	}

	test := func() {
		paddleJob := wideAndDeepService.DeepCopy()
		ctx := context.Background()
		jobKey := types.NamespacedName{
			Name:      paddleJob.ObjectMeta.Name,
			Namespace: paddleJob.ObjectMeta.Namespace,
		}
		createdPaddleJob := &paddlev1.PaddleJob{}

		makeStateTest := func(PsReplica, WorkerReplica int) func() bool {
			return func() bool {
				if err := k8sClient.Get(ctx, jobKey, createdPaddleJob); err != nil {
					return false
				}
				// Ensure pod number
				if len(createdPaddleJob.Status.PS.Refs) != PsReplica ||
					len(createdPaddleJob.Status.Worker.Refs) != WorkerReplica {
					return false
				}
				// TODO: add more test here
				return true
			}
		}

		Expect(k8sClient.Create(ctx, paddleJob)).Should(Succeed())
		Eventually(makeStateTest(3, 2), timeout, interval).Should(BeTrue())
		Expect(createdPaddleJob.Status.Mode).Should(Equal(paddlev1.PaddleJobModePS))

		createdPaddleJob.Spec.PS.Replicas = 1
		createdPaddleJob.Spec.Worker.Replicas = 4
		Expect(k8sClient.Update(ctx, createdPaddleJob)).Should(Succeed())
		Eventually(makeStateTest(1, 4), timeout, interval).Should(BeTrue())
	}

	It("test wide-and-deep in service mode", test)
})
