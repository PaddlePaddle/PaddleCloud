package paddlecloud

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	"github.com/google/subcommands"
	"log"
	"os"
	"path/filepath"
)

type CpCommand struct {
	cmd pfsmod.CpComand
}

func (*CpCommand) Name() string     { return "cp" }
func (*CpCommand) Synopsis() string { return "uoload or download files" }
func (*CpCommand) Usage() string {
	return `cp [-v] <src> <dst>
	upload or downlod files, does't support directories this version
	Options:
	`
}

func (p *CpCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.cmd.V, "v", false, "Cause cp to be verbose, showing files after they are copied.")
}

func (p *CpCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 2 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd := pfsmod.NewCpCommand(f)

	s := NewPfsSubmitter(UserHomeDir() + "/.paddle/config")

	results, err := RunCp(s, cmd)
	if err != nil {
		return subcommands.ExitFailure
	}

	fmt.Println(results)
	return subcommands.ExitSuccess
}

// Run cp command, return err when meet any error
func RunCp(s *NewPfsSubmitter, cmd *pfsmod.CpCommand) ([]pfsmod.CpCommandResult, error) {

	var results []pfmod.CpCommandResult

	for _, arg := range src {
		fmt.Println(cmd.ToString(arg, cmd.Dst))

		var ret []pfmod.CpCommandResult
		var err error

		if pfsmod.IsRemotePath(arg) {
			if pfsmod.IsRemotePath(dst) {
				err := errors.New(pfsmod.StatusText(pfsmod.StatusOnlySupportUploadOrDownloadFiles))
			} else {
				ret, err = Download(s, arg, dst)
			}
		} else {
			if pfsmod.IsRemotePath(dst) {
				ret, err = Upload(s, arg, dst)
			} else {
				//can't do that
				err := errors.New(pfsmod.StatusText(pfsmod.StatusOnlySupportUploadOrDownloadFiles))
			}
		}

		if err != nil {
			fmt.Printf("%v\n", err)
			return err
		}

		if ret != nil {
			results = append(results, ret...)
		}
	}

	return results, nil
}
