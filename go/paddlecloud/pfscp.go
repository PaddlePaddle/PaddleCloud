package main

import (
	"context"
	"flag"
	//"fmt"
	//"errors"
	"fmt"
	"github.com/cloud/go/file_manager/pfsmodules"
	"github.com/google/subcommands"
	"log"
)

type cpCommand struct {
	v bool
}

func (*cpCommand) Name() string     { return "cp" }
func (*cpCommand) Synopsis() string { return "copy files or directories" }
func (*cpCommand) Usage() string {
	return `cp [-v] <src> <dest>
	copy files or directories
	Options:
	`
}

func (p *cpCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.v, "v", false, "Cause cp to be verbose, showing files after they are copied.")
}

func (p *cpCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 2 {
		f.Usage()
		return subcommands.ExitFailure
	}

	attrs := pfsmodules.NewCpCmdAttr("cp", f)
	/*
		if err != nil {
			return subcommands.ExitFailure
		}
	*/

	results, err := RunCp(attrs)
	if err != nil {
		return subcommands.ExitFailure
	}

	log.Println(results)

	return subcommands.ExitSuccess
}

func RunCp(p *pfsmodules.CpCmdAttr) ([]pfsmodules.CpCmdResult, error) {
	src, err := p.GetSrc()
	if err != nil {
		return nil, err
	}

	dest, err := p.GetDest()
	if err != nil {
		return nil, err
	}

	//results := make([]CpCmdResult, 0, 100)

	var results []pfsmodules.CpCmdResult

	for _, arg := range src {
		log.Printf("ls %s\n", arg)

		if pfsmodules.IsRemotePath(arg) {
			if pfsmodules.IsRemotePath(dest) {
				//remotecp
			} else {
				//download
			}
		} else {
			if pfsmodules.IsRemotePath(dest) {
				//upload
			} else {
				//can't do that
			}
		}

		//m := CopyGlobPath(arg, dest)
		//results = append(results, m)
	}

	return results, nil
}

func RemoteCp(cpCmdAttr *pfsmodules.CpCmdAttr, src, dest string) ([]pfsmodules.CpCmdResult, error) {
	resp := pfsmodules.RemoteCpCmdResponse{}
	cmdAttr := pfsmodules.CmdAttr{
		Method:  cpCmdAttr.Method,
		Options: cpCmdAttr.Options,
		Args:    cpCmdAttr.Args,
	}

	s := NewCmdSubmitter(UserHomeDir() + "/.paddle/config")

	remoteCpCmd := pfsmodules.NewRemoteCpCmd(&cmdAttr, &resp)
	_, err := s.SubmitCmdReqeust("POST", "api/v1/files", 8080, remoteCpCmd)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}

	return resp.Results, nil
}

/*
//local ... to remote
//remote ... to local
//remote ... to remote
func check(attr *CpCmdAttr) error {
	if len(attr.Src) < 1 || len(attr.Dest) < 1 {
		return errors.New("too few aruments")
	}

	//if len(attr.Args)
}

func expand(src *CpCmdAttr) (*CpCmdAttr, error) {
}

func expandClientPath(path string) ([]string, error) {
	return nil, nil
}

func expandRemotePath(path string) ([]string, error) {
	return nil, nil
}

func expandPath(attr pfsmodules.CpCmdAttr) ([]pfsmodules.CpCmdAttr, error) {
	return nil, nil
}

func cp(src, dest string) {
}

func upload(attr *pfsmodules.CpCmdAttr) error {
	return nil
}

func download(attr *pfsmodules.CpCmdAttr) error {
	return nil
}

func remoteCp(attr *pfsmodules.CpCmdAttr) error {
	return nil
}



func RunOne(attr pfsmodules.CpCmdAttr) error {
	var err error
	if pfsmodules.IsRemotePath(attr.Src) {
		if pfsmodules.IsRemotePath(attr.Dest) {
			err = remoteCp(attr)
		} else {
			err = download(attr)
		}
		return err
	}

	if !pfsmodules.IsRemotePath(attr.Dest) {
		return errors.New("can't cp local to local")
	}

	err = upload(attr)
	return err
}
*/
