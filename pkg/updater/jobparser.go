/* Copyright (c) 2016 PaddlePaddle Authors All Rights Reserved.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
	 limitations under the License. */

package updater

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	paddlev1 "github.com/paddleflow/elastictraining/pkg/apis/paddlepaddle/v1alpha1"
)

type PodType string

const (
	PSERVER      PodType = "pserver"
	TRAINER      PodType = "trainer"
	MASTER       PodType = "master"
	TrainerLabel         = "paddle-job"
	PserverLabel         = "paddle-job-pserver"
	MasterLabel          = "paddle-job-master"
)

const (
	imagePullPolicy = "Always"
)

// DefaultJobParser implement a basic JobParser.
type DefaultJobParser struct {
}

// setDefaultAndValidate updates default values for the added job and validates the fields.
func setDefaultAndValidate(job *paddlev1.TrainingJob) error {
	// Fill in default values
	// FIXME: Need to test. What is the value if specified "omitempty"
	if job.Spec.Port == 0 {
		job.Spec.Port = 7164
	}
	if job.Spec.PortsNum == 0 {
		job.Spec.PortsNum = 1
	}
	if job.Spec.PortsNumForSparse == 0 {
		job.Spec.PortsNumForSparse = 1
	}
	if job.Spec.Image == "" {
		job.Spec.Image = "paddlepaddle/paddlecloud-job:0.11.0"
	}
	if job.Spec.Passes == 0 {
		job.Spec.Passes = 1
	}

	if !job.Spec.FaultTolerant && job.Elastic() {
		return errors.New("max-instances should be equal to min-instances when fault_tolerant is disabled")
	}
	// TODO: add validations.(helin)
	return nil
}

// NewTrainingJob generates a whole structure of TrainingJob
func (p *DefaultJobParser) NewTrainingJob(job *paddlev1.TrainingJob) (*paddlev1.TrainingJob, error) {
	if err := setDefaultAndValidate(job); err != nil {
		return nil, err
	}

	useHostNetwork := job.Spec.HostNetwork
	if job.Spec.FaultTolerant {
		job.Spec.Master.ReplicaSpec = parseToMaster(job)
		if useHostNetwork {
			job.Spec.Master.ReplicaSpec.Spec.Template.Spec.HostNetwork = true
		}
	}
	job.Spec.Pserver.ReplicaSpec = parseToPserver(job, nil, true)
	job.Spec.Trainer.ReplicaSpec = parseToTrainer(job, nil, true)
	if useHostNetwork {
		job.Spec.Pserver.ReplicaSpec.Spec.Template.Spec.HostNetwork = true
		job.Spec.Trainer.ReplicaSpec.Spec.Template.Spec.HostNetwork = true
	}
	return job, nil
}

// parseToPserver generate a pserver replicaset resource according to "TrainingJob" resource specs.
func parseToPserver(job *paddlev1.TrainingJob, extraEnv []corev1.EnvVar, outter bool) *v1beta1.ReplicaSet {
	replicas := int32(job.Spec.Pserver.MinInstance)
	envs := podEnv(job, job.Spec.Pserver.Envs)
	envs = append(envs, extraEnv...)
	spec := &v1beta1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "extensions/v1beta1",
			APIVersion: "ReplicaSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.ObjectMeta.Name + "-pserver",
			Namespace: job.ObjectMeta.Namespace,
		},
		Spec: v1beta1.ReplicaSetSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{PserverLabel: job.ObjectMeta.Name, "priority": job.Spec.Annotations.Priority},
					Annotations: map[string]string{"scheduling.k8s.io/group-name": job.Spec.PodGroupName},
				},
				Spec: corev1.PodSpec{
					SchedulerName: job.Spec.SchedulerName,
					Volumes:       job.Spec.Volumes,
					Containers: []corev1.Container{
						{
							Name:            string(PSERVER),
							Image:           job.Spec.Image,
							Ports:           podPorts(job, PSERVER),
							Env:             envs,
							ImagePullPolicy: imagePullPolicy,
							Resources:       job.Spec.Pserver.Resources,
							Lifecycle:       parseLifeCycle(job, PSERVER),
							ReadinessProbe:  job.Spec.Pserver.ReadinessProbe,
							LivenessProbe:   job.Spec.Pserver.LivenessProbe,
						},
					},
					Tolerations:                   parseTolerations(job, PSERVER),
					TerminationGracePeriodSeconds: job.Spec.Pserver.GracePeriodSeconds,
					NodeSelector:                  job.Spec.Pserver.NodeSelector,
				},
			},
		},
	}

	if outter {
		var command []string
		entryPoint := job.Spec.Pserver.Entrypoint
		if len(entryPoint) == 0 {
			// default start command
			if job.Spec.FaultTolerant {
				command = []string{"paddle_k8s", "start_new_pserver"}
			} else {
				command = []string{"paddle_k8s", "start_pserver"}
			}
		} else {
			// user-defined start command
			command = strings.Split(entryPoint, " ")
		}
		spec.Spec.Template.Spec.Containers[0].Command = command
	}

	if job.Spec.Annotations.Scheduler != "" {
		spec.Spec.Template.Spec.SchedulerName = job.Spec.Annotations.Scheduler
	}

	if len(job.Spec.Trainer.ImagePullSecrets) != 0 {
		spec.Spec.Template.Spec.ImagePullSecrets = job.Spec.Trainer.ImagePullSecrets
	}

	return spec
}

