package pfsmod

import (
	"errors"
	"flag"
	log "github.com/golang/glog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

const (
	lsCmdName = "ls"
)

type LsResult struct {
	Path    string `json:"Path"`
	ModTime string `json:"ModTime"`
	Size    int64  `json:"Size"`
	IsDir   bool   `json:IsDir`
}

type LsCmd struct {
	Method string
	R      bool
	Args   []string
}

func (p *LsCmd) ToUrlParam() string {
	parameters := url.Values{}
	parameters.Add("method", p.Method)
	parameters.Add("r", strconv.FormatBool(p.R))

	for _, arg := range p.Args {
		parameters.Add("arg", arg)
	}

	return parameters.Encode()

}

func (p *LsCmd) ToJson() ([]byte, error) {
	return nil, nil
}

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

func NewLsCmdFromUrlParam(path string) (*LsCmd, error) {
	cmd := LsCmd{}

	m, err := url.ParseQuery(path)
	if err != nil ||
		len(m["method"]) == 0 ||
		len(m["r"]) == 0 ||
		len(m["arg"]) == 0 {
		return nil, errors.New(StatusText(StatusNotEnoughArgs))
	}

	cmd.Method = m["method"][0]
	if cmd.Method != lsCmdName {
		return nil, errors.New(http.StatusText(http.StatusMethodNotAllowed) + ":" + cmd.Method)
	}

	cmd.R, err = strconv.ParseBool(m["r"][0])
	if err != nil {
		return nil, errors.New(StatusText(StatusInvalidArgs) + ":r")
	}

	cmd.Args = make([]string, 0, len(m["arg"])+1)
	for _, arg := range m["arg"] {
		cmd.Args = append(cmd.Args, arg)
	}

	return &cmd, nil
}

func NewLsCmd(r bool, path string) *LsCmd {
	return &LsCmd{
		Method: lsCmdName,
		R:      r,
		Args:   []string{path},
	}
}

func lsPath(path string, r bool) ([]LsResult, error) {
	ret := make([]LsResult, 0, 100)

	filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		//log.Println("path:\t" + path)

		if err != nil {
			return err
		}

		m := LsResult{}
		m.Path = subpath
		m.Size = info.Size()
		m.ModTime = info.ModTime().Format("2006-01-02 15:04:05")
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

		//log.Println(len(metas))
		return nil
	})

	return ret, nil
}

func (p *LsCmd) CloudCheck() error {
	if len(p.Args) == 0 {
		return errors.New(StatusText(StatusNotEnoughArgs))
	}

	for _, arg := range p.Args {
		if !IsCloudPath(arg) {
			return errors.New(StatusText(StatusShouldBePfsPath) + ":" + arg)
		}

		if !CheckUser(arg) {
			return errors.New(StatusText(StatusShouldBePfsPath) + ":" + arg)
		}
	}

	return nil
}

func (p *LsCmd) LocalCheck() error {
	if len(p.Args) == 0 {
		return errors.New(StatusText(StatusNotEnoughArgs))
	}
	return nil
}

func (p *LsCmd) Run() (interface{}, error) {
	results := make([]LsResult, 0, 100)

	for _, arg := range p.Args {
		log.V(1).Infof("ls %s\n", arg)

		list, err := filepath.Glob(arg)
		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			return results, errors.New(StatusText(StatusFileNotFound))
			break
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
