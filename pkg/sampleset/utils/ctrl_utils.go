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

package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// HasDeletionTimestamp method to check if an object need to delete.
func HasDeletionTimestamp(obj *metav1.ObjectMeta) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}

// HasFinalizer check
func HasFinalizer(obj *metav1.ObjectMeta, finalizer string) bool {
	return ContainsString(obj.GetFinalizers(), finalizer)
}

func RemoveFinalizer(obj *metav1.ObjectMeta, finalizer string) {
	finalizers := RemoveString(obj.Finalizers, finalizer)
	obj.Finalizers = finalizers
}

func NoRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func RequeueWithError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

func RequeueAfter(requeueAfter time.Duration) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: requeueAfter}, nil

}

// ContainsString Determine whether the string array contains a specific string,
// return true if contains the string and return false if not.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func NoZeroOptionToMap(optionMap map[string]reflect.Value, i interface{}) {
	elem := reflect.ValueOf(i).Elem()
	for i := 0; i < elem.NumField(); i++ {
		value := elem.Field(i)
		if value.IsZero() {
			continue
		}
		field := elem.Type().Field(i)
		tag := field.Tag.Get("json")
		option := strings.Split(tag, ",")[0]
		optionMap[option] = value
	}
}

func NoZeroOptionToArgs(options interface{}) []string {
	var args []string

	elem := reflect.ValueOf(options).Elem()
	for i := 0; i < elem.NumField(); i++ {
		v := elem.Field(i)
		if v.IsZero() {
			continue
		}
		field := elem.Type().Field(i)
		tag := field.Tag.Get("json")
		opt := strings.Split(tag, ",")[0]
		switch v.Kind() {
		case reflect.Bool:
			args = append(args, fmt.Sprintf("--%s", opt))
		case reflect.Slice: // []string
			for j := 0; j < v.Len(); j++ {
				args = append(args, fmt.Sprintf(`--%s=%v`, opt, v.Index(j)))
			}
		default:
			args = append(args, fmt.Sprintf("--%s=%v", opt, v))
		}
	}
	return args
}

func DiskUsageOfPaths(timeout time.Duration, paths ...string) (string, error) {
	filePaths := strings.Join(paths, " ")
	arg := "du -scb --exclude \"./.*\" " + filePaths

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "bash", "-c", arg)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cmd:%s, error: %s", cmd.String(), stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("cmd:%s, output:%s", cmd.String(), stdout.String())
	}
	total := strings.TrimSpace(lines[len(lines)-1])
	if !strings.Contains(total, "total") {
		return "", fmt.Errorf("cmd:%s, output:%s", cmd.String(), stdout.String())
	}
	totalSlice := strings.FieldsFunc(total, func(r rune) bool { return r == ' ' || r == '\t' })
	if len(totalSlice) == 0 {
		return "", fmt.Errorf("cmd:%s, output:%s", cmd.String(), stdout.String())
	}
	totalSizeStr := strings.TrimSpace(totalSlice[0])
	_, err := strconv.ParseUint(totalSizeStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("cmd:%s, parseUint error: %s", cmd.String(), err.Error())
	}

	return totalSizeStr, nil
}

func FileNumberOfPaths(timeout time.Duration, paths ...string) (string, error) {
	filePaths := strings.Join(paths, " ")
	arg := "ls -lR " + filePaths + "| grep \"^-\" | wc -l"

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "bash", "-c", arg)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cmd:%s, error: %s", cmd.String(), stderr.String())
	}
	fileNum, err := strconv.ParseInt(strings.TrimSpace(stdout.String()), 10, 64)
	if err != nil {
		return "", fmt.Errorf("cmd:%s, parseUint error: %s", cmd.String(), err)
	}

	return humanize.Comma(fileNum), nil
}