// parseToTrainer parse TrainingJob to a kubernetes job resource.
func parseToTrainer(job *paddlev1.TrainingJob, extraEnv []corev1.EnvVar, outter bool) *batchv1.Job {
	replicas := int32(job.Spec.Trainer.MinInstance)
	backoffLimit := int32(0)

	envs := podEnv(job, job.Spec.Trainer.Envs)
	envs = append(envs, extraEnv...)
	spec := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.ObjectMeta.Name + "-trainer",
			Namespace: job.ObjectMeta.Namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Completions:  &replicas,
			Parallelism:  &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{TrainerLabel: job.ObjectMeta.Name,
						"priority": job.Spec.Annotations.Priority},
					Annotations: map[string]string{"scheduling.k8s.io/group-name": job.Spec.PodGroupName, "priority": job.Spec.Annotations.Priority},
				},
				Spec: corev1.PodSpec{
					SchedulerName: job.Spec.SchedulerName,
					Volumes:       job.Spec.Volumes,
					Containers: []corev1.Container{
						{
							Name:            string(TRAINER),
							Image:           job.Spec.Image,
							ImagePullPolicy: imagePullPolicy,
							VolumeMounts:    job.Spec.VolumeMounts,
							Env:             envs,
							Resources:       job.Spec.Trainer.Resources,
							Ports:           podPorts(job, TRAINER),
							Lifecycle:       parseLifeCycle(job, TRAINER),
							ReadinessProbe:  job.Spec.Trainer.ReadinessProbe,
							LivenessProbe:   job.Spec.Trainer.LivenessProbe,
						},
					},
					Tolerations:                   parseTolerations(job, TRAINER),
					TerminationGracePeriodSeconds: job.Spec.Trainer.GracePeriodSeconds,
					NodeSelector:                  job.Spec.Trainer.NodeSelector,
					RestartPolicy:                 "Never",
				},
			},
		},
	}

	if outter {
		var command []string
		entryPoint := job.Spec.Trainer.Entrypoint
		if len(entryPoint) == 0 {
			// default start command
			if job.Spec.FaultTolerant {
				command = []string{"paddle_k8s", "start_new_trainer"}
			} else {
				command = []string{"paddle_k8s", "start_trainer"}
			}
		} else {
			// user-defined start command
			command = strings.Split(entryPoint, " ")
		}
		spec.Spec.Template.Spec.Containers[0].Command = command
	}

	if job.Spec.Annotations.Scheduler != "" {
		spec.Spec.Template.Spec.SchedulerName = job.Spec.Annotations.Scheduler
	}

	if len(job.Spec.Trainer.ImagePullSecrets) != 0 {
		spec.Spec.Template.Spec.ImagePullSecrets = job.Spec.Trainer.ImagePullSecrets
	}

	return spec
}

func masterResource(job *paddlev1.TrainingJob) *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu":    *apiresource.NewQuantity(int64(2), apiresource.DecimalSI),
			"memory": apiresource.MustParse("1Gi"),
		},
		Requests: corev1.ResourceList{
			"cpu":    *apiresource.NewQuantity(int64(1), apiresource.DecimalSI),
			"memory": apiresource.MustParse("500Mi"),
		},
	}
}

