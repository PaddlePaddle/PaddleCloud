package pfsmodules

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/PaddlePaddle/cloud/go/utils/restclient"
	log "github.com/golang/glog"
	"github.com/google/subcommands"
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

// LsCmd means LsCmd structure.
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

func newLsCmdFromFlag(f *flag.FlagSet) (*LsCmd, error) {
	cmd := LsCmd{}

	cmd.Method = lsCmdName
	cmd.Args = make([]string, 0, f.NArg())

	var err error
	f.Visit(func(flag *flag.Flag) {
		if flag.Name == "r" {
			cmd.R, err = strconv.ParseBool(flag.Value.String())
			if err != nil {
				log.Errorln("meets error when parsing argument r")
				return
			}
		}
	})

	if err != nil {
		return nil, err
	}

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

// ValidateCloudArgs checks the conditions when running on cloud.
func (p *LsCmd) ValidateCloudArgs(userName string) error {
	return ValidatePfsPath(p.Args, userName, lsCmdName)
}

// ValidateLocalArgs checks the conditions when running local.
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

// Name returns LsCmd's name.
func (*LsCmd) Name() string { return "ls" }

// Synopsis returns Synopsis of LsCmd.
func (*LsCmd) Synopsis() string { return "List files on PaddlePaddle Cloud" }

// Usage returns usage of LsCmd.
func (*LsCmd) Usage() string {
	return `ls [-r] <pfspath>:
	List files on PaddlePaddleCloud
	Options:
`
}

// SetFlags sets LsCmd's parameters.
func (p *LsCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.R, "r", false, "list files recursively")
}

// getFormatPrint gets max width of filesize and return format string to print.
func getFormatString(result []LsResult) string {
	max := 0
	for _, t := range result {
		str := fmt.Sprintf("%d", t.Size)

		if len(str) > max {
			max = len(str)
		}
	}

	return fmt.Sprintf("%%s %%s %%%dd %%s\n", max)
}

func formatPrint(result []LsResult) {
	formatStr := getFormatString(result)

	for _, t := range result {
		timeStr := time.Unix(0, t.ModTime).Format("2006-01-02 15:04:05")

		if t.IsDir {
			fmt.Printf(formatStr, timeStr, "d", t.Size, t.Path)
		} else {
			fmt.Printf(formatStr, timeStr, "f", t.Size, t.Path)
		}
	}

	fmt.Printf("\n")
}

// RemoteLs gets LsCmd result from cloud.
func RemoteLs(cmd *LsCmd) ([]LsResult, error) {
	t := fmt.Sprintf("%s/%s", Config.ActiveConfig.Endpoint, RESTFilesPath)
	body, err := restclient.GetCall(t, cmd.ToURLParam())
	if err != nil {
		return nil, err
	}

	type lsResponse struct {
		Err     string     `json:"err"`
		Results []LsResult `json:"results"`
	}

	resp := lsResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp.Results, err
	}

	if len(resp.Err) == 0 {
		return resp.Results, nil
	}

	return resp.Results, errors.New(resp.Err)
}

func remoteLs(cmd *LsCmd) error {
	for _, arg := range cmd.Args {
		subcmd := NewLsCmd(
			cmd.R,
			arg,
		)
		result, err := RemoteLs(subcmd)

		fmt.Printf("%s :\n", arg)
		if err != nil {
			fmt.Printf("  error:%s\n\n", err.Error())
			return err
		}

		formatPrint(result)
	}
	return nil
}

// Execute runs a LsCmd.
func (p *LsCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd, err := newLsCmdFromFlag(f)
	if err != nil {
		return subcommands.ExitFailure
	}
	log.V(1).Infof("%#v\n", cmd)

	if err := remoteLs(cmd); err != nil {
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
