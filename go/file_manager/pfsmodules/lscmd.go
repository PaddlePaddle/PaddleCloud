package pfsmodules

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	//pfsmod "github.com/cloud/go/file_manager/pfsmodules"
	log "github.com/golang/glog"
	"github.com/google/subcommands"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

type flagFormal struct {
	r bool
}

type FileMeta struct {
	Err     string `json:"Err"`
	Path    string `json:"Path"`
	ModTime string `json:"ModTime"`
	Size    int64  `json:"Size"`
	IsDir   bool   `json:IsDir`
}

/*
type LsCmdResponse struct {
	ErrCode      int32      `json:"ErrCode"`
	Err          string     `json:"Err"`
	Metas        []FileMeta `json:"Metas"`
	TotalObjects int        `json:"TotalObjects"`
	TotalSize    int64      `json:"TotalSize"`
}
*/

type LsCmdResult struct {
	Cmd     string     `json:"cmd"`
	ErrCode int32      `json:"ErrCode"`
	Err     string     `json:"err"`
	Metas   []FileMeta `json:"metas"`
}

type LsCmdResponse struct {
	ErrCode int32         `json:"errcode"`
	Err     string        `json:"err"`
	Result  []LsCmdResult `json:"result"`
}

func (p *LsCmdResponse) GetErr() (int32, string) {
	return ErrCode, Err
}

func (p *LsCmdResponse) SetErr(errCode int32, err string) {
	p.errCode = errcode
	p.Err = err
}

type LsCmd struct {
	cmd    *Cmd
	resp   *LsCmdResponse
	formal flagFormal
}

func (p *LsCmd) GetCmd() *Cmd {
	return p.Cmd
}

func (p *LsCmd) GetResponse() *Response {
	return p.resp
}

func (*LsCmd) Name() string     { return "ls" }
func (*LsCmd) Synopsis() string { return "List files on PaddlePaddle Cloud" }
func (*LsCmd) Usage() string {
	return `ls [-r] <pfspath>:
	List files on PaddlePaddleCloud
	Options:
`
}

func (p *LsCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.formal.r, "r", false, "list files recursively")
}

func (p *LsCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd := NewCmd(p.Name(), f)

	s := NewCmdSubmitter(UserHomeDir() + "/.paddle/config")
	/*
		if err != nil {
			fmt.Printf("error NewPfsCommand: %v\n", err)
			return subcommands.ExitFailure
		}
	*/

	body, err := s.SubmitCmdReqeust(*cmd, "GET")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return subcommands.ExitFailure
	}

	fmt.Println(body)

	return subcommands.ExitSuccess
}

func (p *LsCmd) Run() {
}

func lsPath(path string, r bool) ([]FileMeta, error) {

	metas := make([]FileMeta, 0, 100)

	list, err := filepath.Glob(path)
	if err != nil {
		m := FileMeta{}
		m.Err = err.Error()
		metas = append(metas, m)

		log.Printf("glob path:%s error:%s", path, m.Err)
		return metas, err
	}

	if len(list) == 0 {
		m := FileMeta{}
		m.Err = "file or directory not exist"
		metas = append(metas, m)

		log.Printf("glob path:%s error:%s", path, m.Err)
		return metas, err
	}

	for _, v := range list {
		log.Println("v:\t" + v)
		filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
			//log.Println("path:\t" + path)

			m := FileMeta{}
			m.Path = info.Name()
			m.Size = info.Size()
			m.ModTime = info.ModTime().Format("2006-01-02 15:04:05")
			m.IsDir = info.IsDir()
			metas = append(metas, m)

			if info.IsDir() && !r && v != path {
				return filepath.SkipDir
			}

			//log.Println(len(metas))
			return nil
		})

	}

	return metas, nil
}

func LsPaths(paths []string, r bool) ([]FileMeta, error) {

	metas := make([]FileMeta, 0, 100)

	for _, path := range paths {
		m, err := LsPath(path, r)

		if err != nil {
			metas = append(metas, m...)
			continue
		}

		metas = append(metas, m...)
	}

	return metas, nil
}
