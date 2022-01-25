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

package driver

import (
	"context"
	"os"
	"testing"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/paddleflow/paddle-operator/api/v1alpha1"
	"github.com/paddleflow/paddle-operator/controllers/extensions/common"
)

func TestBaseDriver_DoClearJob(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Error(err)
	}
	var paths []string
	path1 := home + "/test_DoClearJob"
	err = os.MkdirAll(path1, os.ModePerm)
	if err != nil {
		t.Errorf("make dir %s error: %s", path1, err.Error())
	}
	paths = append(paths, path1)
	path2 := home + "/test_DoClearJob" + "/test"
	err = os.MkdirAll(path2, os.ModePerm)
	if err != nil {
		t.Errorf("make dir %s error: %s", path2, err.Error())
	}
	paths = append(paths, path2)
	logr := zap.New(func(o *zap.Options) {
		o.Development = true
	})
	opt := &v1alpha1.ClearJobOptions{Paths: paths}
	d, _ := GetDriver("juicefs")
	if err := d.DoClearJob(context.Background(), opt, logr); err != nil {
		t.Errorf("DoClearJob error: %s", err.Error())
	}
}

func TestBaseDriver_CreateClearJobOptions(t *testing.T) {
	volumes := []v1.Volume{
		{
			Name: "dev-shm-imagenet-0",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/dev/shm/imagenet-0",
				},
			},
		},
		{
			Name: "dev-shm-imagenet-1",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/dev/shm/imagenet-1",
				},
			},
		},
	}
	volumeMounts := []v1.VolumeMount{
		{
			Name:      "dev-shm-imagenet-0",
			MountPath: "/cache/dev-shm-imagenet-0",
		},
		{
			Name:      "dev-shm-imagenet-1",
			MountPath: "/cache/dev-shm-imagenet-1",
		},
	}
	statefulSet := &appv1.StatefulSet{
		Spec: appv1.StatefulSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: volumes,
					Containers: []v1.Container{
						{
							VolumeMounts: volumeMounts,
						},
					},
				},
			},
		},
	}

	sampleJob := &v1alpha1.SampleJob{
		Spec: v1alpha1.SampleJobSpec{
			JobOptions: v1alpha1.JobOptions{
				ClearOptions: &v1alpha1.ClearJobOptions{
					Paths: []string{
						"/dev/shm/imagenet-0/train",
						"/cache/dev-shm-imagenet-1",
						"/dev/shm/imagenet-0/",
					},
				},
			},
		},
	}
	ctx := &common.RequestContext{StatefulSet: statefulSet, SampleJob: sampleJob}

	d, _ := GetDriver("juicefs")
	opts := &v1alpha1.ClearJobOptions{}
	err := d.CreateClearJobOptions(opts, ctx)
	if err != nil {
		t.Error("create clear job options error: ", err.Error())
		return
	}
	if len(opts.Paths) != 3 {
		t.Error("length of paths should be 3")
		return
	}
	if opts.Paths[0] != "/cache/dev-shm-imagenet-0/train" {
		t.Error("create clear job options not as expected")
	}
	if opts.Paths[1] != "/cache/dev-shm-imagenet-1/*" {
		t.Error("create clear job options not as expected")
	}
	if opts.Paths[2] != "/cache/dev-shm-imagenet-0/*" {
		t.Error("create clear job options not as expected")
	}
}

func TestBaseDriver_CreateRmrJobOptions(t *testing.T) {
	sampleJob := &v1alpha1.SampleJob{
		Spec: v1alpha1.SampleJobSpec{
			JobOptions: v1alpha1.JobOptions{
				RmrOptions: &v1alpha1.RmrJobOptions{
					Paths: []string{
						"/train",
						"val/n123323",
					},
				},
			},
		},
	}
	request := &ctrl.Request{NamespacedName: types.NamespacedName{Name: "imagenet"}}
	ctx := &common.RequestContext{SampleJob: sampleJob, Req: request}
	d, _ := GetDriver("juicefs")
	opts := &v1alpha1.RmrJobOptions{}
	err := d.CreateRmrJobOptions(opts, ctx)
	if err != nil {
		t.Error("create rmr job option error: ", err.Error())
		return
	}
	if len(opts.Paths) != 2 {
		t.Error("length of paths should be 2")
		return
	}
	if opts.Paths[0] != "/mnt/imagenet/train" {
		t.Error("create rmr job options not as expected")
	}
	if opts.Paths[1] != "/mnt/imagenet/val/n123323" {
		t.Error("create rmr job options not as expected")
	}
}
