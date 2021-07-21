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
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pdv1 "github.com/paddleflow/paddle-operator/api/v1"
	volcano "volcano.sh/apis/pkg/apis/scheduling/v1beta1"
)

const (
	initContainerName            = "init-paddle"
	schedulerNameVolcano         = "volcano"
	schedulingPodGroupAnnotation = "scheduling.k8s.io/group-name"
)

func isAllPodsReady(pdj *pdv1.PaddleJob) bool {
	return isPodReady(pdj.Spec.PS, pdj.Status.PS) &&
		isPodReady(pdj.Spec.Worker, pdj.Status.Worker) &&
		isPodReady(pdj.Spec.Heter, pdj.Status.Heter)
}

func isPodReady(spec *pdv1.ResourceSpec, status *pdv1.ResourceStatus) bool {
	if spec == nil {
		return true
	}
	if status != nil && len(status.Refs) == spec.Replicas {
		return true
	}
	return false
}

func getPaddleJobPhase(pdj *pdv1.PaddleJob) pdv1.PaddleJobPhase {

	if pdj.Status.Phase == pdv1.Completed {
		return pdv1.Completed
	} else if pdj.Status.Phase == pdv1.Failed {
		return pdv1.Failed
	} else if (pdj.Status.PS != nil && pdj.Status.PS.Failed > 0) ||
		(pdj.Status.Worker != nil && pdj.Status.Worker.Failed > 0) ||
		(pdj.Status.Heter != nil && pdj.Status.Heter.Failed > 0) {
		return pdv1.Failed
	} else if (pdj.Spec.PS == nil || (pdj.Status.PS != nil && pdj.Spec.PS.Replicas == pdj.Status.PS.Running)) &&
		(pdj.Spec.Worker == nil || (pdj.Status.Worker != nil && pdj.Spec.Worker.Replicas == pdj.Status.Worker.Running)) &&
		(pdj.Spec.Heter == nil || (pdj.Status.Heter != nil && pdj.Spec.Heter.Replicas == pdj.Status.Heter.Running)) {
		return pdv1.Running
	} else if (pdj.Spec.PS == nil || (pdj.Status.PS != nil && pdj.Spec.PS.Replicas == pdj.Status.PS.Succeeded)) &&
		(pdj.Spec.Worker == nil || (pdj.Status.Worker != nil && pdj.Spec.Worker.Replicas == pdj.Status.Worker.Succeeded)) &&
		(pdj.Spec.Heter == nil || (pdj.Status.Heter != nil && pdj.Spec.Heter.Replicas == pdj.Status.Heter.Succeeded)) {
		return pdv1.Completed
	} else if (pdj.Status.PS != nil && pdj.Status.PS.Pending > 0) ||
		(pdj.Status.Worker != nil && pdj.Status.Worker.Pending > 0) ||
		(pdj.Status.Heter != nil && pdj.Status.Heter.Pending > 0) {
		return pdv1.Pending
	} else if (pdj.Status.PS != nil && pdj.Status.PS.Starting > 0) ||
		(pdj.Status.Worker != nil && pdj.Status.Worker.Starting > 0) ||
		(pdj.Status.Heter != nil && pdj.Status.Heter.Starting > 0) {
		return pdv1.Starting
	}

	return pdv1.Pending
}

func getPaddleJobStartTime(pdj *pdv1.PaddleJob) *metav1.Time {
	if pdj.Status.StartTime.IsZero() && pdj.Status.Phase == pdv1.Running {
		tmp := metav1.Now()
		return &tmp
	}
	return pdj.Status.StartTime
}

func getPaddleJobCompleteTime(pdj *pdv1.PaddleJob) *metav1.Time {
	if pdj.Status.CompletionTime.IsZero() && (pdj.Status.Phase == pdv1.Completed || pdj.Status.Phase == pdv1.Failed) {
		tmp := metav1.Now()
		return &tmp
	}
	return pdj.Status.CompletionTime
}

