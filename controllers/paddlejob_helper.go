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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pdv1 "github.com/paddleflow/paddle-operator/api/v1"
	volcano "volcano.sh/apis/pkg/apis/scheduling/v1beta1"
)

const (
	schedulerNameVolcano         = "volcano"
	schedulingPodGroupAnnotation = "scheduling.k8s.io/group-name"

	coordContainerName = "coord-paddle"
	coordContainerCpu  = "10m"
	coordContainerMem  = "10m"
)

var (
	coordContainerCmd = []string{"sh", "-c", "while true; do if [ -f goon ]; then exit 0; else sleep 0.1; fi; done"}
)

func isAllPodsReady(pdj *pdv1.PaddleJob, childPods corev1.PodList) bool {
	if !isAllPodsCreated(pdj) {
		return false
	}
	for _, pod := range childPods.Items {
		if pod.Status.PodIP == "" {
			return false
		}
	}
	return true
}

func isAllPodsCreated(pdj *pdv1.PaddleJob) bool {
	specs := pdj.GetSpecs()
	statuses := pdj.GetStatuses()
	for k, _ := range specs {
		if !isPodCreated(specs[k], statuses[k]) {
			return false
		}
	}
	return true
}

func isPodCreated(spec *pdv1.ResourceSpec, status *pdv1.ResourceStatus) bool {
	if spec == nil {
		return true
	}
	if status != nil && len(status.Refs) == spec.Replicas {
		return true
	}
	return false
}

func isFailed(status *pdv1.ResourceStatus) bool {
	return status != nil && status.Failed > 0
}
func isPending(status *pdv1.ResourceStatus) bool {
	return status != nil && status.Pending > 0
}
func isStarting(status *pdv1.ResourceStatus) bool {
	return status != nil && status.Starting > 0
}
func isRunning(spec *pdv1.ResourceSpec, status *pdv1.ResourceStatus) bool {
	return spec == nil || (status != nil && spec.Replicas == status.Running)
}
func isCompleted(spec *pdv1.ResourceSpec, status *pdv1.ResourceStatus) bool {
	return spec == nil || (status != nil && spec.Replicas == status.Succeeded)
}