func getEtcdPodSpec(job *paddlev1.TrainingJob) *corev1.Container {
	command := []string{"etcd", "-name", "etcd0",
		"-advertise-client-urls", "http://$(POD_IP):2379,http://$(POD_IP):4001",
		"-listen-client-urls", "http://0.0.0.0:2379,http://0.0.0.0:4001",
		"-initial-advertise-peer-urls", "http://$(POD_IP):2380",
		"-listen-peer-urls", "http://0.0.0.0:2380",
		"-initial-cluster", "etcd0=http://$(POD_IP):2380",
		"-initial-cluster-state", "new"}

	return &corev1.Container{
		Name:            "etcd",
		Image:           "m3ngyang/etcd:v3.2.1",
		ImagePullPolicy: imagePullPolicy,
		Env:             podEnv(job, nil),
		Command:         command,
	}
}

// parseToMaster parse TrainingJob to a kubernetes replicaset resource.
func parseToMaster(job *paddlev1.TrainingJob) *v1beta1.ReplicaSet {
	replicas := int32(1)
	command := []string{"paddle_k8s", "start_master"}

	return &v1beta1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "extensions/v1beta1",
			APIVersion: "ReplicaSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.ObjectMeta.Name + "-master",
			Namespace: job.ObjectMeta.Namespace,
		},
		Spec: v1beta1.ReplicaSetSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{MasterLabel: job.ObjectMeta.Name,
						"priority": job.Spec.Annotations.Priority},
					Annotations: map[string]string{"scheduling.k8s.io/group-name": job.Spec.PodGroupName},
				},
				Spec: corev1.PodSpec{
					SchedulerName: job.Spec.SchedulerName,
					Volumes:       job.Spec.Volumes,
					Containers: []corev1.Container{
						{
							Name:            string(MASTER),
							Image:           job.Spec.Image,
							ImagePullPolicy: imagePullPolicy,
							Ports:           masterPorts(job),
							Command:         command,
							VolumeMounts:    job.Spec.VolumeMounts,
							Resources:       *masterResource(job),
							Lifecycle:       parseLifeCycle(job, MASTER),
							ReadinessProbe:  job.Spec.Master.ReadinessProbe,
							LivenessProbe:   job.Spec.Master.LivenessProbe,
						},
						*getEtcdPodSpec(job),
					},
					Tolerations:                   parseTolerations(job, MASTER),
					TerminationGracePeriodSeconds: job.Spec.Master.GracePeriodSeconds,
					NodeSelector:                  job.Spec.Master.NodeSelector,
				},
			},
		},
	}
}

// general functions that pserver, trainer use the same
func podPorts(job *paddlev1.TrainingJob, podType PodType) []corev1.ContainerPort {

	var portsTotal = 0
	var basePort int32 = 0
	ports := make([]corev1.ContainerPort, 0)

	if job.Spec.FrameWork == nil {

		if job.Spec.IsNccl && TRAINER == podType ||
			!job.Spec.IsNccl && PSERVER == podType {

			basePort = int32(job.Spec.Port)
			portsTotal = job.Spec.PortsNum + job.Spec.PortsNumForSparse

			for i := 0; i < portsTotal; i++ {
				ports = append(ports, corev1.ContainerPort{
					Name:          fmt.Sprintf("jobport-%d", basePort),
					ContainerPort: basePort,
				})
				basePort++
			}

			return ports
		}
		return nil
	}

	framework := *job.Spec.FrameWork

	if framework.Type == paddlev1.Multi {

		if PSERVER == podType {

			portsTotal = job.Spec.PortsNum + job.Spec.PortsNumForSparse
			basePort = int32(job.Spec.Port)

		} else if TRAINER == podType {

			if paddlev1.TensorFlow != framework.Name {
				return nil
			}

			portsTotal = job.Spec.TrainerPortsNum
			basePort = int32(job.Spec.TrainerPort)

		}

		for i := 0; i < portsTotal; i++ {
			ports = append(ports, corev1.ContainerPort{
				Name:          fmt.Sprintf("jobport-%d", basePort),
				ContainerPort: basePort,
			})
			basePort++
		}
		return ports
	}

	if framework.Type == paddlev1.Nccl2 {

		if TRAINER != podType {
			return nil
		}

		basePort = int32(job.Spec.TrainerPort)
		portsTotal = job.Spec.TrainerPortsNum

		for i := 0; i < portsTotal; i++ {
			ports = append(ports, corev1.ContainerPort{
				Name:          fmt.Sprintf("jobport-%d", basePort),
				ContainerPort: basePort,
			})
			basePort++
		}

		return ports

	}

	return nil
}

