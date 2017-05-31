package pfsmodules

import (
	"io"
	"strings"

	log "github.com/golang/glog"
)

const (
	// DefaultMultiPartBoundary is the default multipart form boudary
	DefaultMultiPartBoundary = "8d7b0e5709d756e21e971ff4d9ac3b20"
)

const (
	// MaxJSONRequestSize is the max body size when server receives a request
	MaxJSONRequestSize = 2048
)

// Command is a interface of all commands
type Command interface {
	ToURLParam() string
	ToJSON() ([]byte, error)
	Run() (interface{}, error)
	LocalCheck() error
	CloudCheck() error
}

// IsCloudPath returns whether a path is a pfspath
func IsCloudPath(path string) bool {
	return strings.HasPrefix(path, "/pfs/")
}

// CheckUser checks if a user has authority to access a path
func CheckUser(path string) bool {
	//TODO
	return true
}

// Close closes c and log it
func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Error(err)
	}
}