func getPaddleJobMode(pdj *pdv1.PaddleJob) pdv1.PaddleJobMode {
	if pdj.Spec.PS != nil {
		return pdv1.PaddleJobModePS
	} else if pdj.Spec.Worker != nil && pdj.Spec.Worker.Replicas > 1 {
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
	var pservers []string
	if pdj.Spec.PS != nil {
		pservers = make([]string, pdj.Spec.PS.Replicas)
	}

	var workers, workerHosts []string
	if pdj.Spec.Worker != nil {
		workers = make([]string, pdj.Spec.Worker.Replicas)
		workerHosts = make([]string, pdj.Spec.Worker.Replicas)
	}

	var heters []string
	if pdj.Spec.Heter != nil {
		heters = make([]string, pdj.Spec.Heter.Replicas)
	}

	for _, pod := range childPods.Items {
		if len(strings.Split(pod.Status.PodIP, ".")) != 4 {
			return nil
		}
		resType, idx := extractNameIndex(pod.Name)
		if pdj.Spec.Intranet == pdv1.Service {
			switch resType {
			case pdv1.ResourcePS:
				pservers[idx] = fmt.Sprintf("%s:%d", pod.Name, PADDLE_PORT)
			case pdv1.ResourceWorker:
				workers[idx] = fmt.Sprintf("%s:%d", pod.Name, PADDLE_PORT)
				workerHosts[idx] = fmt.Sprintf("%s", pod.Name)
			case pdv1.ResourceHeter:
				heters[idx] = fmt.Sprintf("%s:%d", pod.Name, PADDLE_PORT)
			}
		} else {
			switch resType {
			case pdv1.ResourcePS:
				pservers[idx] = fmt.Sprintf("%s:%d", pod.Status.PodIP, PADDLE_PORT)
			case pdv1.ResourceWorker:
				workers[idx] = fmt.Sprintf("%s:%d", pod.Status.PodIP, PADDLE_PORT)
				workerHosts[idx] = fmt.Sprintf("%s", pod.Status.PodIP)
			case pdv1.ResourceHeter:
				heters[idx] = fmt.Sprintf("%s:%d", pod.Status.PodIP, PADDLE_PORT)
			}
		}
	}
	var paddle_port string
	if pdj.Spec.Intranet == pdv1.HostNetwork {
		paddle_port = pdj.ObjectMeta.Annotations[hostPort]
	} else {
		paddle_port = fmt.Sprintf("%d", PADDLE_PORT)
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
			"TRAINER_PORTS_NUM": fmt.Sprintf("%d", HOST_PORT_NUM),
			"PADDLE_PORT":       paddle_port,
		},
	}
	if pdj.Spec.PS != nil {
		cm.Data["PADDLE_PSERVERS_IP_PORT_LIST"] = strings.Join(pservers, ",")
	}
	if pdj.Spec.Worker != nil {
		cm.Data["PADDLE_TRAINER_ENDPOINTS"] = strings.Join(workers, ",")
		cm.Data["PADDLE_TRAINERS"] = strings.Join(workerHosts, ",")
		cm.Data["PADDLE_TRAINERS_NUM"] = fmt.Sprintf("%d", pdj.Spec.Worker.Replicas)
	}
	if pdj.Spec.Heter != nil {
		cm.Data["PADDLE_HETER_ENDPOINTS"] = strings.Join(heters, ",")
	}

	if pdj.Spec.WithGloo != nil && *pdj.Spec.WithGloo > 0 && pdj.Spec.Intranet != pdv1.Service && len(pservers) > 0 {
		cm.Data["PADDLE_WITH_GLOO"] = fmt.Sprintf("%d", *pdj.Spec.WithGloo)
		cm.Data["PADDLE_GLOO_RENDEZVOUS"] = "3"
		cm.Data["PADDLE_GLOO_HTTP_ENDPOINT"] = strings.Replace(pservers[0],
			fmt.Sprintf(":%d", PADDLE_PORT),
			fmt.Sprintf(":%d", PADDLE_PORT+HOST_PORT_NUM-2),
			1)
	}
	return cm
}

func constructPod(pdj *pdv1.PaddleJob, resType string, idx int) (pod *corev1.Pod) {
	name := genPaddleResName(pdj.Name, resType, idx)

	pod = &corev1.Pod{}
	if resType == pdv1.ResourcePS {
		pod.ObjectMeta = *pdj.Spec.PS.Template.ObjectMeta.DeepCopy()
		pod.Spec = *pdj.Spec.PS.Template.Spec.DeepCopy()
	} else if resType == pdv1.ResourceWorker {
		pod.ObjectMeta = *pdj.Spec.Worker.Template.ObjectMeta.DeepCopy()
		pod.Spec = *pdj.Spec.Worker.Template.Spec.DeepCopy()
	} else if resType == pdv1.ResourceHeter {
		pod.ObjectMeta = *pdj.Spec.Heter.Template.ObjectMeta.DeepCopy()
		pod.Spec = *pdj.Spec.Heter.Template.Spec.DeepCopy()
	} else {
		return nil
	}

	if pod.ObjectMeta.Labels == nil {
		pod.ObjectMeta.Labels = map[string]string{}
	}
	pod.ObjectMeta.Labels[pdv1.ResourceName] = name
	pod.ObjectMeta.Labels[pdv1.ResourceType] = resType

	if pod.ObjectMeta.Annotations == nil {
		pod.ObjectMeta.Annotations = map[string]string{}
	}
	pod.ObjectMeta.Annotations[pdv1.ResourceAnnotation] = resType

	pod.ObjectMeta.Name = name
	pod.ObjectMeta.Namespace = pdj.Namespace

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

	if pdj.Spec.Elastic != nil {
		envJobID := corev1.EnvVar{
			Name:  "PADDLE_ELASTIC_JOB_ID",
			Value: fmt.Sprintf("%s-%s", pdj.Namespace, pdj.Name),
		}
		envNP := corev1.EnvVar{
			Name:  "PADDLE_ELASTIC_NP",
			Value: fmt.Sprintf("%d", pdj.Spec.Worker.Replicas),
		}
		envTimeout := corev1.EnvVar{
			Name:  "PADDLE_ELASTIC_TIMEOUT",
			Value: "60",
		}

		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, envJobID, envNP, envTimeout)
	} else {
		envF := corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: pdj.Name,
				},
			},
		}

		pod.Spec.Containers[0].EnvFrom = append(pod.Spec.Containers[0].EnvFrom, envF)
	}

	if pdj.Spec.Intranet == pdv1.Service {
		pod.Spec.Containers[0].Ports = append(pod.Spec.Containers[0].Ports, corev1.ContainerPort{ContainerPort: PADDLE_PORT})
	} else if pdj.Spec.Intranet == pdv1.HostNetwork {
		pod.Spec.HostNetwork = true
	}

	if pdj.Spec.Elastic != nil {
		pod.Spec.RestartPolicy = "OnFailure"
	} else if pod.Spec.RestartPolicy == "" {
		if resType == pdv1.ResourceWorker && pdj.Spec.Intranet == pdv1.Service {
			pod.Spec.RestartPolicy = "OnFailure"
		} else {
			pod.Spec.RestartPolicy = "Never"
		}
	}

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
	var ports = []corev1.ServicePort{}
	for i := 0; i < HOST_PORT_NUM; i++ {
		ports = append(ports, corev1.ServicePort{
			Name: fmt.Sprintf("p-%d", i),
			Port: int32(PADDLE_PORT + i),
		})
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels:    map[string]string{},
		},
		Spec: corev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				pdv1.ResourceName: pod.Name,
			},
			ClusterIP: "None",
		},
	}
	return svc
}

