package pfsmodules

import (
	"encoding/json"
	//errors"
	//"io"
	"log"
	"net/http"
	//"os"
	//"path/filepath"
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

func (p *RemoteCpCmd) Run() error {
	cpCmdAttr := CpCmdAttr{
		Method:  p.cmdAttr.Method,
		Options: p.cmdAttr.Options,
		Args:    p.cmdAttr.Args,
	}

	srcs, err := cpCmdAttr.GetSrc()
	if err != nil {
		return err
	}

	dest, err := cpCmdAttr.GetDest()
	if err != nil {
		return nil
	}

	results := make([]CpCmdResult, 0, 100)
	for _, src := range srcs {
		m, err := CopyGlobPath(src, dest)
		results = append(results, m...)

		if err != nil {
			return err
		}
	}

	p.resp.Results = results
	return nil
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
