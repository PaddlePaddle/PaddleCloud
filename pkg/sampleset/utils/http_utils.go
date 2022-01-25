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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/paddleflow/paddle-operator/api/v1alpha1"
	"github.com/paddleflow/paddle-operator/controllers/extensions/common"
)

var DefaultClient = &HttpClient{}

type HttpClient struct {
	http.Client
}

func (c *HttpClient) Post(url string, filename types.UID, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("filename", string(filename))
	req.Header.Set("Content-Type", "application/json")
	return c.Do(req)
}

func Post(url string, filename types.UID, body io.Reader) (resp *http.Response, err error) {
	return DefaultClient.Post(url, filename, body)
}

func Get(url string, filename types.UID) (resp *http.Response, err error) {
	return DefaultClient.Get(url + "/" + string(filename))
}

func GetBaseUriByIndex(runtimeName, serviceName string, index int) string {
	return fmt.Sprintf("http://%s-%d.%s:%d", runtimeName, index, serviceName, common.RuntimeServicePort)
}

func GetBaseUriByName(runtimePodName, serviceName string) string {
	return fmt.Sprintf("http://%s.%s:%d", runtimePodName, serviceName, common.RuntimeServicePort)
}

func GetUploadUri(baseUri, uploadPath string) string {
	return baseUri + common.PathUploadPrefix + uploadPath
}

func GetResultUri(baseUri, resultPath string) string {
	return baseUri + resultPath
}

func GetOptionUri(baseUri, optionPath string) string {
	return baseUri + optionPath
}

func GetCacheStatus(runtimePodName, serviceName string) (*v1alpha1.CacheStatus, error) {
	baseUri := GetBaseUriByName(runtimePodName, serviceName)
	resultUri := GetResultUri(baseUri, common.PathCacheStatus)
	resp, err := Get(resultUri, common.FilePathCacheInfo)
	if err != nil {
		return nil, fmt.Errorf("get uri %s, error: %s", resultUri, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resp status code not ok: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp body error: %s", err.Error())
	}
	defer resp.Body.Close()

	status := &v1alpha1.CacheStatus{}
	err = json.Unmarshal(body, status)
	if err != nil {
		return nil, fmt.Errorf("unmarshal resp body error: %s", err.Error())
	}
	return status, nil
}

type CacheStatusResult struct {
	Status *v1alpha1.CacheStatus
	Error  error
}

func CollectAllCacheStatus(runtimePodNames []string, serviceName string) (*v1alpha1.CacheStatus, error) {
	resultChan := make(chan *CacheStatusResult, len(runtimePodNames))

	for _, runtimePodName := range runtimePodNames {
		go func(podName string) {
			status, err := GetCacheStatus(podName, serviceName)
			if err != nil {
				err = fmt.Errorf("get cache status from server %s error: %s", podName, err.Error())
			}
			cacheStatusResult := &CacheStatusResult{Status: status, Error: err}
			resultChan <- cacheStatusResult
		}(runtimePodName)
	}

	var errStrList []string
	var statusList []*v1alpha1.CacheStatus

	for i := 0; i < len(runtimePodNames); i++ {
		result := <-resultChan
		if result.Error != nil {
			return nil, result.Error
		}
		statusList = append(statusList, result.Status)
	}
	close(resultChan)

	statusAll := &v1alpha1.CacheStatus{}
	var totalSize, cachedSizeTotal, diskSizeTotal, diskAvailTotal, diskUsedTotal uint64

	for _, status := range statusList {
		if status.TotalSize != "" && totalSize == 0 {
			totalSize, _ = strconv.ParseUint(status.TotalSize, 10, 64)
		}
		if status.TotalFiles != "" && statusAll.TotalFiles == "" {
			statusAll.TotalFiles = status.TotalFiles
		}

		cachedSize, _ := strconv.ParseUint(status.CachedSize, 10, 64)
		cachedSizeTotal += cachedSize

		diskSize, _ := strconv.ParseUint(status.DiskSize, 10, 64)
		diskSizeTotal += diskSize

		diskAvail, _ := strconv.ParseUint(status.DiskAvail, 10, 64)
		diskAvailTotal += diskAvail

		diskUsed, _ := strconv.ParseUint(status.DiskUsed, 10, 64)
		diskUsedTotal += diskUsed

		if status.ErrorMassage != "" {
			errStrList = append(errStrList, status.ErrorMassage)
		}
	}
	statusAll.TotalSize = humanize.IBytes(totalSize)
	// If cacheSize is bigger than totalSize may cause some confusion
	if totalSize >= cachedSizeTotal {
		statusAll.CachedSize = humanize.IBytes(cachedSizeTotal)
	} else {
		statusAll.CachedSize = statusAll.TotalSize
	}
	statusAll.DiskSize = humanize.IBytes(diskSizeTotal * 1024)
	statusAll.DiskUsed = humanize.IBytes(diskUsedTotal * 1024)
	statusAll.DiskAvail = humanize.IBytes(diskAvailTotal * 1024)

	var diskUsageRate float64
	if diskSizeTotal == 0 {
		diskUsageRate = 0.0
	} else {
		diskUsageRate = float64(diskUsedTotal) / float64(diskSizeTotal) * 100
	}
	statusAll.DiskUsageRate = fmt.Sprintf("%.1f%%", diskUsageRate)

	if len(errStrList) != 0 {
		statusAll.ErrorMassage = strings.Join(errStrList, "; ")
	}

	return statusAll, nil
}

func GetJobResult(filename types.UID, baseUri, resultPath string) (*common.JobResult, error) {
	resultUri := GetResultUri(baseUri, resultPath)

	resp, err := Get(resultUri, filename)
	if err != nil {
		return nil, fmt.Errorf("get uri %s, filename: %s, error: %s", resultUri, filename, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resp status code not ok: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp body error: %s", err.Error())
	}
	defer resp.Body.Close()

	result := &common.JobResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal resp body error: %s", err.Error())
	}
	return result, nil
}

func GetJobOption(option interface{}, filename types.UID, baseUri, optionPath string) error {
	optionUri := GetOptionUri(baseUri, optionPath)

	resp, err := Get(optionUri, filename)
	if err != nil {
		return fmt.Errorf("get uri %s, filename: %s, error: %s", optionUri, filename, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("resp status code not ok: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read resp body error: %s", err.Error())
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, option)
	if err != nil {
		return fmt.Errorf("unmarshal resp body error: %s", err.Error())
	}
	return nil
}

func PostJobOption(option interface{}, filename types.UID, baseUri, optionPath, param string) error {
	body, err := json.Marshal(option)
	if err != nil {
		return fmt.Errorf("marshal option %+v error: %s", option, err.Error())
	}
	uploadUri := GetUploadUri(baseUri, optionPath)
	if param != "" {
		uploadUri = uploadUri + param
	}
	resp, err := Post(uploadUri, filename, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("post uri %s, filename: %s, error: %s", uploadUri, filename, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("resp status code not ok: %d", resp.StatusCode)
	}

	return nil
}
