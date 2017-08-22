package pfsmodules

import (
	"errors"
	"io"
	"net/url"
	"path"
	"strings"

	log "github.com/golang/glog"
)

const (
	// DefaultMultiPartBoundary is the default multipart form boudary.
	DefaultMultiPartBoundary = "8d7b0e5709d756e21e971ff4d9ac3b20"

	// MaxJSONRequestSize is the max body size when server receives a request.
	MaxJSONRequestSize = 2048
)

const (
	// RESTChunksStoragePath is chunk's storage path of REST API.
	RESTChunksStoragePath = "api/v1/pfs/storage/chunks"
	// RESTFilesPath is files' path of REST API.
	RESTFilesPath = "api/v1/pfs/files"
	// RESTChunksPath is chunks' path of REST API.
	RESTChunksPath = "api/v1/pfs/chunks"
)

// Command is a interface of all commands.
type Command interface {
	// ToURLParam generates url.Values of the command struct.
	ToURLParam() url.Values
	// ToJSON generates JSON string of the command struct.
	ToJSON() ([]byte, error)
	// Run runs a command.
	Run() (interface{}, error)
	// ValidateLocalArgs validates arguments when running locally.
	ValidateLocalArgs() error
	// ValidateCloudArgs validates arguments when running on cloud.
	ValidateCloudArgs(userName string) error
}

// CheckUser checks if a user has authority to access a path.
// path example:/pfs/$datacenter/home/$user
func checkUser(pathStr string, user string) error {
	pathStr = path.Clean(strings.TrimSpace(pathStr))
	a := strings.Split(pathStr, "/")
	// the first / is convert to " "
	if len(a) < 5 {
		return errors.New(StatusBadPath)
	}

	if a[4] != user {
		log.V(4).Infof("request path:%s user:%s split_path:%s\n", pathStr, user, a[4])
		return errors.New(StatusUnAuthorized)
	}
	return nil
}

func isPublic(pathStr string) bool {
	pathStr = path.Clean(strings.TrimSpace(pathStr))
	a := strings.Split(pathStr, "/")

	if len(a) >= 4 && a[3] == "public" {
		return true
	}

	return false
}

func checkPublic(cmdName string) error {
	switch cmdName {
	case "ls", "stat":
		return nil
	default:
		return errors.New("public data supports only ls or stat command")
	}
}

// ValidatePfsPath returns whether a pfspath is valid and autorized
func ValidatePfsPath(paths []string, userName string, cmdName string) error {
	if len(paths) == 0 {
		return errors.New(StatusNotEnoughArgs)
	}

	for _, path := range paths {
		if !strings.HasPrefix(path, "/pfs/") {
			return errors.New(StatusShouldBePfsPath + ":" + path)
		}

		if isPublic(path) {
			if err := checkPublic(cmdName); err != nil {
				return err
			}
			continue
		}

		if err := checkUser(path, userName); err != nil {
			return err
		}
	}
	return nil
}

// IsCloudPath returns whether a path is a pfspath.
func IsCloudPath(path string) bool {
	return strings.HasPrefix(path, "/pfs/")
}

// Close closes c and log it.
func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Error(err)
	}
}
