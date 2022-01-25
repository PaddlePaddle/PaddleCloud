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
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"strconv"
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
	JuiceFSDriver          v1alpha1.DriverName = "juicefs"
	JuiceFSCacheDirOption                      = "cache-dir"
	JuiceFSCacheSizeOption                     = "cache-size"
	JuiceFSCSIDriverName                       = "csi.juicefs.com"
)

const (
	JuiceFSSecretName    string = "name"
	JuiceFSSecretStorage string = "storage"
	JuiceFSSecretMetaURL string = "metaurl"
	JuiceFSSecretBucket  string = "bucket"
	JuiceFSSecretSK      string = "secret-key"
	JuiceFSSecretAK      string = "access-key"
)

var (
	JuiceFSSecretDataKeys      []string
	JuiceFSSupportStorage      []string
	JuiceFSDefaultMountOptions *v1alpha1.JuiceFSMountOptions
)

func init() {
	// data keys of secret must contains when use JuiceFS as csi driver
	JuiceFSSecretDataKeys = []string{
		JuiceFSSecretName, JuiceFSSecretStorage,
		JuiceFSSecretMetaURL, JuiceFSSecretBucket,
	}

	// default JuiceFS mount volume options pass to pv
	JuiceFSDefaultMountOptions = &v1alpha1.JuiceFSMountOptions{
		CacheSize: 1024 * 1024 * 1024, CacheDir: "/dev/shm/",
	}

	// all JuiceFS supported storage backend can refer to
	// https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/how_to_setup_object_storage.md
	JuiceFSSupportStorage = []string{
		common.StorageBOS, common.StorageS3, common.StorageHDFS,
		common.StorageGCS, common.StorageWASB, common.StorageOSS,
		common.StorageCOS, common.StorageKS3, common.StorageUFILE,
		common.StorageQingStor, common.StorageJSS, common.StorageQiNiu,
		common.StorageB2, common.StorageSpace, common.StorageOBS,
		common.StorageOOS, common.StorageSCW, common.StorageMinio,
		common.StorageSCS, common.StorageIBMCOS, common.StorageWASABI,
		common.StorageMSS, common.StorageNOS, common.StorageEOS,
		common.StorageSpeedy, common.StorageCeph, common.StorageSwift,
		common.StorageWebDAV, common.StorageRedis, common.StorageTiKV,
		common.StorageFile,
	}

}

type JuiceFS struct {
	BaseDriver
}

func NewJuiceFSDriver() *JuiceFS {
	return &JuiceFS{
		BaseDriver{
			Name: JuiceFSDriver,
		},
	}
}

// CreatePV create JuiceFS persistent volume with mount options.
// How to set parameters of pv can refer to https://github.com/juicedata/juicefs-csi-driver/tree/master/examples/static-provisioning-mount-options
func (j *JuiceFS) CreatePV(pv *v1.PersistentVolume, ctx *common.RequestContext) error {
	if !isSecretValid(ctx.Secret) {
		return fmt.Errorf("secret %s is not valid", ctx.Secret.Name)
	}

	label := j.GetLabel(ctx.Req.Name)
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
	pv.ObjectMeta = objectMeta

	volumeMode := v1.PersistentVolumeFilesystem
	mountOptions, err := getMountOptions(ctx.SampleSet)
	if err != nil {
		return err
	}

	secretReference := &v1.SecretReference{
		Name:      ctx.Secret.Name,
		Namespace: ctx.Secret.Namespace,
	}
	namespacedName := ctx.Req.NamespacedName.String()
	volumeHandle := strings.ReplaceAll(namespacedName, "/", "-")
	spec := v1.PersistentVolumeSpec{
		AccessModes: []v1.PersistentVolumeAccessMode{
			v1.ReadWriteMany,
		},
		Capacity: v1.ResourceList{
			v1.ResourceStorage: resource.MustParse(common.ResourceStorage),
		},
		StorageClassName: StorageClassName,
		PersistentVolumeSource: v1.PersistentVolumeSource{
			CSI: &v1.CSIPersistentVolumeSource{
				Driver:       JuiceFSCSIDriverName,
				FSType:       string(JuiceFSDriver),
				VolumeHandle: volumeHandle,
				VolumeAttributes: map[string]string{
					"mountOptions": mountOptions,
				},
				NodePublishSecretRef: secretReference,
			},
		},
		PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRetain,
		VolumeMode:                    &volumeMode,
	}
	pv.Spec = spec

	return nil
}

