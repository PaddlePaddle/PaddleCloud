package pfsmodules

import (
	//"bytes"
	//"context"
	//"crypto/tls"
	//"crypto/x509"
	"encoding/hex"
	"encoding/json"
	//"flag"
	//"fmt"
	//pfsmod "github.com/cloud/go/file_manager/pfsmodules"
	//	log "github.com/golang/glog"
	//"github.com/google/subcommands"
	//"gopkg.in/yaml.v2"
	//"io/ioutil"
	"crypto/md5"
	"io"
	"log"
	"net/http"
	//"math"
	"os"
	"path/filepath"
)

const (
	defaultChunkLen = 4096
)

type MD5SumResult struct {
	//Cmd    string `json:"cmd"`
	Err    string `json:"err"`
	Path   string `json:"path"`
	MD5Sum string `json:"md5sum"`
}

type MD5SumResponse struct {
	Err     string         `json:"err"`
	Results []MD5SumResult `json:"result"`
}

func (p *MD5SumResponse) GetErr() string {
	return p.Err
}

func (p *MD5SumResponse) SetErr(err string) {
	p.Err = err
}

type MD5SumCmd struct {
	cmdAttr *CmdAttr
	resp    *MD5SumResponse
}

func (p *MD5SumCmd) Name() string {
	return "MD5Sum"
}

func NewMD5SumCmd(cmdAttr *CmdAttr, resp *MD5SumResponse) *MD5SumCmd {
	return &MD5SumCmd{
		cmdAttr: cmdAttr,
		resp:    resp,
	}
}

func (p *MD5SumCmd) GetMD5Sum(path string) (*MD5SumResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	data := make([]byte, defaultChunkLen)
	hash := md5.New()
	for {
		n, err := f.Read(data)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if err == io.EOF {
			break
		}

		if _, err := hash.Write(data[:n]); err != nil {
			return nil, err
		}
	}

	result := MD5SumResult{}
	result.Path = path
	result.MD5Sum = hex.EncodeToString(hash.Sum(nil))

	log.Printf("%s MD5Sum:%s\n", path, result.MD5Sum)

	return &result, err
}

func (p *MD5SumCmd) Run() {

	results := make([]MD5SumResult, 0, 100)

	for _, arg := range p.cmdAttr.Args {
		log.Printf("%s %s\n", p.cmdAttr.Name(), arg)

		m := MD5SumResult{}
		m.Path = arg

		list, err := filepath.Glob(arg)
		if err != nil {
			m.Err = err.Error()
			results = append(results, m)
			log.Printf("glob path:%s error:%s", arg, m.Err)
			continue
		}

		if len(list) == 0 {
			m.Err = FileNotFound
			results = append(results, m)
			log.Printf("glob path:%s error:%s", arg, m.Err)
			continue
		}

		for _, path := range list {
			fi, _ := os.Lstat(path)
			if fi.IsDir() {
				m.Err = DirectoryNotAFile
				m.Path = path
				results = append(results, m)
				log.Printf("path:%s error:%s", path, m.Err)
				continue
			}

			log.Printf("%s %s\n", p.cmdAttr.Name(), path)
			m, err := p.GetMD5Sum(path)
			if err != nil {
				results = append(results, *m)
				continue
			}
			results = append(results, *m)
		}
	}

	p.resp.Results = results
}
func (p *MD5SumCmd) RunAndResponse(w http.ResponseWriter) error {
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