func getPaddleJobPhase(pdj *pdv1.PaddleJob) pdv1.PaddleJobPhase {

	// final phase won't change any more
	if pdj.Status.Phase == pdv1.Completed {
		return pdv1.Completed
	} else if pdj.Status.Phase == pdv1.Failed {
		return pdv1.Failed
	}

	specs := pdj.GetSpecs()
	statuses := pdj.GetStatuses()
	for _, status := range statuses {
		if isFailed(status) {
			return pdv1.Failed
		} else if isStarting(status) {
			return pdv1.Starting
		} else if isPending(status) {
			return pdv1.Pending
		}
	}
	checkAll := func(check func(spec *pdv1.ResourceSpec, status *pdv1.ResourceStatus) bool) bool {
		for k, _ := range statuses {
			if !check(specs[k], statuses[k]) {
				return false
			}
		}
		return true
	}
	if checkAll(isRunning) {
		return pdv1.Running
	}
	if checkAll(isCompleted) {
		return pdv1.Completed
	}

	if pdj.Status.Phase == "" {
		return pdv1.Pending
	}

	return pdj.Status.Phase
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

func isAllCoordContainerRunning(childPods corev1.PodList) bool {
	for i, _ := range childPods.Items {
		if !isCoordContainerRunning(&childPods.Items[i]) {
			return false
		}
	}
	return true
}

func isCoordContainerRunning(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodPending {
		return false
	}
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		if container.Name == coordContainerName && container.State.Running != nil {
			return true
		}
	}
	return false
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
	resources := map[string][]string{}

	specs := pdj.GetSpecs()
	for resType, spec := range specs {
		if spec != nil {
			resources[resType] = make([]string, spec.Replicas)
		}
	}

	for _, pod := range childPods.Items {
		if len(strings.Split(pod.Status.PodIP, ".")) != 4 {
			return nil
		}
		resType, idx := extractNameIndex(pod.Name)
		if pdj.Spec.Intranet == pdv1.Service {
			resources[resType][idx] = fmt.Sprintf("%s:%d", pod.Name, PADDLE_PORT)
		} else {
			resources[resType][idx] = fmt.Sprintf("%s:%d", pod.Status.PodIP, PADDLE_PORT)
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
		cm.Data["PADDLE_PSERVERS_IP_PORT_LIST"] = strings.Join(resources[pdv1.ResourcePS], ",")
	}
	if pdj.Spec.Worker != nil {
		cm.Data["PADDLE_TRAINER_ENDPOINTS"] = strings.Join(resources[pdv1.ResourceWorker], ",")
		cm.Data["PADDLE_TRAINERS"] = endpoints2hosts(resources[pdv1.ResourceWorker])
		cm.Data["PADDLE_TRAINERS_NUM"] = fmt.Sprintf("%d", pdj.Spec.Worker.Replicas)
	}
	if pdj.Spec.Heter != nil {
		cm.Data["PADDLE_HETER_ENDPOINTS"] = strings.Join(resources[pdv1.ResourceHeter], ",")
	}

	if pdj.Spec.WithGloo != nil && *pdj.Spec.WithGloo > 0 && len(resources[pdv1.ResourcePS]) > 0 {
		cm.Data["PADDLE_WITH_GLOO"] = fmt.Sprintf("%d", *pdj.Spec.WithGloo)
		cm.Data["PADDLE_GLOO_RENDEZVOUS"] = "3"
		cm.Data["PADDLE_GLOO_HTTP_ENDPOINT"] = strings.Replace(resources[pdv1.ResourcePS][0],
			fmt.Sprintf(":%d", PADDLE_PORT),
			fmt.Sprintf(":%d", PADDLE_PORT+HOST_PORT_NUM-2),
			1)
	}
	return cm
}

func constructPod(pdj *pdv1.PaddleJob, resType string, idx int) (pod *corev1.Pod) {
	name := genPaddleResName(pdj.Name, resType, idx)

	pod = &corev1.Pod{}
	specs := pdj.GetSpecs()

	pod.ObjectMeta = *specs[resType].Template.ObjectMeta.DeepCopy()
	pod.Spec = *specs[resType].Template.Spec.DeepCopy()

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

	pod.Spec.Hostname = name
	pod.Spec.Subdomain = name

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

func genCoordinateInitContainer(coordContainerImage string) corev1.Container {
	c := corev1.Container{
		Name:            coordContainerName,
		Image:           coordContainerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         coordContainerCmd,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(coordContainerCpu),
				corev1.ResourceMemory: resource.MustParse(coordContainerMem),
				//corev1.ResourceEphemeralStorage: resource.MustParse(),
			},
		},
	}
	return c
}

func endpoints2hosts(eps []string) string {
	hosts := make([]string, len(eps))
	for i, ep := range eps {
		p := strings.Split(ep, ":")
		hosts[i] = p[0]
	}
	return strings.Join(hosts, ",")
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
	specs := pdj.GetSpecs()
	for _, spec := range specs {
		if check(spec) {
			return true
		}
	}
	return false
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
	specs := pdj.GetSpecs()
	total := 0
	for _, spec := range specs {
		if spec != nil {
			total += spec.Replicas
		}
	}
	return int32(total)
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
	specs := pdj.GetSpecs()
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		for i := 0; i < spec.Replicas; i++ {
			for _, c := range spec.Template.Spec.Containers {
				if c.Resources.Requests != nil {
					addRes(totalRes, c.Resources.Requests)
				} else {
					addRes(totalRes, c.Resources.Limits)
				}
			}
		}
	}
	return &totalRes
}