// isSecretValid check if the secret created by user is valid for JuiceFS csi driver
func isSecretValid(secret *v1.Secret) bool {
	for _, key := range JuiceFSSecretDataKeys {
		if _, exists := secret.Data[key]; !exists {
			return false
		}
	}
	return true
}

// GetMountOptions get the JuiceFS mount command options set by user
func getMountOptions(sampleSet *v1alpha1.SampleSet) (string, error) {
	optionMap := make(map[string]reflect.Value)
	// get default mount options and values
	utils.NoZeroOptionToMap(optionMap, JuiceFSDefaultMountOptions)

	// if mount options is not set by user, then cover default value with users
	var userOptions *v1alpha1.JuiceFSMountOptions
	if sampleSet.Spec.CSI != nil {
		csi := sampleSet.Spec.CSI
		if csi.JuiceFSMountOptions != nil {
			userOptions = csi.JuiceFSMountOptions.DeepCopy()
		}
	}
	if userOptions != nil {
		utils.NoZeroOptionToMap(optionMap, userOptions)
	}

	// check if free-space-ratio is valid
	if ratioStr, exist := optionMap["free-space-ratio"]; exist {
		ratio, err := strconv.ParseFloat(ratioStr.String(), 64)
		if err != nil {
			return "", fmt.Errorf("parse free-space-ratio:%s error", ratioStr)
		}
		if ratio >= 1.0 || ratio < 0.0 {
			return "", fmt.Errorf("free-space-ratio:%s is not valid", ratioStr)
		}
	}

	// Get the cache dir set by user, use default path if not set
	cacheSize := 0
	var cacheDir []string
	if len(sampleSet.Spec.Cache.Levels) > 0 {
		levels := sampleSet.Spec.Cache.Levels
		for _, cacheLevel := range levels {
			if cacheLevel.Path == "" {
				continue
			}
			cacheDir = append(cacheDir, cacheLevel.Path)
			cacheSize += cacheLevel.CacheSize
		}
	}

	var optionSlice []string
	for option, value := range optionMap {
		// Override cache-dir with the value from Cache Levels
		if option == JuiceFSCacheDirOption {
			var cacheDirOption string
			if len(cacheDir) > 0 {
				cacheDirs := strings.Join(cacheDir, ":")
				cacheDirOption = JuiceFSCacheDirOption + "=" + cacheDirs
			} else {
				sampleSetName := strings.ToLower(sampleSet.Name)
				cacheDirOption = option + "=" + value.String() + sampleSetName
			}
			optionSlice = append(optionSlice, cacheDirOption)
			continue
		}
		// Override cache-size with the value from Cache Levels
		if cacheSize > 0 && option == JuiceFSCacheSizeOption {
			cacheSizeOption := JuiceFSCacheSizeOption + "=" + strconv.Itoa(cacheSize)
			optionSlice = append(optionSlice, cacheSizeOption)
			continue
		}

		if value.Kind() != reflect.Bool {
			option = fmt.Sprintf("%s=%v", option, value)
		}
		optionSlice = append(optionSlice, option)
	}

	return strings.Join(optionSlice, ","), nil
}

