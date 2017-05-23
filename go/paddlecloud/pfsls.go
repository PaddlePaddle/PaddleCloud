package main

import (
	//"bytes"
	"context"
	//"crypto/tls"
	//"crypto/x509"
	//"encoding/json"
	"flag"
	"fmt"
	pfsmod "github.com/cloud/go/file_manager/pfsmodules"
	//log "github.com/golang/glog"
	"github.com/google/subcommands"
	//"gopkg.in/yaml.v2"
	//"io/ioutil"
	//"net/http"
	//"log"
)

type lsCommand struct {
	r bool
}

func (*lsCommand) Name() string     { return "ls" }
func (*lsCommand) Synopsis() string { return "List files on PaddlePaddle Cloud" }
func (*lsCommand) Usage() string {
	return `ls [-r] <pfspath>:
	List files on PaddlePaddleCloud
	Options:
`
}

func (p *lsCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.r, "r", false, "list files recursively")
}

//func RemoteLs(path string, r bool) (*pfsmod.LsCmdResponse, error) {
func RemoteLs(cmdAttr *pfsmod.CmdAttr) (*pfsmod.LsCmdResponse, error) {
	resp := pfsmod.LsCmdResponse{}
	s := NewCmdSubmitter(UserHomeDir() + "/.paddle/config")

	lsCmd := pfsmod.NewLsCmd(cmdAttr, &resp)
	_, err := s.SubmitCmdReqeust("GET", "api/v1/files", 8080, lsCmd)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return &resp, err
	}

	return &resp, err

}

func (p *lsCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmdAttr := pfsmod.NewCmdAttr(p.Name(), f)

	resp, err := RemoteLs(cmdAttr)
	if err != nil {
		return subcommands.ExitFailure
	}

	fmt.Println(resp)

	return subcommands.ExitSuccess
}
