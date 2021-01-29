/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pdv1 "github.com/paddleflow/paddle-operator/api/v1"
)

func getPaddleJobMode(pdj *pdv1.PaddleJob) pdv1.PaddleJobMode {
	if pdj.Spec.PS.Replicas > 0 {
		return pdv1.PaddleJobModePS
	} else if pdj.Spec.Worker.Replicas > 0 {
		return pdv1.PaddleJobModeCollective
	} else {
		return pdv1.PaddleJobModeSingle
	}
}

func genPaddlePodName(name string, resType string, idx int) string {
	return fmt.Sprintf("%s-%s-%d", name, resType, idx)
}

func constructPS4PaddleJob(pdj *pdv1.PaddleJob, idx int) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        genPaddlePodName(pdj.Name, pdv1.ResourcePS, idx),
			Namespace:   pdj.Namespace,
		},
		Spec: *pdj.Spec.Worker.Template.Spec.DeepCopy(),
	}
	pod.Annotations[pdv1.ResourceAnnotation] = pdv1.ResourcePS
	return pod, nil
}

func constructWorker4PaddleJob(pdj *pdv1.PaddleJob, idx int) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        genPaddlePodName(pdj.Name, pdv1.ResourceWorker, idx),
			Namespace:   pdj.Namespace,
		},
		Spec: *pdj.Spec.Worker.Template.Spec.DeepCopy(),
	}
	pod.Annotations[pdv1.ResourceAnnotation] = pdv1.ResourceWorker
	return pod, nil
}

func constructService4Pod(pod *corev1.Pod) (*corev1.Service, error) {
	svc := &corev1.Service{}
	return svc, nil
}