func masterPorts(job *paddlev1.TrainingJob) []corev1.ContainerPort {
	ports := []corev1.ContainerPort{
		{
			Name:          "master-port",
			ContainerPort: 8080,
		},
		{
			Name:          "etcd-port",
			ContainerPort: 2379,
		},
	}
	return ports
}

func podEnv(job *paddlev1.TrainingJob, envs map[string]string) []corev1.EnvVar {
	needGPU := "0"
	if job.NeedGPU() {
		needGPU = "1"
	}
	trainerCount := 1
	if job.NeedGPU() {
		q := job.Spec.Trainer.Resources.Requests.NvidiaGPU()
		trainerCount = int(q.Value())
	} else {
		q := job.Spec.Trainer.Resources.Requests.Cpu()
		// FIXME: CPU resource value can be less than 1.
		trainerCount = int(q.Value())
	}

	podEnv := []corev1.EnvVar{
		{Name: "PADDLE_JOB_NAME", Value: job.ObjectMeta.Name},
		// NOTICE: TRAINERS, PSERVERS, PADDLE_INIT_NUM_GRADIENT_SERVERS
		//         these env are used for non-faulttolerant training,
		//         use min-instance all the time. When job is elastic,
		//         these envs are not used.
		{Name: "TRAINERS", Value: strconv.Itoa(job.Spec.Trainer.MinInstance)},
		{Name: "PSERVERS", Value: strconv.Itoa(job.Spec.Pserver.MinInstance)},
		{Name: "ENTRY", Value: job.Spec.Trainer.Entrypoint},
		// FIXME: TOPOLOGY deprecated
		{Name: "TOPOLOGY", Value: job.Spec.Trainer.Entrypoint},
		{Name: "TRAINER_PACKAGE", Value: job.Spec.Trainer.Workspace},
		{Name: "PADDLE_INIT_PORT", Value: strconv.Itoa(job.Spec.Port)},
		// PADDLE_INIT_TRAINER_COUNT should be same to gpu number when use gpu
		// and cpu cores when using cpu
		{Name: "PADDLE_INIT_TRAINER_COUNT", Value: strconv.Itoa(trainerCount)},
		{Name: "PADDLE_INIT_PORTS_NUM", Value: strconv.Itoa(job.Spec.PortsNum)},
		{Name: "PADDLE_INIT_PORTS_NUM_FOR_SPARSE", Value: strconv.Itoa(job.Spec.PortsNumForSparse)},
		{Name: "PADDLE_INIT_NUM_GRADIENT_SERVERS", Value: strconv.Itoa(job.Spec.Trainer.MinInstance)},
		{Name: "PADDLE_INIT_NUM_PASSES", Value: strconv.Itoa(job.Spec.Passes)},
		{Name: "PADDLE_INIT_USE_GPU", Value: needGPU},
		{Name: "LD_LIBRARY_PATH", Value: "/usr/local/cuda/lib64"},
		{Name: "NAMESPACE", ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		}},
		{Name: "POD_IP", ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		}},
		{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		}},
	}

	for k, v := range envs {
		item := corev1.EnvVar{
			Name:  k,
			Value: v,
		}
		podEnv = append(podEnv, item)
	}
	return podEnv
}

// Validate updates default values for the added job and validates the fields.
func (p *DefaultJobParser) Validate(job *paddlev1.TrainingJob) error {

	var frameWork *paddlev1.Framework = nil

	if job.Spec.FrameWork != nil {
		frameWork = job.Spec.FrameWork
	}
	// Fill in default values
	// FIXME: Need to test. What is the value if specified "omitempty"
	if job.Spec.Port == 0 {
		job.Spec.Port = 7164
	}
	if job.Spec.PortsNum == 0 {
		job.Spec.PortsNum = 1
	}
	if job.Spec.PortsNumForSparse == 0 {
		job.Spec.PortsNumForSparse = 1
	}
	if job.Spec.Image == "" {
		job.Spec.Image = "paddlepaddle/paddlecloud-job"
	}
	if job.Spec.Passes == 0 {
		job.Spec.Passes = 1
	}
	// only one trainer instance for local job
	if frameWork == nil && job.Spec.LocalJob ||
		frameWork != nil && frameWork.Type == paddlev1.Local {
		job.Spec.Trainer.MaxInstance = 1
		job.Spec.Trainer.MinInstance = 1
	}

	//if !job.Spec.FaultTolerant && job.Elastic() {
	//	return errors.New("max-instances should equal to min-instances when fault_tolerant is disabled")
	//}
	// TODO: add validations.

	return nil
}

