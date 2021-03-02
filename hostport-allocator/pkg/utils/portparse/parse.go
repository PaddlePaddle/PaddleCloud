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