func (j *JuiceFS) CreateRuntime(ds *appv1.StatefulSet, ctx *common.RequestContext) error {
	label := j.GetLabel(ctx.Req.Name)
	runtimeName := j.GetRuntimeName(ctx.Req.Name)
	image, err := utils.GetRuntimeImage()
	if err != nil {
		return err
	}

	objectMeta := metav1.ObjectMeta{
		Name:      runtimeName,
		Namespace: ctx.Req.Namespace,
		Labels: map[string]string{
			label: "true",
		},
		Annotations: map[string]string{
			"CreatedBy": common.PaddleOperatorLabel,
		},
	}
	ds.ObjectMeta = objectMeta

	volumes, volumeMounts, serverOpt, err := j.getVolumeInfo(ctx.PV)
	if err != nil {
		return fmt.Errorf("getVolumeInfo error: %s", err.Error())
	}

	rootOpt := &common.RootCmdOptions{
		Driver:      string(j.Name),
		Development: true,
	}
	command := []string{common.CmdRoot, common.CmdServer}
	rootArgs := utils.NoZeroOptionToArgs(rootOpt)
	serverArgs := utils.NoZeroOptionToArgs(serverOpt)
	command = append(command, rootArgs...)
	command = append(command, serverArgs...)

	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			label:  "true",
			"name": runtimeName,
		},
	}
	podAffinityTerm := v1.PodAffinityTerm{
		LabelSelector: &labelSelector,
		Namespaces:    []string{ctx.Req.Namespace},
		TopologyKey:   "kubernetes.io/hostname",
	}
	podAntiAffinity := v1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
			podAffinityTerm,
		},
	}

	// get cache data mount paths and add it to pre-stop command
	var clearDataPaths []string
	for _, path := range serverOpt.CacheDirs {
		validPath := strings.TrimSuffix(path, "/") + "/*"
		clearDataPaths = append(clearDataPaths, validPath)
	}
	clearDataPath := strings.Join(clearDataPaths, " ")
	clearArgs := "rm -rf " + clearDataPath
	clearCacheCmd := &v1.ExecAction{
		Command: []string{"/bin/sh", "-c", clearArgs},
	}
	preStopHandler := &v1.Handler{
		Exec: clearCacheCmd,
	}
	lifecycle := &v1.Lifecycle{
		PreStop: preStopHandler,
	}

	isPrivileged := true
	container := v1.Container{
		Name:  common.RuntimeContainerName,
		Image: image,
		Ports: []v1.ContainerPort{
			{
				Name:          ctx.Req.Name,
				ContainerPort: common.RuntimeServicePort,
			},
		},
		SecurityContext: &v1.SecurityContext{
			Privileged: &isPrivileged,
		},
		Command:      command,
		VolumeMounts: volumeMounts,
		Lifecycle:    lifecycle,
	}

	template := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label:  "true",
				"name": runtimeName,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				container,
			},
			Volumes: volumes,
			Affinity: &v1.Affinity{
				PodAntiAffinity: &podAntiAffinity,
				NodeAffinity:    ctx.SampleSet.Spec.NodeAffinity.DeepCopy(),
			},
			Tolerations: ctx.SampleSet.Spec.Tolerations,
		},
	}

	selector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			label:  "true",
			"name": runtimeName,
		},
	}
	// construct StatefulSetSpec
	serviceName := j.GetServiceName(ctx.Req.Name)
	replicas := ctx.SampleSet.Spec.Partitions
	spec := appv1.StatefulSetSpec{
		Replicas:            &replicas,
		Selector:            &selector,
		Template:            template,
		ServiceName:         serviceName,
		PodManagementPolicy: appv1.OrderedReadyPodManagement,
	}
	ds.Spec = spec

	return nil
}

