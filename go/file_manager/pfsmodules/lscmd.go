package pfsmodules

import (
	//"bytes"
	//"context"
	//"crypto/tls"
	//"crypto/x509"
	//"encoding/json"
	//"flag"
	//"fmt"
	//pfsmod "github.com/cloud/go/file_manager/pfsmodules"
	//	log "github.com/golang/glog"
	//"github.com/google/subcommands"
	//"gopkg.in/yaml.v2"
	//"io/ioutil"
	//"net/http"
	"log"
	"os"
	"path/filepath"
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

func (p *LsCmd) GetCmd() *Cmd {
	return p.cmd
}

func (p *LsCmd) GetResponse() Response {
	return p.resp
}

func (p *LsCmd) Run() {
}

func LsPath(path string, r bool) ([]FileMeta, error) {

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