func JuiceFileNumberOfPath(timeout time.Duration, path ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	args := []string{"info"}
	args = append(args, path...)

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "juicefs", args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cmd:%s, error: %s", cmd.String(), stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("cmd:%s, output:%s", cmd.String(), stdout.String())
	}
	var fileLines []string
	for _, line := range lines {
		if strings.Contains(line, "files") {
			fileLines = append(fileLines, line)
		}
	}

	var fileNumTotal int64
	for _, fileLine := range fileLines {
		if fileLine == "" || len(strings.Split(fileLine, ":")) != 2 {
			return "", fmt.Errorf("cmd:%s, fileLine:%s", cmd.String(), fileLine)
		}
		fileNumStr := strings.TrimSpace(strings.Split(fileLine, ":")[1])
		fileNum, err := strconv.ParseInt(fileNumStr, 10, 64)
		if err != nil {
			return "", fmt.Errorf("cmd:%s, parseInt error: %s", cmd.String(), err)
		}
		fileNumTotal += fileNum

	}

	return humanize.Comma(fileNumTotal), nil
}

func DiskSpaceOfPaths(timeout time.Duration, paths ...string) (map[string]string, error) {
	args := []string{"--output=fstype,size,used,avail", "-BK"}
	args = append(args, paths...)

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "df", args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cmd:%s, error: %s", cmd.String(), stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) <= 1 {
		return nil, fmt.Errorf("cmd:%s, output:%s", cmd.String(), stdout.String())
	}

	infoMap := make(map[string][]string)
	for _, line := range lines[1:] {
		infoStr := strings.TrimSpace(line)
		infoList := strings.FieldsFunc(infoStr, func(r rune) bool { return r == ' ' || r == '\t' })
		if len(infoList) != 4 {
			return nil, fmt.Errorf("cmd:%s, output:%s", cmd.String(), stdout.String())
		}
		fsType := strings.TrimSpace(infoList[0])
		infoMap[fsType] = infoList
	}

	var diskSizeTotal, diskUsedTotal, diskAvailTotal uint64
	for _, infoList := range infoMap {
		diskSizeStr := strings.TrimSuffix(strings.TrimSpace(infoList[1]), "K")
		diskSize, err := strconv.ParseUint(diskSizeStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cmd:%s, output:%s", cmd.String(), stdout.String())
		}
		diskSizeTotal += diskSize

		diskUsedStr := strings.TrimSuffix(strings.TrimSpace(infoList[2]), "K")
		diskUsed, err := strconv.ParseUint(diskUsedStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cmd:%s, output:%s", cmd.String(), stdout.String())
		}
		diskUsedTotal += diskUsed

		diskAvailStr := strings.TrimSuffix(strings.TrimSpace(infoList[3]), "K")
		diskAvail, err := strconv.ParseUint(diskAvailStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cmd:%s, output:%s", cmd.String(), stdout.String())
		}
		diskAvailTotal += diskAvail
	}

	diskStatus := map[string]string{
		"diskSize":  strconv.FormatUint(diskSizeTotal, 10),
		"diskUsed":  strconv.FormatUint(diskUsedTotal, 10),
		"diskAvail": strconv.FormatUint(diskAvailTotal, 10),
	}

	return diskStatus, nil
}

func WaitFileCreatedWithTimeout(ctx context.Context, path string, duration time.Duration) bool {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		for _, err := os.Stat(path); err != nil; _, err = os.Stat(path) {
			if os.IsNotExist(err) {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			break
		}
		done <- true
	}()
	for {
		select {
		case <-done:
			return true
		case <-ctx.Done():
			return false
		case <-ticker.C:
			return false
		}
	}
}

func GetRuntimeImage() (string, error) {
	image := os.Getenv("RUNTIME_IMAGE")
	if image == "" {
		return "", errors.New("RUNTIME_IMAGE is not in environment variable")
	}
	return image, nil
}

func GetHostName() (string, error) {
	hostName := os.Getenv("HOSTNAME")
	if hostName == "" {
		return "", errors.New("HOSTNAME is not in environment variable")
	}
	return hostName, nil
}
