package updater

import (
	"fmt"
	"strings"
)

// Labels represents labels of k8s resources
type Labels map[string]string

// LabelsParser parse labels to selector
func (l Labels) LabelsParser() (string, error) {
	pieces := make([]string, 0, len(l))
	for k, v := range l {
		pieces = append(pieces, fmt.Sprintf("%v=%v", k, v))
	}
	return strings.Join(pieces, ","), nil
}

// HasLabel check if has the label
func (l Labels) HasLabel(labname string) (bool, string) {
	for k, v := range l {
		if k == labname {
			return true, v
		}
	}
	return false, ""
}
