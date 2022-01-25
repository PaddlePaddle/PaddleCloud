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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-logr/logr"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/paddleflow/paddle-operator/api/v1alpha1"
	"github.com/paddleflow/paddle-operator/controllers/extensions/common"
	"github.com/paddleflow/paddle-operator/controllers/extensions/utils"
)

const (
	DefaultDriver = JuiceFSDriver
)

var (
	StorageClassName = "paddle-operator"
	driverMap        map[v1alpha1.DriverName]Driver
)

func init() {
	juiceFS := NewJuiceFSDriver()
	driverMap = map[v1alpha1.DriverName]Driver{
		JuiceFSDriver: juiceFS,
	}
}

type Driver interface {
	// CreatePV create persistent volume by specified driver
	CreatePV(pv *v1.PersistentVolume, ctx *common.RequestContext) error

	// CreatePVC create persistent volume claim for PaddleJob
	CreatePVC(pvc *v1.PersistentVolumeClaim, ctx *common.RequestContext) error

	// GetLabel get the label to mark pv、pvc and nodes which have cached data
	GetLabel(sampleSetName string) string

	// CreateService create a service for runtime StatefulSet
	CreateService(service *v1.Service, ctx *common.RequestContext) error

	// GetServiceName get the name of runtime StatefulSet service
	GetServiceName(sampleSetName string) string

	// CreateRuntime create runtime StatefulSet to manager cache data
	CreateRuntime(ds *appv1.StatefulSet, ctx *common.RequestContext) error

	// GetRuntimeName get the runtime StatefulSet name
	GetRuntimeName(sampleSetName string) string

	// CreateSyncJobOptions create the options of sync job, the controller will post it to runtime server
	CreateSyncJobOptions(opt *v1alpha1.SyncJobOptions, ctx *common.RequestContext) error

	// CreateWarmupJobOptions create the options of warmup job, this method now only use by SampleJob Controller
	CreateWarmupJobOptions(opt *v1alpha1.WarmupJobOptions, ctx *common.RequestContext) error

	// CreateRmrJobOptions create the options of rmr job, used by SampleJob controller
	CreateRmrJobOptions(opt *v1alpha1.RmrJobOptions, ctx *common.RequestContext) error

	// CreateClearJobOptions create the options of clear job, used by SampleJob controller
	CreateClearJobOptions(opt *v1alpha1.ClearJobOptions, ctx *common.RequestContext) error

	// CreateCacheStatus get the data status in mount and cache paths
	CreateCacheStatus(opt *common.ServerOptions, status *v1alpha1.CacheStatus) error

	// DoSyncJob call by runtime server, sync data from remote storage to cache engine
	DoSyncJob(ctx context.Context, opt *v1alpha1.SyncJobOptions, log logr.Logger) error

	// DoClearJob call by runtime server, clear the cached data
	DoClearJob(ctx context.Context, opt *v1alpha1.ClearJobOptions, log logr.Logger) error

	// DoWarmupJob call by runtime server, warmup data to local storage on each node respectively
	DoWarmupJob(ctx context.Context, opt *v1alpha1.WarmupJobOptions, log logr.Logger) error

	// DoRmrJob call by runtime server, remove the data of specified path from cache engine
	DoRmrJob(ctx context.Context, opt *v1alpha1.RmrJobOptions, log logr.Logger) error
}

// GetDriver get csi driver by name, return error if not found
func GetDriver(name v1alpha1.DriverName) (Driver, error) {
	if string(name) == "" {
		name = DefaultDriver
	}
	if d, exists := driverMap[name]; exists {
		return d, nil
	}
	return nil, fmt.Errorf("driver %s not found", name)
}

type BaseDriver struct {
	Name v1alpha1.DriverName
}

// CreatePVC create persistent volume claim, and it will be used by runtime server and PaddleJob worker pods
func (d *BaseDriver) CreatePVC(pvc *v1.PersistentVolumeClaim, ctx *common.RequestContext) error {
	label := d.GetLabel(ctx.Req.Name)
	objectMeta := metav1.ObjectMeta{
		Name:      ctx.Req.Name,
		Namespace: ctx.Req.Namespace,
		Labels: map[string]string{
			label: "true",
		},
		Annotations: map[string]string{
			"CreatedBy": common.PaddleOperatorLabel,
		},
	}
	pvc.ObjectMeta = objectMeta

	spec := v1.PersistentVolumeClaimSpec{
		AccessModes: []v1.PersistentVolumeAccessMode{
			v1.ReadWriteMany,
		},
		Resources: v1.ResourceRequirements{
			Requests: map[v1.ResourceName]resource.Quantity{
				v1.ResourceStorage: resource.MustParse(common.ResourceStorage),
			},
		},
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				label: "true",
			},
		},
		StorageClassName: &StorageClassName,
	}
	pvc.Spec = spec

	return nil
}

