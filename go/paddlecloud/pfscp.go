package main

import (
	"context"
	"flag"
	//"fmt"
	"errors"
	"github.com/cloud/go/file_manager/pfsmodules"
	"github.com/google/subcommands"
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

	attrs, err := NewCpCmdAttr()
	if err != nil {
		return subcommands.ExitFailure
	}

	err = RunCp(attrs)
	if err != nil {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
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
*/

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

func NewCpCmdAttr() ([]pfsmodules.CpCmdAttr, error) {
	//attr := make([]pfsmodules.CpCmdAttr, 0, 100)
	//expandPath(attr)
	return nil, nil
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

func RunCp(attrs []pfsmodules.CpCmdAttr) error {
	for _, attr := range attrs {
		err := RunOne(attr)
		if err != nil {
			//ToDo
		}
		continue
	}

	return nil
}
