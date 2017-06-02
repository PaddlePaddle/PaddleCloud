package pfsmodules

import (
	"errors"
	"io"
	"net/url"
	"strings"

	log "github.com/golang/glog"
)

const (
	// DefaultMultiPartBoundary is the default multipart form boudary.
	DefaultMultiPartBoundary = "8d7b0e5709d756e21e971ff4d9ac3b20"

	// MaxJSONRequestSize is the max body size when server receives a request.
	MaxJSONRequestSize = 2048
)

// Command is a interface of all commands.
type Command interface {
	ToURLParam() url.Values
	ToJSON() ([]byte, error)
	Run() (interface{}, error)
	ValidateLocalArgs() error
	ValidateCloudArgs() error
}

// getUserName gets user's name by token
func getUserName(url string, query url.Values, token string) (string, error) {
	// TODO
	return "", nil
}

// CheckUser checks if a user has authority to access a path.
func checkUser(path string) error {
	// TODO
	return nil
}

// IsCloudPath returns whether a path is a pfspath.
func ValidatePfsPath(paths []string) error {
	if len(paths) == 0 {
		return errors.New(StatusNotEnoughArgs)
	}

	for _, path := range paths {
		if !strings.HasPrefix(path, "/pfs/") {
			return errors.New(StatusShouldBePfsPath + ":" + path)
		}

		if err := checkUser(path); err != nil {
			return errors.New(StatusShouldBePfsPath + ":" + path)
		}
	}
	return nil
}

// Close closes c and log it.
func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Error(err)
	}
}
