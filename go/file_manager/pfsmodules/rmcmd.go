package pfsmodules

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type RmCmdResult struct {
	Err  string `json:"err"`
	Path string `json:"path"`
}

type RmCmdResponse struct {
	Err     string        `json:"err"`
	Results []RmCmdResult `json:"path"`
}

func (p *RmCmdResponse) SetErr(err string) {
	p.Err = err
}

func (p *RmCmdResponse) GetErr() string {
	return p.Err
}

type RmCmd struct {
	cmdAttr *CmdAttr
	resp    *RmCmdResponse
}

func NewRmCmd(cmdAttr *CmdAttr, resp *RmCmdResponse) *RmCmd {
	return &RmCmd{
		cmdAttr: cmdAttr,
		resp:    resp,
	}
}

func (p *RmCmd) GetCmdAttr() *CmdAttr {
	return p.cmdAttr
}

func (p *RmCmd) GetResponse() Response {
	return p.resp
}

func (p *RmCmd) Run() {
	r := false
	for _, t := range p.cmdAttr.Options {
		if t.Name == "r" {
			r = true
			break
		}
	}

	//log.Println(p.cmd.Args)
	results := make([]RmCmdResult, 0, 100)
	for _, arg := range p.cmdAttr.Args {
		log.Printf("ls %s\n", arg)
		m := RmCmdResult{}
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
			m.Path = path

			fi, _ := os.Stat(path)
			if fi.IsDir() && !r {
				m.Err = "add -r to remove a directory"
				results = append(results, m)
				log.Printf("%s %s", arg, m.Err)
				continue
			}

			if err := os.RemoveAll(path); err != nil {
				m.Err = err.Error()
				results = append(results, m)
				log.Printf("rm path:%s", arg, m.Err)
				continue
			}
			results = append(results, m)
		}
	}

	p.resp.Results = results
}

func (p *RmCmd) RunAndResponse(w http.ResponseWriter) error {
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