// CreateService create service for runtime StatefulSet server
func (d *BaseDriver) CreateService(service *v1.Service, ctx *common.RequestContext) error {
	label := d.GetLabel(ctx.Req.Name)
	serviceName := d.GetServiceName(ctx.Req.Name)
	objectMeta := metav1.ObjectMeta{
		Name:      serviceName,
		Namespace: ctx.Req.Namespace,
		Labels: map[string]string{
			label: "true",
		},
		Annotations: map[string]string{
			"CreatedBy": common.PaddleOperatorLabel,
		},
	}
	service.ObjectMeta = objectMeta

	runtimeName := d.GetRuntimeName(ctx.Req.Name)
	selector := map[string]string{
		label:  "true",
		"name": runtimeName,
	}

	port := v1.ServicePort{
		Name: common.RuntimeServiceName,
		Port: common.RuntimeServicePort,
	}

	spec := v1.ServiceSpec{
		Selector:  selector,
		Ports:     []v1.ServicePort{port},
		ClusterIP: "None",
	}
	service.Spec = spec

	return nil
}

// DoClearJob clear the cache data in folders specified by options
func (d *BaseDriver) DoClearJob(ctx context.Context, opt *v1alpha1.ClearJobOptions, log logr.Logger) error {
	if len(opt.Paths) == 0 {
		return errors.New("clear job option paths not set")
	}
	for _, path := range opt.Paths {
		if _, err := os.Stat(strings.TrimSuffix(path, "*")); err != nil {
			return fmt.Errorf("path %s is not valid, error: %s", path, err.Error())
		}
	}
	rmCmd := "rm -rf " + strings.Join(opt.Paths, " ")
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", rmCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clear job cmd: %s; error: %s", cmd.String(), err.Error())
	}
	log.V(1).Info(cmd.String())
	return nil
}

func (d *BaseDriver) CreateCacheStatus(opt *common.ServerOptions, status *v1alpha1.CacheStatus) error {
	if opt.CacheDirs == nil || len(opt.CacheDirs) == 0 {
		return fmt.Errorf("option cacheDirs not set")
	}
	if !strings.HasPrefix(opt.CacheDirs[0], "/") {
		return fmt.Errorf("option cacheDirs: %s is not valid", opt.CacheDirs[0])
	}
	if !strings.HasPrefix(opt.DataDir, "/") {
		return fmt.Errorf("option dataDir: %s is not valid", opt.DataDir)
	}

	var errs []string
	timeout := time.Duration(opt.Timeout)

	// get total data size from data mount path
	totalSize, err := utils.DiskUsageOfPaths(timeout, opt.DataDir)
	if err == nil {
		status.TotalSize = totalSize
	} else {
		errs = append(errs, fmt.Sprintf("get totalSize error: %s", err.Error()))
	}

	var fileNumberOfPaths func(timeout time.Duration, paths ...string) (string, error)
	if d.Name == JuiceFSDriver {
		fileNumberOfPaths = utils.JuiceFileNumberOfPath
	} else {
		fileNumberOfPaths = utils.FileNumberOfPaths
	}
	// get total data file number from data mount path
	totalFiles, err := fileNumberOfPaths(timeout, opt.DataDir)
	if err == nil {
		status.TotalFiles = totalFiles
	} else {
		errs = append(errs, fmt.Sprintf("get totalFiles error: %s", err.Error()))
	}

	// get cache data size from cache data mount paths
	cacheSize, err := utils.DiskUsageOfPaths(timeout, opt.CacheDirs...)
	if err == nil {
		status.CachedSize = cacheSize
	} else {
		errs = append(errs, fmt.Sprintf("get cachedSize error: %s", err.Error()))
	}

	// get disk space status of cache paths
	diskStatus, err := utils.DiskSpaceOfPaths(timeout, opt.CacheDirs...)
	if err == nil {
		status.DiskSize = diskStatus["diskSize"]
		status.DiskUsed = diskStatus["diskUsed"]
		status.DiskAvail = diskStatus["diskAvail"]
	} else {
		errs = append(errs, fmt.Sprintf("get disk status error: %s", err.Error()))
	}

	if len(errs) != 0 {
		return errors.New(strings.Join(errs, ";"))
	}
	return nil
}

