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

package k8s

import (
	"errors"
	"fmt"
	"strconv"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	v1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// JobParser is a interface can parse "TrainingJob" to
// ReplicaSet and job.
type JobParser interface {
	Validate(job *paddlejob.TrainingJob) error
	ParseToTrainer(job *paddlejob.TrainingJob) *batchv1.Job
	ParseToPserver(job *paddlejob.TrainingJob) *v1beta1.ReplicaSet
	ParseToMaster(job *paddlejob.TrainingJob) *v1beta1.ReplicaSet
}

// DefaultJobParser implement a basic JobParser.
type DefaultJobParser int

// Validate updates default values for the added job and validates the fields.
func (p *DefaultJobParser) Validate(job *paddlejob.TrainingJob) error {
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

	if !job.Spec.FaultTolerant && job.Elastic() {
		return errors.New("max-instances should equal to min-instances when fault_tolerant is disabled")
	}
	// TODO: add validations.
	return nil
}

// ParseToPserver generate a pserver replicaset resource according to "TrainingJob" resource specs.
func (p *DefaultJobParser) ParseToPserver(job *paddlejob.TrainingJob) *v1beta1.ReplicaSet {
	replicas := int32(job.Spec.Pserver.MinInstance)
	command := make([]string, 2, 2)
	// FIXME: refine these part.
	if job.Spec.FaultTolerant {
		command = []string{"paddle_k8s", "start_pserver"}
	} else {
		command = []string{"paddle_k8s", "start_new_pserver"}
	}

	return &v1beta1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "extensions/v1beta1",
			APIVersion: "ReplicaSet",
		},
		ObjectMeta: job.ObjectMeta,
		Spec: v1beta1.ReplicaSetSpec{
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"paddle-job-pserver": job.ObjectMeta.Name},
				},
				Spec: v1.PodSpec{
					// TODO: setup pserver volumes on cloud.
					Volumes: podVolumes(job),
					Containers: []v1.Container{
						v1.Container{
							Name:      job.ObjectMeta.Name,
							Image:     job.Spec.Image,
							Ports:     podPorts(job),
							Env:       podEnv(job),
							Command:   command,
							Resources: job.Spec.Pserver.Resources,
						},
					},
				},
			},
		},
	}
}

// ParseToTrainer parse TrainingJob to a kubernetes job resource.
func (p *DefaultJobParser) ParseToTrainer(job *paddlejob.TrainingJob) *batchv1.Job {
	// TODO: create job.
	return &batchv1.Job{}
}

// ParseToMaster parse TrainingJob to a kubernetes replicaset resource.
func (p *DefaultJobParser) ParseToMaster(job *paddlejob.TrainingJob) *v1beta1.ReplicaSet {
	// TODO: create master if needed.
	return &v1beta1.ReplicaSet{}
}

// -----------------------------------------------------------------------
// general functions that pserver, trainer use the same
// -----------------------------------------------------------------------
func podPorts(job *paddlejob.TrainingJob) []v1.ContainerPort {
	portsTotal := job.Spec.PortsNum + job.Spec.PortsNumForSparse
	ports := make([]v1.ContainerPort, 8)
	basePort := int32(job.Spec.Port)
	for i := 0; i < portsTotal; i++ {
		ports = append(ports, v1.ContainerPort{
			Name:          fmt.Sprintf("jobport-%d", basePort),
			ContainerPort: basePort,
		})
		basePort++
	}
	return []v1.ContainerPort{}
}

func podEnv(job *paddlejob.TrainingJob) []v1.EnvVar {
	needGPU := "0"
	if job.NeedGPU() {
		needGPU = "1"
	}
	trainerCount := 1
	if job.NeedGPU() {
		q := job.Spec.Trainer.Resources.Requests[paddlejob.GPUResourceName]
		trainerCount = int(q.Value())
	} else {
		q := job.Spec.Trainer.Resources.Requests.Cpu()
		// FIXME: CPU resource value can be less than 1.
		trainerCount = int(q.Value())
	}
	return []v1.EnvVar{
		v1.EnvVar{Name: "PADDLE_JOB_NAME", Value: job.ObjectMeta.Name},
		// NOTICE: TRAINERS, PSERVERS, PADDLE_INIT_NUM_GRADIENT_SERVERS
		//         these env are used for non-faulttolerant training,
		//         use min-instance all the time. When job is elastic,
		//         these envs are not used.
		v1.EnvVar{Name: "TRAINERS", Value: strconv.Itoa(job.Spec.Trainer.MinInstance)},
		v1.EnvVar{Name: "PSERVERS", Value: strconv.Itoa(job.Spec.Pserver.MinInstance)},
		v1.EnvVar{Name: "ENTRY", Value: job.Spec.Trainer.Entrypoint},
		// FIXME: TOPOLOGY deprecated
		v1.EnvVar{Name: "TOPOLOGY", Value: job.Spec.Trainer.Entrypoint},
		v1.EnvVar{Name: "TRAINER_PACKAGE", Value: job.Spec.Trainer.Workspace},
		v1.EnvVar{Name: "PADDLE_INIT_PORT", Value: strconv.Itoa(job.Spec.Port)},
		// PADDLE_INIT_TRAINER_COUNT should be same to gpu number when use gpu
		// and cpu cores when using cpu
		v1.EnvVar{Name: "PADDLE_INIT_TRAINER_COUNT", Value: strconv.Itoa(trainerCount)},
		v1.EnvVar{Name: "PADDLE_INIT_PORTS_NUM", Value: strconv.Itoa(job.Spec.PortsNum)},
		v1.EnvVar{Name: "PADDLE_INIT_PORTS_NUM_FOR_SPARSE", Value: strconv.Itoa(job.Spec.PortsNumForSparse)},
		v1.EnvVar{Name: "PADDLE_INIT_NUM_GRADIENT_SERVERS", Value: strconv.Itoa(job.Spec.Trainer.MinInstance)},
		v1.EnvVar{Name: "PADDLE_INIT_NUM_PASSES", Value: strconv.Itoa(job.Spec.Passes)},
		v1.EnvVar{Name: "PADDLE_INIT_USE_GPU", Value: needGPU},
		v1.EnvVar{Name: "LD_LIBRARY_PATH", Value: job.Spec.Trainer.Entrypoint},
		v1.EnvVar{Name: "NAMESPACE", ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		}},
	}
}

func podVolumes(job *paddlejob.TrainingJob) []v1.Volume {
	// TODO: prepare volumes.
	return []v1.Volume{}
}

func podVolumeMounts(job *paddlejob.TrainingJob) []v1.VolumeMount {
	// TODO: preapare volume mounts for pods.
	return []v1.VolumeMount{}
}

// -----------------------------------------------------------------------
// general functions end
// -----------------------------------------------------------------------