func (j *JuiceFS) getVolumeInfo(pv *v1.PersistentVolume) (
	[]v1.Volume, []v1.VolumeMount, *common.ServerOptions, error) {
	// Get cache dir configuration from PV
	if pv.Spec.CSI == nil || pv.Spec.CSI.VolumeAttributes == nil {
		return nil, nil, nil, fmt.Errorf("pv csi field %s is not valid", pv.Name)
	}
	if _, exits := pv.Spec.CSI.VolumeAttributes["mountOptions"]; !exits {
		return nil, nil, nil, fmt.Errorf("pv mountOptions %s is not exist", pv.Name)
	}
	mountOptions := pv.Spec.CSI.VolumeAttributes["mountOptions"]
	optionList := strings.Split(mountOptions, ",")
	serverOpt := &common.ServerOptions{}

	var cacheDirList []string
	for _, option := range optionList {
		if !strings.HasPrefix(option, JuiceFSCacheDirOption) {
			continue
		}
		cacheDir := strings.Split(option, "=")[1]
		cacheDirList = strings.Split(cacheDir, ":")
	}
	if len(cacheDirList) == 0 {
		return nil, nil, nil, fmt.Errorf("cache-dir is not valid")
	}

	hostPathType := v1.HostPathDirectoryOrCreate
	mountPropagation := v1.MountPropagationBidirectional

	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount
	// construct cache host path volume
	for _, path := range cacheDirList {
		pathTrimPrefix := strings.TrimPrefix(path, "/")
		name := strings.ReplaceAll(pathTrimPrefix, "/", "-")

		hostPath := v1.HostPathVolumeSource{
			Path: path,
			Type: &hostPathType,
		}
		hostPathVolume := v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				HostPath: &hostPath,
			},
		}
		volumes = append(volumes, hostPathVolume)
		mountPath := j.getRuntimeCacheMountPath(name)
		volumeMount := v1.VolumeMount{
			Name:             name,
			MountPath:        mountPath,
			MountPropagation: &mountPropagation,
		}
		volumeMounts = append(volumeMounts, volumeMount)
		serverOpt.CacheDirs = append(serverOpt.CacheDirs, mountPath)
	}

	// construct data persistent volume claim source pods
	pvcs := v1.PersistentVolumeClaimVolumeSource{
		ClaimName: pv.Name,
	}
	volume := v1.Volume{
		Name: pv.Name,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &pvcs,
		},
	}
	volumes = append(volumes, volume)

	mountPath := j.getRuntimeDataMountPath(pv.Name)
	volumeMount := v1.VolumeMount{
		Name:             pv.Name,
		MountPath:        mountPath,
		MountPropagation: &mountPropagation,
	}
	volumeMounts = append(volumeMounts, volumeMount)
	serverOpt.DataDir = mountPath

	return volumes, volumeMounts, serverOpt, nil
}

// DoSyncJob sync data from source databases to JuiceFS backend object storage, this job will only work in the
// first runtime server. According to the design concept of container, it is not a good practice to specify --worker
// option when do sync job. When executor sync command in first runtime server, the data will be automatically
// warmed up to this node, this may bring duplicate cached data problem in kubernetes cluster.
// TODO: clean cached data after sync command done or is there a better way?
func (j *JuiceFS) DoSyncJob(ctx context.Context, opt *v1alpha1.SyncJobOptions, log logr.Logger) error {
	syncArgs := utils.NoZeroOptionToArgs(&opt.JuiceFSSyncOptions)

	args := []string{"sync"}
	args = append(args, syncArgs...)
	args = append(args, opt.Source)
	args = append(args, opt.Destination)

	cmd := exec.CommandContext(ctx, "juicefs", args...)
	output, err := cmd.CombinedOutput()
	log.V(1).Info(cmd.String())
	log.V(1).Info(string(output))
	if err != nil {
		return fmt.Errorf("juice sync cmd: %s; error: %s", cmd.String(), err.Error())
	}
	return nil
}

