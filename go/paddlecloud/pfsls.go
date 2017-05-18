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

func (p *lsCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd := pfsmod.NewCmd(p.Name(), f)
	resp := pfsmod.LsCmdResponse{}
	s := NewCmdSubmitter(UserHomeDir() + "/.paddle/config")

	/*
		lsCmd := &{
			cmd :
		}
			if err != nil {
				fmt.Printf("error NewPfsCommand: %v\n", err)
				return subcommands.ExitFailure
			}
	*/

	lsCmd := pfsmod.NewLsCmd(cmd, &resp)
	_, err := s.SubmitCmdReqeust("GET", lsCmd)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return subcommands.ExitFailure
	}

	//fmt.Println(body)
	return subcommands.ExitSuccess
}
