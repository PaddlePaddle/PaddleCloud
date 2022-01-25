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
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/paddleflow/paddle-operator/api/v1alpha1"
	"github.com/paddleflow/paddle-operator/controllers/extensions/common"
	"github.com/paddleflow/paddle-operator/controllers/extensions/utils"
)

func TestJuiceFS_getMountOptions(t *testing.T) {
	mountOptions := v1alpha1.MountOptions{
		JuiceFSMountOptions: &v1alpha1.JuiceFSMountOptions{
			OpenCache:      7200,
			CacheSize:      307200,
			AttrCache:      7200,
			EntryCache:     7200,
			DirEntryCache:  7200,
			Prefetch:       1,
			BufferSize:     1024,
			EnableXattr:    true,
			WriteBack:      true,
			FreeSpaceRatio: "0.2",
			CacheDir:       "/dev/shm/imagenet",
		},
	}

	sampleSet := &v1alpha1.SampleSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "imagenet",
			Namespace: "paddle-system",
		},
		Spec: v1alpha1.SampleSetSpec{
			Partitions: 2,
			Source: &v1alpha1.Source{
				URI: "https://imagenet.bj.bcebos.com/juicefs",
			},
			SecretRef: &corev1.SecretReference{
				Name: "imagenet",
			},
			CSI: &v1alpha1.CSI{
				Driver:       JuiceFSDriver,
				MountOptions: mountOptions,
			},
			Cache: v1alpha1.Cache{
				Levels: []v1alpha1.CacheLevel{
					{
						MediumType: common.MediumTypeMEM,
						Path:       "/dev/shm/imagenet-0:/dev/shm/imagenet-1",
						CacheSize:  150,
					},
					{
						MediumType: common.MediumTypeSSD,
						Path:       "/data/imagenet",
						CacheSize:  150,
					},
				},
			},
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      "sampleset",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
						},
					},
					},
				},
			},
		},
	}

	options, err := getMountOptions(sampleSet)
	if err != nil {
		t.Errorf("test getMountOptions error: %s", err.Error())
	}
	for _, option := range strings.Split(options, ",") {
		if strings.HasPrefix(option, "cache-dir") {
			cacheDir := strings.Split(option, "=")[1]
			if len(strings.Split(cacheDir, ":")) != 3 {
				t.Errorf("length of cache-dir should be 2: %s", cacheDir)
			}
		}
	}
}

func TestJuiceFS_getVolumeInfo(t *testing.T) {
	mountOptions := "dir-entry-cache=7200,buffer-size=1024,prefetch=1,cache-dir=/dev/shm/imagenet:/data/imagenet:/dev/ssd/imagenet,cache-size=1048576"
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "imagenet",
		},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					VolumeAttributes: map[string]string{
						"mountOptions": mountOptions,
					},
				},
			},
		},
	}

	driver := NewJuiceFSDriver()
	volumes, volumeMounts, serverOpt, err := driver.getVolumeInfo(pv)
	if err != nil {
		t.Error(err)
	}
	if len(volumes) != 4 || len(volumeMounts) != 4 || serverOpt.DataDir == "" || len(serverOpt.CacheDirs) != 3 {
		t.Errorf("len of volumes or volumeMounts not right \n")
	}

}

func TestJuiceFS_DoWarmupJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mountPath, err := os.UserHomeDir()
	if err != nil {
		t.Error(err, "get user home dir error")
	}
	path, err := os.Getwd()
	if err != nil {
		t.Error(err, "get current directory error")
	}
	total := strings.TrimSuffix(path, "/controllers/extensions/driver")
	opts := &v1alpha1.WarmupJobOptions{
		Paths:      []string{total},
		Partitions: 4,
		Strategy: v1alpha1.SampleStrategy{
			Name: common.StrategyRandom,
		},
	}
	logr := zap.New(func(o *zap.Options) {
		o.Development = true
	})
	doWarmupJob := func(index int, opt *v1alpha1.WarmupJobOptions) error {
		logr.V(1).Info("get index", "index", index)

		if index == 0 {
			defer postWarmupMaster(ctx, mountPath, int(opt.Partitions), logr)
			if err = preWarmupMaster(ctx, mountPath, opt); err != nil {
				return fmt.Errorf("master node do pre-warmup job error: %s", err.Error())
			}
			logr.V(1).Info("preWarmupMaster finish....")
		} else {
			defer postWarmupWorker(ctx, mountPath, index, logr)
			if err = preWarmupWorker(ctx, mountPath, index); err != nil {
				return fmt.Errorf("worker %d do pre-warmup job error: %s", index, err.Error())
			}
			logr.V(1).Info("preWarmupWoker finish....")
		}
		opt.File = mountPath + common.WarmupDirPath + common.WarmupFilePrefix + "." + strconv.Itoa(index)
		warmupArgs := utils.NoZeroOptionToArgs(&opt.JuiceFSWarmupOptions)
		args := []string{"warmup"}
		args = append(args, warmupArgs...)
		args = append(args, opt.Paths...)
		cmd := exec.CommandContext(ctx, "juicefs", args...)
		logr.V(1).Info(cmd.String())
		return nil
	}

	var wg sync.WaitGroup
	for i := 0; i < int(opts.Partitions); i++ {
		wg.Add(1)
		go func(index int) {
			opt := opts.DeepCopy()
			if err := doWarmupJob(index, opt); err != nil {
				t.Error(err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