// DoWarmupJob warmup data from remote object storage to cache nodes, this can speed up model training process in kubernetes cluster
// TODO: different cache nodes should warmup different data, the warmup Strategy should match the sampler api
// defined in paddle.io submodule, like RandomSampler/SequenceSampler/DistributedBatchSampler etc...
// More information: https://www.paddlepaddle.org.cn/documentation/docs/zh/api/paddle/io/Overview_cn.html
func (j *JuiceFS) DoWarmupJob(ctx context.Context, opt *v1alpha1.WarmupJobOptions, log logr.Logger) error {
	if len(opt.Paths) == 0 {
		return errors.New("warmup job option paths not set")
	}
	// 1. get the host name of this job pods and extract the pod index
	hostName, err := utils.GetHostName()
	if err != nil {
		return err
	}
	hostNameSplit := strings.Split(hostName, "-")
	if len(hostNameSplit) == 0 {
		return fmt.Errorf("warmup job must run in statefulset pods")
	}
	index, err := strconv.Atoi(hostNameSplit[len(hostNameSplit)-1])
	if err != nil {
		return fmt.Errorf("hostname %s is not valid", hostName)
	}
	runtimeSuffix := "-" + common.RuntimeContainerName + "-" + strconv.Itoa(index)
	sampleSetName := strings.TrimSuffix(hostName, runtimeSuffix)
	mountPath := j.getRuntimeDataMountPath(sampleSetName)

	// 2. preform pre-warmup task and add defer post-warmup task in master node and worker nodes separately
	if index == 0 {
		defer postWarmupMaster(ctx, mountPath, int(opt.Partitions), log)
		if err = preWarmupMaster(ctx, mountPath, opt); err != nil {
			return fmt.Errorf("master node do pre-warmup job error: %s", err.Error())
		}
	} else {
		defer postWarmupWorker(ctx, mountPath, index, log)
		if err = preWarmupWorker(ctx, mountPath, index); err != nil {
			return fmt.Errorf("worker %d do pre-warmup job error: %s", index, err.Error())
		}
	}

	// 3. add --file option and construct warmup job args
	opt.File = mountPath + common.WarmupDirPath + common.WarmupFilePrefix + "." + strconv.Itoa(index)
	warmupArgs := utils.NoZeroOptionToArgs(&opt.JuiceFSWarmupOptions)
	args := []string{"warmup"}
	args = append(args, warmupArgs...)

	// 4. executor juicefs warmup --file xxx paths... command
	cmd := exec.CommandContext(ctx, "juicefs", args...)
	output, err := cmd.CombinedOutput()
	log.V(1).Info(cmd.String())
	log.V(1).Info(string(output))
	if err != nil {
		return fmt.Errorf("juice warmup cmd: %s; error: %s", cmd.String(), err.Error())
	}
	return nil
}

