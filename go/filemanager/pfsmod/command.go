package pfsmod

import (
	"strings"
)

const (
	DefaultMultiPartBoundary = "8d7b0e5709d756e21e971ff4d9ac3b20"
)

const (
	MaxJsonRequestSize = 2048
)

type Command interface {
	ToUrlParam() string
	ToJson() ([]byte, error)
	Run() (interface{}, error)
}

func IsCloudPath(path string) bool {
	return strings.HasPrefix(path, "/pfs/")
}

func CheckUser(path string) bool {
	//TODO
	return true
}
