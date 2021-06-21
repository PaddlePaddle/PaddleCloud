/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package portparse

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePortNumAnno parse an object's host annotation
func ParsePortNumAnno(anno map[string]string) (int, error) {
	var portNum string
	var ok bool
	if portNum, ok = anno["hostport-manager/portnum"]; !ok {
		return 0, nil
	}
	num, err := strconv.Atoi(portNum)
	if err != nil {
		return 0, fmt.Errorf("Annotation hostport-manager/portnum %s is not valid", portNum)
	}
	return num, nil
}

// ParsePortsAllocAnno parse allocated ports
func ParsePortsAllocAnno(anno map[string]string) ([]int, error) {
	var ports string
	var ok bool
	if ports, ok = anno["hostport-manager/hostport"]; !ok {
		return nil, nil
	}
	ps := strings.Split(ports, ",")
	ret := make([]int, 0)
	for _, port := range ps {
		portnum, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("Port %s is not valid", port)
		}
		ret = append(ret, portnum)
	}
	return ret, nil
}
