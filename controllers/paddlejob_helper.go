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
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pdv1 "github.com/paddleflow/paddle-operator/api/v1"
)

const (
	initContainerName            = "init-paddle"
	schedulingPodGroupAnnotation = "scheduling.k8s.io/group-name"
	schedulerNameVolcano         = "volcano"
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

func constructConfigMap(pdj *pdv1.PaddleJob, childPods corev1.PodList) (cm *corev1.ConfigMap) {
	pservers := make([]string, pdj.Spec.PS.Replicas)
	workers := make([]string, pdj.Spec.Worker.Replicas)

	for _, pod := range childPods.Items {
		if len(strings.Split(pod.Status.PodIP, ".")) != 4 {
			return nil
		}
		resType, idx := extractNameIndex(pod.Name)
		if resType == pdv1.ResourcePS {
			if pdj.Spec.Intranet == pdv1.Service {
				pservers[idx] = fmt.Sprintf("%s:%d", pod.Name, pdv1.PADDLE_PORT)
			} else {
				pservers[idx] = fmt.Sprintf("%s:%d", pod.Status.PodIP, pdv1.PADDLE_PORT)
			}
		} else if resType == pdv1.ResourceWorker {
			if pdj.Spec.Intranet == pdv1.Service {
				workers[idx] = fmt.Sprintf("%s:%d", pod.Name, pdv1.PADDLE_PORT)
			} else {
				workers[idx] = fmt.Sprintf("%s:%d", pod.Status.PodIP, pdv1.PADDLE_PORT)
			}
		}
	}
	cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				pdv1.ResourceName: pdj.Name,
			},
			Annotations: map[string]string{},
			Name:        pdj.Name,
			Namespace:   pdj.Namespace,
		},
		Data: map[string]string{
			"PADDLE_PSERVERS_IP_PORT_LIST":      strings.Join(pservers, ","),
			"PADDLE_TRAINERS_NUM":               fmt.Sprintf("%d", pdj.Spec.Worker.Replicas),
			"PADDLE_HETER_TRAINER_IP_PORT_LIST": "",
			"PADDLE_PORT":                       fmt.Sprintf("%d", pdv1.PADDLE_PORT),
			"PADDLE_TRAINER_ENDPOINTS":          strings.Join(workers, ","),
		},
	}
	if pdj.Spec.WithGloo > 0 && pdj.Spec.Intranet != pdv1.Service {
		cm.Data["PADDLE_WITH_GLOO"] = fmt.Sprintf("%d", pdj.Spec.WithGloo)
		cm.Data["PADDLE_GLOO_RENDEZVOUS"] = "3"
		cm.Data["PADDLE_GLOO_HTTP_ENDPOINT"] = strings.Replace(pservers[0],
			fmt.Sprintf(":%d", pdv1.PADDLE_PORT),
			fmt.Sprintf(":%d", pdv1.PADDLE_PORT+10),
			1)
	}
	return cm
}

func constructPod(pdj *pdv1.PaddleJob, resType string, idx int) (pod *corev1.Pod) {
	// metadata is missing due to controller-gen,
	// c.f. https://github.com/kubernetes-sigs/controller-tools/issues/448
	// c.f. https://github.com/kubernetes-sigs/controller-tools/pull/539
	name := genPaddleResName(pdj.Name, resType, idx)
	pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				pdv1.ResourceName: name,
				pdv1.ResourceType: resType,
			},
			Annotations: map[string]string{
				pdv1.ResourceAnnotation: resType,
			},
			Name:      name,
			Namespace: pdj.Namespace,
		},
	}
	if resType == pdv1.ResourcePS {
		pod.Spec = *pdj.Spec.PS.Template.Spec.DeepCopy()
	} else {
		pod.Spec = *pdj.Spec.Worker.Template.Spec.DeepCopy()
	}

	if pdj.Spec.Worker.Template.Spec.SchedulerName == schedulerNameVolcano {
		pod.ObjectMeta.Annotations[schedulingPodGroupAnnotation] = pdj.Name
		pod.Spec.SchedulerName = schedulerNameVolcano
	}

	// TODO(kuizhiqing)
	// initContainer will ensure pods are ready, then create cm to remove resource not ready error
	// Now it simply wait, since kubernetes ensure cm created before pod running indeed
	ic := corev1.Container{
		Name:    "init-paddle",
		Image:   "busybox:1.28",
		Command: []string{"sh", "-c", "sleep 12"},
	}
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, ic)

	envIP := corev1.EnvVar{
		Name: "POD_IP",
	}
	if pdj.Spec.Intranet == pdv1.Service {
		envIP.Value = name
	} else {
		envIP.ValueFrom = &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		}
	}
	envRank := corev1.EnvVar{
		Name:  "PADDLE_TRAINER_ID",
		Value: fmt.Sprintf("%d", idx),
	}
	envRole := corev1.EnvVar{
		Name:  "TRAINING_ROLE",
		Value: pdv1.TrainingRole[resType],
	}
	envRole2 := corev1.EnvVar{
		Name:  "PADDLE_TRAINING_ROLE",
		Value: pdv1.TrainingRole[resType],
	}
	pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, envIP, envRank, envRole, envRole2)

	envF := corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: pdj.Name,
			},
		},
	}
	pod.Spec.Containers[0].EnvFrom = append(pod.Spec.Containers[0].EnvFrom, envF)

	if pdj.Spec.Intranet == pdv1.Service {
		pod.Spec.Containers[0].Ports = append(pod.Spec.Containers[0].Ports, corev1.ContainerPort{ContainerPort: pdv1.PADDLE_PORT})
	}
	pod.Spec.RestartPolicy = "Never"

	return pod
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
		if !container.Ready {
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

func isPodInitializing(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodPending {
		return false
	}
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		if container.Name == initContainerName && container.State.Running != nil {
			return true
		}
	}
	return false
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
				{
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
