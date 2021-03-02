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

	resSpec := paddlev1.ResourceSpec{
		Replicas: 2,
		Template: podSpec,
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
			Worker:         resSpec,
			PS:             resSpec,
		},
	}

	test := func() {
		ctx := context.Background()
		Expect(k8sClient.Create(ctx, &wideAndDeepService)).Should(Succeed())

		jobKey := types.NamespacedName{
			Name:      wideAndDeepService.ObjectMeta.Name,
			Namespace: wideAndDeepService.ObjectMeta.Namespace,
		}
		createdPaddleJob := &paddlev1.PaddleJob{}

		Eventually(func() bool {
			err := k8sClient.Get(ctx, jobKey, createdPaddleJob)
			if err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())

		Eventually(func() int {
			err := k8sClient.Get(ctx, jobKey, createdPaddleJob)
			if err != nil {
				return -1
			}
			return len(createdPaddleJob.Status.PS.Refs)
		}, timeout, interval).Should(Equal(resSpec.Replicas))

	}

	It("test wide-and-deep", test)
})
