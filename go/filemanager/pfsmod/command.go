package pfsmod

import (
	"encoding/json"
	"errors"
	"log"
)

type Command interface {
	ToUrl() string
	ToJson() []byte
	Run() (interface{}, error)
}

func IsCloudPath(path string) bool {
	return strings.HasPrefix(path, "/pfs/")
}

const (
	DefaultMultiPartBoundary = "8d7b0e5709d756e21e971ff4d9ac3b20"
)

const (
	defaultMaxCreateFileSize = int64(4 * 1024 * 1024 * 1024)
)

const (
	MaxJsonRequestSize = 2048
)