func preWarmupMaster(ctx context.Context, mountPath string, opt *v1alpha1.WarmupJobOptions) error {
	// 1. create .warmup dir in mount path
	warmupPath := mountPath + common.WarmupDirPath
	if err := os.Mkdir(warmupPath, os.ModePerm); err != nil {
		return fmt.Errorf("preWarmupMaster create warmup dir %s error: %s", warmupPath, err.Error())
	}
	partitions := int(opt.Partitions)

	// 2. wait file worker.{index} created by worker nodes util timeout
	for i := 1; i < partitions; i++ {
		workerFile := warmupPath + common.WarmupWorkerPrefix + "." + strconv.Itoa(i)
		utils.WaitFileCreatedWithTimeout(ctx, workerFile, 120*time.Second)
	}

	// 3. if opt.File not set by user, create .warmup file by the command find
	if opt.File == "" {
		filePaths := strings.Join(opt.Paths, " ")
		totalFile := warmupPath + common.WarmupTotalFile
		findCmd := "find " + filePaths + " -type f -not -path '*/\\.*' > " + totalFile
		cmd := exec.CommandContext(ctx, "/bin/bash", "-c", findCmd)
		if stderr, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("preWarmupMaster cmd: %s; error: %s; stderr: %s",
				cmd.String(), err.Error(), string(stderr))
		}
		opt.File = totalFile
	}

	// 4. create tmp files .file.{index} and get writers for them
	tmpPrefix := warmupPath + common.WarmupTmpPrefix + "."
	var tmpFiles []*os.File
	var writers []*bufio.Writer
	for i := 0; i < partitions; i++ {
		path := tmpPrefix + strconv.Itoa(i)
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("preWarmupMaster create file %s error: %s", path, err.Error())
		}
		tmpFiles = append(tmpFiles, file)
		writers = append(writers, bufio.NewWriter(file))
	}

	// 5. create tmp files .file.{index} with different strategy specified in options
	switch opt.Strategy.Name {
	case common.StrategySequence:
		if err := warmupSequentially(opt, writers); err != nil {
			return fmt.Errorf("preWarmupMaster create tmp warmup file error: %s", err.Error())
		}
	case common.StrategyRandom:
		if err := warmupRandomly(opt, writers); err != nil {
			return fmt.Errorf("preWarmupMaster create tmp warmup file error: %s", err.Error())
		}
	default:
		return fmt.Errorf("preWarmupMaster the warmup strategy %s is not supported", opt.Strategy.Name)
	}

	// 6. flush writers and close files
	for i, writer := range writers {
		if err := writer.Flush(); err != nil {
			path := tmpPrefix + strconv.Itoa(i)
			return fmt.Errorf("preWarmupMaster flush file %s error: %s", path, err.Error())
		}
	}
	for _, file := range tmpFiles {
		if err := file.Close(); err != nil {
			return fmt.Errorf("preWarmupMaster close file %s error: %s", file.Name(), err.Error())
		}
	}

	// 7. move tmp file .file.{index} to file.{index}
	targetPrefix := warmupPath + common.WarmupFilePrefix + "."
	for i := 0; i < partitions; i++ {
		tmpPath := tmpPrefix + strconv.Itoa(i)
		targetPath := targetPrefix + strconv.Itoa(i)
		if err := exec.CommandContext(ctx, "mv", tmpPath, targetPath).Run(); err != nil {
			return fmt.Errorf("preWarmupMaster mv %s to %s error: %s", tmpPath, targetPath, err.Error())
		}
	}

	opt.File = warmupPath + common.WarmupFilePrefix + ".0"
	return nil
}

func warmupSequentially(opt *v1alpha1.WarmupJobOptions, writers []*bufio.Writer) error {
	file, err := os.Open(opt.File)
	if err != nil {
		return fmt.Errorf("warmupSequentially open file %s error: %s", opt.File, err.Error())
	}
	defer file.Close()
	var i uint64 = 0

	partition := uint64(opt.Partitions)
	scanner := bufio.NewScanner(file)
	for ; scanner.Scan(); i++ {
		index := i % partition
		if _, err := fmt.Fprintln(writers[index], scanner.Text()); err != nil {
			return fmt.Errorf("warmupSequentially writer partition %d error: %s", index, err.Error())
		}
	}
	return nil
}

func warmupRandomly(opt *v1alpha1.WarmupJobOptions, writers []*bufio.Writer) error {
	file, err := os.Open(opt.File)
	if err != nil {
		return fmt.Errorf("warmupRandomly open file %s error: %s", opt.File, err.Error())
	}
	defer file.Close()

	rand.Seed(time.Now().UnixNano())
	//partition := uint64(opt.Partitions)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		index := rand.Int31n(opt.Partitions)
		if _, err := fmt.Fprintln(writers[index], scanner.Text()); err != nil {
			return fmt.Errorf("warmupRandomly writer partition %d error: %s", index, err.Error())
		}
	}
	return nil
}

func preWarmupWorker(ctx context.Context, mountPath string, index int) error {
	// 1. wait .warmup dir created by master node util timeout
	warmupPath := mountPath + common.WarmupDirPath
	if ok := utils.WaitFileCreatedWithTimeout(ctx, warmupPath, 120*time.Second); !ok {
		return fmt.Errorf("preWarmupWorker wait master node create dir %s timeout", warmupPath)
	}
	// 2. create worker.{index} file in mount path
	workerFile := warmupPath + common.WarmupWorkerPrefix + "." + strconv.Itoa(index)
	if err := exec.CommandContext(ctx, "touch", workerFile).Run(); err != nil {
		return fmt.Errorf("preWarmupWorker create file %s error: %s", workerFile, err.Error())
	}
	// 3. wait util file.{index} created by master node util timeout
	filepath := warmupPath + common.WarmupFilePrefix + "." + strconv.Itoa(index)
	if ok := utils.WaitFileCreatedWithTimeout(ctx, filepath, 600*time.Second); !ok {
		return fmt.Errorf("preWarmupWorker wait master node create file %s timeout", filepath)
	}
	return nil
}