// for volcano

func withoutVolcano(pdj *pdv1.PaddleJob) bool {
	check := func(rs *pdv1.ResourceSpec) bool {
		if rs != nil &&
			rs.Template.Spec.SchedulerName != "" &&
			rs.Template.Spec.SchedulerName != schedulerNameVolcano {
			return true
		} else {
			return false
		}
	}
	if check(pdj.Spec.PS) || check(pdj.Spec.Worker) || check(pdj.Spec.Heter) {
		return true
	} else {
		return false
	}
}

func constructPodGroup(pdj *pdv1.PaddleJob) *volcano.PodGroup {
	pg := &volcano.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pdj.Namespace,
			Name:      pdj.Name,
		},
	}

	pg.Spec.MinMember = getTotalReplicas(pdj)
	pg.Spec.MinResources = getPGMinResource(pdj)

	if pdj.Spec.SchedulingPolicy != nil {
		// minAvailable specified by user which not equals to total replicas
		// DO NOT make sense in current paddle scenario
		if pdj.Spec.SchedulingPolicy.MinAvailable != nil {
			pg.Spec.MinMember = *pdj.Spec.SchedulingPolicy.MinAvailable
		}
		if pdj.Spec.SchedulingPolicy.Queue != "" {
			pg.Spec.Queue = pdj.Spec.SchedulingPolicy.Queue
		}
		if pdj.Spec.SchedulingPolicy.PriorityClass != "" {
			pg.Spec.PriorityClassName = pdj.Spec.SchedulingPolicy.PriorityClass
		}
		if pdj.Spec.SchedulingPolicy.MinResources != nil {
			pg.Spec.MinResources = &pdj.Spec.SchedulingPolicy.MinResources
		}
	}

	return pg
}

func getTotalReplicas(pdj *pdv1.PaddleJob) int32 {
	count := func(rs *pdv1.ResourceSpec) int32 {
		if rs != nil {
			return int32(rs.Replicas)
		} else {
			return 0
		}
	}
	return count(pdj.Spec.PS) + count(pdj.Spec.Worker) + count(pdj.Spec.Heter)
}

func getPGMinResource(pdj *pdv1.PaddleJob) *corev1.ResourceList {
	addRes := func(crr, res corev1.ResourceList) {
		for name, quantity := range res {
			if value, ok := crr[name]; !ok {
				crr[name] = quantity.DeepCopy()
			} else {
				value.Add(quantity)
				crr[name] = value
			}
		}
	}
	// consider only the case minMember == minAvailable
	totalRes := corev1.ResourceList{}
	countRes := func(rs *pdv1.ResourceSpec) {
		if rs == nil {
			return
		}
		for i := 0; i < rs.Replicas; i++ {
			for _, c := range rs.Template.Spec.Containers {
				if c.Resources.Requests != nil {
					addRes(totalRes, c.Resources.Requests)
				} else {
					addRes(totalRes, c.Resources.Limits)
				}
			}
		}
	}
	countRes(pdj.Spec.PS)
	countRes(pdj.Spec.Worker)
	countRes(pdj.Spec.Heter)
	return &totalRes
}