func (d *BaseDriver) CreateRmrJobOptions(opt *v1alpha1.RmrJobOptions, ctx *common.RequestContext) error {
	if ctx.SampleJob == nil || ctx.SampleJob.Spec.RmrOptions == nil {
		return fmt.Errorf("the options of rmr job cannot be empty")
	}
	ctx.SampleJob.Spec.RmrOptions.DeepCopyInto(opt)
	if len(opt.Paths) == 0 {
		return fmt.Errorf("option of paths cannot be empty, and it is relative path from source directory")
	}
	mountPath := d.getRuntimeDataMountPath(ctx.Req.Name)
	var validPaths []string
	for _, path := range opt.Paths {
		validPath := mountPath + "/" + strings.TrimPrefix(path, "/")
		validPaths = append(validPaths, validPath)
	}
	opt.Paths = validPaths
	return nil
}

func (d *BaseDriver) CreateClearJobOptions(opt *v1alpha1.ClearJobOptions, ctx *common.RequestContext) error {
	// if SampleJob is not nil, use clear job options from SampleJob
	if ctx.SampleJob != nil && ctx.SampleJob.Spec.ClearOptions != nil {
		ctx.SampleJob.Spec.ClearOptions.DeepCopyInto(opt)
	}
	// check the paths given by user, whether it has prefixed with host path or mount path
	volumes := ctx.StatefulSet.Spec.Template.Spec.Volumes
	if len(ctx.StatefulSet.Spec.Template.Spec.Containers) != 1 {
		return fmt.Errorf("length of statefulset %s containers is not equal 1", ctx.StatefulSet.Name)
	}
	volumeMounts := ctx.StatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts
	hostMountPathMap := getHostMountPathMap(volumes, volumeMounts)
	if len(opt.Paths) != 0 {
		var validPaths []string
		for _, path := range opt.Paths {
			validPath, valid := getValidClearPath(path, hostMountPathMap)
			if !valid {
				return fmt.Errorf("given path %s is not valid, it should in host path or mount path", path)
			}
			validPaths = append(validPaths, validPath)
		}
		opt.Paths = validPaths
	} else {
		for _, mountPath := range hostMountPathMap {
			opt.Paths = append(opt.Paths, strings.TrimSuffix(mountPath, "/")+"/*")
		}
	}
	return nil
}

// GetLabel label is concatenated by PaddleLabel、driver name and SampleSet name
func (d *BaseDriver) GetLabel(sampleSetName string) string {
	return common.PaddleLabel + "/" + string(d.Name) + "-" + sampleSetName
}

func (d *BaseDriver) GetRuntimeName(sampleSetName string) string {
	return sampleSetName + "-" + common.RuntimeContainerName
}

func (d *BaseDriver) GetServiceName(sampleSetName string) string {
	return d.GetRuntimeName(sampleSetName) + "-" + common.RuntimeServiceName
}

func (d *BaseDriver) getRuntimeCacheMountPath(name string) string {
	mountPath := os.Getenv("CACHE_MOUNT_PATH")
	if mountPath == "" {
		return common.RuntimeCacheMountPath + "/" + name
	}
	return mountPath + "/" + name
}

func (d *BaseDriver) getRuntimeDataMountPath(name string) string {
	mountPath := os.Getenv("DATA_MOUNT_PATH")
	if mountPath == "" {
		return common.RuntimeDateMountPath + "/" + name
	}
	return mountPath + "/" + name
}

func getHostMountPathMap(volumes []v1.Volume, volumeMounts []v1.VolumeMount) map[string]string {
	hostMountPathMap := make(map[string]string)
	for _, volume := range volumes {
		for _, mount := range volumeMounts {
			if volume.HostPath == nil {
				continue
			}
			if volume.Name != mount.Name {
				continue
			}
			hostMountPathMap[volume.HostPath.Path] = mount.MountPath
		}
	}
	return hostMountPathMap
}

func getValidClearPath(path string, hostMountMap map[string]string) (string, bool) {
	for hostPath, mountPath := range hostMountMap {
		if strings.HasPrefix(path, hostPath) {
			validPath := strings.Replace(path, hostPath, mountPath, 1)
			if strings.TrimSuffix(validPath, "/") == mountPath {
				return strings.TrimSuffix(validPath, "/") + "/*", true
			}
			return validPath, true
		}
		if strings.HasPrefix(path, mountPath) {
			if strings.TrimSuffix(path, "/") == mountPath {
				return strings.TrimSuffix(path, "/") + "/*", true
			}
			return path, true
		}
	}
	return "", false
}