func postWarmupMaster(ctx context.Context, mountPath string, partitions int, log logr.Logger) {
	// 1. check if .warmup dir exists in mount path
	warmupPath := mountPath + common.WarmupDirPath
	if _, err := os.Stat(warmupPath); err != nil {
		log.Error(err, "postWarmupMaster stat path error", "path", warmupPath)
		return
	}

	// 2. check if file worker.{index} exists, and wait file done.{index} created by worker
	donePrefix := warmupPath + common.WarmupDonePrefix + "."
	workerPrefix := warmupPath + common.WarmupWorkerPrefix + "."
	// begin from index 1
	for i := 1; i < partitions; i++ {
		doneFile := donePrefix + strconv.Itoa(i)
		workerFile := workerPrefix + strconv.Itoa(i)
		if _, err := os.Stat(workerFile); err != nil {
			log.V(1).Info("postWarmupMaster worker file not exists", "path", workerFile)
			continue
		}
		for _, err := os.Stat(doneFile); err != nil; _, err = os.Stat(doneFile) {
			log.V(1).Info("postWarmupMaster wait util done file created", "path", doneFile)
			time.Sleep(1 * time.Second)
			if _, e := os.Stat(workerFile); e != nil || ctx.Err() != nil {
				log.V(1).Info("postWarmupMaster worker file has been deleted", "path", workerFile)
				break
			}
		}
		log.V(1).Info("postWarmupMaster done file has been created", "path", doneFile)
	}

	// 3. rm dir .warmup
	if err := exec.Command("rm", "-rf", warmupPath).Run(); err != nil {
		log.Error(err, "postWarmupMaster rm warmup dir error", "path", warmupPath)
		return
	}
	log.V(1).Info("postWarmupMaster done successfully")
}

func postWarmupWorker(ctx context.Context, mountPath string, index int, log logr.Logger) {
	// 1. wait master node create .warmup dir with timeout
	path := mountPath + common.WarmupDirPath
	if ok := utils.WaitFileCreatedWithTimeout(ctx, path, 30*time.Second); !ok {
		log.Error(fmt.Errorf("postWarmupWorker wait master create path %s timeout", path), "")
		return
	}
	// 2. create file with name as .warmup.done.{index}
	filename := path + common.WarmupDonePrefix + "." + strconv.Itoa(index)
	if err := exec.Command("touch", filename).Run(); err != nil {
		log.Error(err, "touch warmup done file error, please touch it manually.", "filename", filename)
		return
	}
}

// DoRmrJob delete the data of JuiceFS storage backend under the specified paths.
// TODO: there some bugs in JuiceFS rmr command, after rmr paths the sync command
// can't work correctly in container, but posix rm can work well with JuiceFS sync command.
func (j *JuiceFS) DoRmrJob(ctx context.Context, opt *v1alpha1.RmrJobOptions, log logr.Logger) error {
	if len(opt.Paths) == 0 {
		return errors.New("rmr job option paths not set")
	}
	for _, path := range opt.Paths {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("path %s is not valid, error: %s", path, err.Error())
		}
	}
	//cmd := exec.CommandContext(ctx,"juicefs", args...)
	rmCmd := "rm -rf " + strings.Join(opt.Paths, " ")
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", rmCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("juice rmr cmd: %s; error: %s", cmd.String(), err.Error())
	}
	return nil
}

