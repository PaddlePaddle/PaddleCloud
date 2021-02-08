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
	"strconv"
	"strings"

	pdv1 "github.com/paddleflow/paddle-operator/api/v1"
)

func getPaddleJobPhase(pdj *pdv1.PaddleJob) pdv1.PaddleJobPhase {
	if pdj.Status.Phase == pdv1.Completed {
		return pdv1.Completed
	} else if pdj.Status.Phase == pdv1.Failed {
		return pdv1.Failed
	} else if pdj.Spec.PS.Replicas == pdj.Status.PS.Running && pdj.Spec.Worker.Replicas == pdj.Status.Worker.Running {
		return pdv1.Running
	} else if pdj.Status.PS.Failed > 0 || pdj.Status.Worker.Failed > 0 {
		return pdv1.Failed
	} else if pdj.Spec.PS.Replicas >= pdj.Status.PS.Succeeded && pdj.Spec.Worker.Replicas == pdj.Status.Worker.Succeeded {
		return pdv1.Completed
	} else if pdj.Status.PS.Pending > 0 || pdj.Status.Worker.Pending > 0 {
		return pdv1.Starting
	}
	return pdv1.Starting
}

func getPaddleJobMode(pdj *pdv1.PaddleJob) pdv1.PaddleJobMode {
	if pdj.Spec.PS.Replicas > 0 {
		return pdv1.PaddleJobModePS
	} else if pdj.Spec.Worker.Replicas > 0 {
		return pdv1.PaddleJobModeCollective
	} else {
		return pdv1.PaddleJobModeSingle
	}
}

// genPaddleResName generate the identifier for pod and service
func genPaddleResName(name string, resType string, idx int) string {
	return fmt.Sprintf("%s-%s-%d", name, resType, idx)
}

func extractNameIndex(name string) (string, int) {
	s := strings.Split(name, "-")
	if i, err := strconv.Atoi(s[len(s)-1]); err != nil {
		return "", 0
	} else {
		return s[len(s)-2], i
	}
}

func constructPS4PaddleJob(pdj *pdv1.PaddleJob, idx int) *corev1.Pod {
	name := genPaddleResName(pdj.Name, pdv1.ResourcePS, idx)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				pdv1.ResourceName: name,
				pdv1.ResourceType: pdv1.ResourcePS,
			},
			Annotations: map[string]string{
				pdv1.ResourceAnnotation: pdv1.ResourcePS,
			},
			Name:      genPaddleResName(pdj.Name, pdv1.ResourcePS, idx),
			Namespace: pdj.Namespace,
		},
		Spec: *pdj.Spec.Worker.Template.Spec.DeepCopy(),
	}
	envs := map[string]string{
		"PADDLE_PSERVERS_IP_PORT_LIST":      genEndpoints(pdj.Name, pdv1.ResourcePS, pdj.Spec.PS.Replicas, pdv1.PADDLE_PORT),
		"PADDLE_TRAINERS_NUM":               fmt.Sprintf("%d", pdj.Spec.Worker.Replicas),
		"TRAINING_ROLE":                     pdv1.TrainingRole[pdv1.ResourcePS],
		"PADDLE_HETER_TRAINER_IP_PORT_LIST": "",
		"PADDLE_PORT":                       fmt.Sprintf("%d", pdv1.PADDLE_PORT),
		"POD_IP":                            name,
	}
	for k, v := range envs {
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, corev1.EnvVar{Name: k, Value: v})
	}
	pod.Spec.Containers[0].Ports = append(pod.Spec.Containers[0].Ports, corev1.ContainerPort{ContainerPort: pdv1.PADDLE_PORT})
	pod.Spec.RestartPolicy = "Never"
	return pod
}

func constructWorker4PaddleJob(pdj *pdv1.PaddleJob, idx int) *corev1.Pod {
	name := genPaddleResName(pdj.Name, pdv1.ResourceWorker, idx)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				pdv1.ResourceName: name,
				pdv1.ResourceType: pdv1.ResourceWorker,
			},
			Annotations: map[string]string{
				pdv1.ResourceAnnotation: pdv1.ResourceWorker,
			},
			Name:      name,
			Namespace: pdj.Namespace,
		},
		Spec: *pdj.Spec.Worker.Template.Spec.DeepCopy(),
	}
	// ugly env, hard to change
	envs := map[string]string{
		"PADDLE_PSERVERS_IP_PORT_LIST":      genEndpoints(pdj.Name, pdv1.ResourcePS, pdj.Spec.PS.Replicas, pdv1.PADDLE_PORT),
		"PADDLE_TRAINERS_NUM":               fmt.Sprintf("%d", pdj.Spec.Worker.Replicas),
		"TRAINING_ROLE":                     pdv1.TrainingRole[pdv1.ResourceWorker],
		"PADDLE_HETER_TRAINER_IP_PORT_LIST": "",
		"PADDLE_TRAINER_ID":                 fmt.Sprintf("%d", idx),
		"PADDLE_TRAINING_ROLE":              pdv1.TrainingRole[pdv1.ResourceWorker],
		"PADDLE_TRAINER_ENDPOINTS":          genEndpoints(pdj.Name, pdv1.ResourcePS, pdj.Spec.Worker.Replicas, pdv1.PADDLE_PORT),
		"PADDLE_CURRENT_ENDPOINT":           fmt.Sprintf("%s:%d", name, pdv1.PADDLE_PORT),
	}
	for k, v := range envs {
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, corev1.EnvVar{Name: k, Value: v})
	}
	pod.Spec.Containers[0].Ports = append(pod.Spec.Containers[0].Ports, corev1.ContainerPort{ContainerPort: pdv1.PADDLE_PORT})
	pod.Spec.RestartPolicy = "Never"
	return pod
}

func constructService4Pod(pod corev1.Pod) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels:    map[string]string{},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port: pdv1.PADDLE_PORT,
				},
			},
			Selector: map[string]string{
				pdv1.ResourceName: pod.Name,
			},
			ClusterIP: "None",
		},
	}
	return svc
}

func genEndpoints(name string, resType string, num int, port int) string {
	ret := []string{}
	for i := 0; i < num; i++ {
		name := genPaddleResName(name, resType, i)
		ret = append(ret, fmt.Sprintf("%s:%d", name, port))
	}
	return strings.Join(ret, ",")
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func isPodRealRuning(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		if !container.Ready || container.State.Running == nil {
			return false
		}
	}
	for i := range pod.Status.ContainerStatuses {
		container := pod.Status.ContainerStatuses[i]
		if !container.Ready || container.State.Running == nil {
			return false
		}
	}
	return true
}