func (p *DefaultJobParser) GetExtraEnv(job *paddlev1.TrainingJob, kube kubernetes.Interface) ([]corev1.EnvVar, error) {
	var envs []corev1.EnvVar

	if !job.Spec.Matrix {
		kubeSvc, err := kube.CoreV1().Services("default").Get("kubernetes", metav1.GetOptions{})
		if err != nil {
			return envs, err
		}
		item := corev1.EnvVar{
			Name:  "KUBERNETES_SERVICE_HOST",
			Value: kubeSvc.Spec.ClusterIP,
		}
		envs = append(envs, item)
	}

	return envs, nil
}

func parseLifeCycle(job *paddlev1.TrainingJob, podType PodType) *corev1.Lifecycle {

	cmd := []string{}
	switch podType {
	case TRAINER:
		if job.Spec.Trainer.GracePeriodSeconds != nil {
			cmd = job.Spec.Trainer.PreStopCmd
		}
		break
	case MASTER:
		if job.Spec.Master.GracePeriodSeconds != nil {
			cmd = job.Spec.Master.PreStopCmd
		}
		break
	case PSERVER:
		if job.Spec.Pserver.GracePeriodSeconds != nil {
			cmd = job.Spec.Pserver.PreStopCmd
		}
		break
	default:
		return nil
	}
	if len(cmd) == 0 {
		return nil
	}

	return &corev1.Lifecycle{
		PreStop: &corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: cmd,
			},
		},
	}
}

func parseTolerations(job *paddlev1.TrainingJob, podType PodType) []corev1.Toleration {
	switch podType {
	case TRAINER:
		if len(job.Spec.Trainer.Tolerations) > 0 {
			return job.Spec.Trainer.Tolerations
		}
		return nil
	case MASTER:
		if len(job.Spec.Master.Tolerations) > 0 {
			return job.Spec.Trainer.Tolerations
		}
		return nil
	case PSERVER:
		if len(job.Spec.Pserver.Tolerations) > 0 {
			return job.Spec.Trainer.Tolerations
		}
		return nil
	default:
		return nil
	}
}

func (p *DefaultJobParser) ParseToTrainingJob(job *paddlev1.TrainingJob,
	extraEnv []corev1.EnvVar) *paddlev1.TrainingJob {
	var frameWork *paddlev1.Framework = nil

	if job.Spec.FaultTolerant {
		job.Spec.Master.ReplicaSpec = parseToMaster(job)
		if job.Spec.HostNetwork {
			job.Spec.Master.ReplicaSpec.Spec.Template.Spec.HostNetwork = true
		}
	} else {
		job.Spec.Master = paddlev1.MasterSpec{}
	}

	job.Spec.Trainer.ReplicaSpec = parseToTrainer(job, extraEnv, false)

	if job.Spec.FrameWork != nil {
		frameWork = job.Spec.FrameWork

		if frameWork.Type == paddlev1.Multi {
			job.Spec.Pserver.ReplicaSpec = parseToPserver(job, extraEnv, false)

			if job.Spec.HostNetwork {
				job.Spec.Pserver.ReplicaSpec.Spec.Template.Spec.HostNetwork = true

				if frameWork.Name == paddlev1.TensorFlow {
					job.Spec.Trainer.ReplicaSpec.Spec.Template.Spec.HostNetwork = true
				}
			}

		} else {
			job.Spec.Pserver = paddlev1.PserverSpec{}
		}

		if frameWork.Type == paddlev1.Nccl2 && job.Spec.HostNetwork {
			job.Spec.Trainer.ReplicaSpec.Spec.Template.Spec.HostNetwork = true
		}
		return job
	}

	if !job.Spec.LocalJob && !job.Spec.IsNccl {
		job.Spec.Pserver.ReplicaSpec = parseToPserver(job, extraEnv, false)
		if job.Spec.HostNetwork {
			job.Spec.Pserver.ReplicaSpec.Spec.Template.Spec.HostNetwork = true
		}
	} else {
		job.Spec.Pserver = paddlev1.PserverSpec{}
	}

	if job.Spec.IsNccl && job.Spec.HostNetwork {
		job.Spec.Trainer.ReplicaSpec.Spec.Template.Spec.HostNetwork = true
	}

	return job
}

// general functions end
