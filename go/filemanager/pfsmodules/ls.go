package pfsmodules

import (
	"errors"
	"flag"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/golang/glog"
)

const (
	lsCmdName = "ls"
)

// LsResult represents a LsCmd's result.
type LsResult struct {
	Path    string `json:"Path"`
	ModTime int64  `json:"ModTime"`
	Size    int64  `json:"Size"`
	IsDir   bool   `json:"IsDir"`
}

// LsCmd means LsCommand structure.
type LsCmd struct {
	Method string
	R      bool
	Args   []string
}

// ToURLParam encoding LsCmd to URL Encoding string.
func (p *LsCmd) ToURLParam() url.Values {
	parameters := url.Values{}
	parameters.Add("method", p.Method)
	parameters.Add("r", strconv.FormatBool(p.R))

	for _, arg := range p.Args {
		parameters.Add("arg", arg)
	}

	return parameters
}

// ToJSON does't need to be implemented.
func (p *LsCmd) ToJSON() ([]byte, error) {
	panic("not implemented")
}

// NewLsCmdFromFlag returen a new LsCmd.
func NewLsCmdFromFlag(f *flag.FlagSet) (*LsCmd, error) {
	cmd := LsCmd{}

	cmd.Method = lsCmdName
	cmd.Args = make([]string, 0, f.NArg())

	var err error
	f.Visit(func(flag *flag.Flag) {
		if flag.Name == "r" {
			cmd.R, err = strconv.ParseBool(flag.Value.String())
			if err != nil {
				panic(err)
			}
		}
	})

	for _, arg := range f.Args() {
		log.V(2).Info(arg)
		cmd.Args = append(cmd.Args, arg)
	}

	return &cmd, nil
}

// NewLsCmdFromURLParam returns a new LsCmd according path variable.
func NewLsCmdFromURLParam(path string) (*LsCmd, error) {
	cmd := LsCmd{}

	m, err := url.ParseQuery(path)
	if err != nil {
		return nil, err
	}

	if len(m["method"]) == 0 ||
		len(m["r"]) == 0 ||
		len(m["arg"]) == 0 {
		return nil, errors.New(StatusNotEnoughArgs)
	}

	cmd.Method = m["method"][0]
	if cmd.Method != lsCmdName {
		return nil, errors.New(http.StatusText(http.StatusMethodNotAllowed) + ":" + cmd.Method)
	}

	cmd.R, err = strconv.ParseBool(m["r"][0])
	if err != nil {
		return nil, errors.New(StatusInvalidArgs + ":r")
	}

	cmd.Args = m["arg"]

	return &cmd, nil
}

// NewLsCmd return a new LsCmd according r and path variable.
func NewLsCmd(r bool, path string) *LsCmd {
	return &LsCmd{
		Method: lsCmdName,
		R:      r,
		Args:   []string{path},
	}
}

func lsPath(path string, r bool) ([]LsResult, error) {
	var ret []LsResult

	err := filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		m := LsResult{}
		m.Path = subpath
		m.Size = info.Size()
		m.ModTime = info.ModTime().UnixNano()
		m.IsDir = info.IsDir()

		if subpath == path {
			if info.IsDir() {
			} else {
				ret = append(ret, m)
			}
		} else {
			ret = append(ret, m)
		}

		if info.IsDir() && !r && subpath != path {
			return filepath.SkipDir
		}

		return nil
	})

	return ret, err
}

// CloudCheck checks the conditions when running on cloud.
func (p *LsCmd) ValidateCloudArgs() error {
	return ValidatePfsPath(p.Args)
}

// LocalCheck checks the conditions when running local.
func (p *LsCmd) ValidateLocalArgs() error {
	if len(p.Args) == 0 {
		return errors.New(StatusNotEnoughArgs)
	}
	return nil
}

// Run functions runs LsCmd and return LsResult and error if any happened.
func (p *LsCmd) Run() (interface{}, error) {
	var results []LsResult

	for _, arg := range p.Args {
		log.V(1).Infof("ls %s\n", arg)

		list, err := filepath.Glob(arg)
		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			return results, errors.New(StatusFileNotFound)
		}

		for _, path := range list {
			ret, err := lsPath(path, p.R)
			if err != nil {
				return results, err
			}
			results = append(results, ret...)
		}
	}

	return results, nil
}
