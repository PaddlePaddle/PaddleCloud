package pfsmodules

import (
	//"bytes"
	//"context"
	//"crypto/tls"
	//"crypto/x509"
	"encoding/json"
	//"flag"
	//"fmt"
	//pfsmod "github.com/cloud/go/file_manager/pfsmodules"
	//	log "github.com/golang/glog"
	//"github.com/google/subcommands"
	//"gopkg.in/yaml.v2"
	//"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type flagFormal struct {
	r bool
}

type FileMeta struct {
	//Err     string `json:"Err"`
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
	//Cmd string `json:"cmd"`
	//ErrCode int32      `json:"ErrCode"`
	Path  string     `json:"path"`
	Err   string     `json:"err"`
	Metas []FileMeta `json:"metas"`
}

type LsCmdResponse struct {
	//ErrCode int32         `json:"errcode"`
	Err     string        `json:"err"`
	Results []LsCmdResult `json:"result"`
}

func (p *LsCmdResponse) GetErr() string {
	return p.Err
}

func (p *LsCmdResponse) SetErr(err string) {
	p.Err = err
}

/*
func (p *LsCmdResponse) GetErrCode() string {
	return p.ErrCode
}

func (p *LsCmdResponse) SetErr(errCode int32, err string) {
	p.errCode = errcode
	p.Err = err
}
*/

type LsCmd struct {
	cmd  *Cmd
	resp *LsCmdResponse
}

func NewLsCmd(cmd *Cmd, resp *LsCmdResponse) *LsCmd {
	return &LsCmd{
		cmd:  cmd,
		resp: resp,
	}
}

func (p *LsCmd) GetCmd() *Cmd {
	return p.cmd
}

func (p *LsCmd) GetResponse() Response {
	return p.resp
}

func lsPath(path string, r bool) []FileMeta {

	log.Println("path:\t" + path)

	metas := make([]FileMeta, 0, 100)

	filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		//log.Println("path:\t" + path)

		m := FileMeta{}
		m.Path = info.Name()
		m.Size = info.Size()
		m.ModTime = info.ModTime().Format("2006-01-02 15:04:05")
		m.IsDir = info.IsDir()
		metas = append(metas, m)

		if info.IsDir() && !r && subpath != path {
			return filepath.SkipDir
		}

		//log.Println(len(metas))
		return nil
	})

	return metas
}

func (p *LsCmd) Run() {
	r := false
	for _, t := range p.cmd.Options {
		if t.Name == "r" {
			r = true
			break
		}
	}

	results := make([]LsCmdResult, 0, 100)
	for _, arg := range p.cmd.Args {
		log.Printf("ls %s\n", arg)
		m := LsCmdResult{}
		m.Path = arg

		list, err := filepath.Glob(arg)
		if err != nil {
			m.Err = err.Error()
			results = append(results, m)
			log.Printf("glob path:%s error:%s", arg, m.Err)
			continue
		}

		if len(list) == 0 {
			m.Err = "file or directory not exist"
			results = append(results, m)
			log.Printf("glob path:%s error:%s", arg, m.Err)
			continue
		}

		for _, path := range list {
			m.Path = path
			m.Metas = lsPath(path, r)
			results = append(results, m)
		}
	}

	p.resp.Results = results
}

func (p *LsCmd) RunAndResponse(w http.ResponseWriter) error {
	p.Run()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(p.resp); err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		log.Printf("write response error:%v", err)
		return err
	}
	return nil
}