// CreateSyncJobOptions create sync job options by the information from request context, the options is used by
// controller to request runtime server do sync data task asynchronously.
// TODO: Support different uri format for all storage in JuiceFSSupportStorage,
// some data storage may need additional secret setting in v1alpha1.Source.SecretRef
// more info: https://github.com/juicedata/juicesync
func (j *JuiceFS) CreateSyncJobOptions(opt *v1alpha1.SyncJobOptions, ctx *common.RequestContext) error {
	// if SampleJob is not nil, use sync job options from SampleJob
	if ctx.SampleJob != nil && ctx.SampleJob.Spec.SyncOptions != nil {
		ctx.SampleJob.Spec.SyncOptions.DeepCopyInto(opt)
	}
	// if SampleJob is nil or source has not been set, use source from SampleSet
	if opt.Source == "" && ctx.SampleSet.Spec.Source != nil {
		opt.Source = ctx.SampleSet.Spec.Source.URI
	}
	// if source has not been set in SampleSet or SampleJob return error
	if opt.Source == "" {
		return fmt.Errorf("data source cannot be empty")
	}
	// verify the format of data source uri
	if !strings.Contains(opt.Source, "://") || len(strings.Split(opt.Source, "://")) != 2 {
		return fmt.Errorf("the format of data source uri is not support")
	}
	// verify the data source storage is supported
	storage := strings.TrimSpace(strings.Split(opt.Source, "://")[0])
	if !utils.ContainsString(JuiceFSSupportStorage, storage) {
		return fmt.Errorf("the object storage %s of is not support", storage)
	}

	var secretKeys string
	delimiter := "://"
	if strings.Contains(opt.Source, ":///") && len(strings.Split(opt.Source, ":///")) == 2 {
		delimiter = ":///"
	}
	path := strings.TrimSpace(strings.Split(opt.Source, delimiter)[1])
	// add access key to source uri if exists in secret
	if akByte, exist := ctx.Secret.Data[JuiceFSSecretAK]; exist && len(akByte) > 0 {
		secretKeys = string(akByte)
	}
	// add secret key to source uri if exists in secret
	if skByte, exist := ctx.Secret.Data[JuiceFSSecretSK]; exist && len(skByte) > 0 {
		secretKeys = secretKeys + ":" + string(skByte)
	}
	if secretKeys != "" {
		opt.Source = storage + delimiter + secretKeys + "@" + path
	}

	// add relative path to sync destination uri
	mountPath := "file://" + j.getRuntimeDataMountPath(ctx.SampleSet.Name)
	if opt.Destination != "" {
		opt.Destination = mountPath + "/" + strings.TrimPrefix(opt.Destination, "/")
	} else {
		opt.Destination = mountPath
	}

	// source and destination should both end with / or not
	if strings.HasSuffix(opt.Source, "/") != strings.HasSuffix(opt.Destination, "/") {
		opt.Source = strings.TrimSuffix(opt.Source, "/") + "/"
		opt.Destination = strings.TrimSuffix(opt.Destination, "/") + "/"
	}

	return nil
}

func (j *JuiceFS) CreateWarmupJobOptions(opt *v1alpha1.WarmupJobOptions, ctx *common.RequestContext) error {
	// if SampleJob is not nil, use sync job options from SampleJob
	if ctx.SampleJob != nil && ctx.SampleJob.Spec.WarmupOptions != nil {
		ctx.SampleJob.Spec.WarmupOptions.DeepCopyInto(opt)
	}
	mountPath := j.getRuntimeDataMountPath(ctx.Req.Name)
	if len(opt.Paths) != 0 {
		var validPaths []string
		for _, path := range opt.Paths {
			validPath := mountPath + "/" + strings.TrimPrefix(path, "/")
			validPaths = append(validPaths, validPath)
		}
		opt.Paths = validPaths
	} else {
		opt.Paths = append(opt.Paths, mountPath)
	}
	if opt.Strategy.Name == "" {
		opt.Strategy.Name = common.StrategySequence
	}
	if opt.File != "" {
		opt.File = mountPath + "/" + strings.TrimPrefix(opt.File, "/")
	}
	opt.Partitions = ctx.SampleSet.Spec.Partitions
	return nil
}
