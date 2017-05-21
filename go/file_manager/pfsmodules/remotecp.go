package pfsmodules

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type RemoteCpCmdResponse struct {
	Err     string        `json:"err"`
	Results []CpCmdResult `json:"result"`
}

func (p *RemoteCpCmdResponse) GetErr() string {
	return p.Err
}

func (p *RemoteCpCmdResponse) SetErr(err string) {
	p.Err = err
}

type RemoteCpCmd struct {
	cmdAttr *CmdAttr
	resp    *RemoteCpCmdResponse
}

func NewRemoteCpCmd(cmdAttr *CmdAttr, resp *RemoteCpCmdResponse) *RemoteCpCmd {
	return &RemoteCpCmd{
		cmdAttr: cmdAttr,
		resp:    resp,
	}
}

func (p *RemoteCpCmd) GetCmdAttr() *CmdAttr {
	return p.cmdAttr
}

func (p *RemoteCpCmd) GetResponse() Response {
	return p.resp
}

func (p *RemoteCpCmd) Run() {
	//log.Println(p.cmd.Args)
	results := make([]RemoteCpCmdResult, 0, 100)

	src, err := p.getSrc()
	if err != nil {
		return err
	}

	dest, err := p.getDest()
	if err != nil {
		return err
	}

	for _, arg := range src {
		log.Printf("ls %s\n", arg)
		m := RemoteCpCmdResult{}
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
			m.Metas = lsPath(path, r)
			results = append(results, m)
		}
	}

	p.resp.Results = results
}

func (p *RemoteCpCmd) RunAndResponse(w http.ResponseWriter) error {
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
