package paddlecloud

import (
	"context"
	"errors"
	"flag"
	"fmt"

	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
	"github.com/google/subcommands"
)

const (
	defaultChunkSize = 2 * 1024 * 1024
)

// CpCommand represents a copy command
type CpCommand struct {
	cmd pfsmod.CpCmd
}

// Name returns CpCommand's name
func (*CpCommand) Name() string { return "cp" }

// Synopsis returns synopsis of CpCommand
func (*CpCommand) Synopsis() string { return "uoload or download files" }

// Usage returns usage of CpCommand
func (*CpCommand) Usage() string {
	return `cp [-v] <src> <dst>
	upload or downlod files, does't support directories this version
	Options:
	`
}

// SetFlags sets CpCommand's parameter
func (p *CpCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.cmd.V, "v", false, "Cause cp to be verbose, showing files after they are copied.")
}

// Execute runs CpCommand
func (p *CpCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 2 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd := pfsmod.NewCpCmdFromFlag(f)

	s := newPfsCmdSubmitter(UserHomeDir() + "/.paddle/config")

	err := RunCp(s, cmd)
	if err != nil {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

// RunCp runs CpCommand
func RunCp(s *pfsSubmitter, cmd *pfsmod.CpCmd) error {

	var results []pfsmod.CpCmdResult

	for _, arg := range cmd.Src {
		fmt.Println(cmd.PartToString(arg, cmd.Dst))

		var ret []pfsmod.CpCmdResult
		var err error

		if pfsmod.IsCloudPath(arg) {
			if pfsmod.IsCloudPath(cmd.Dst) {
				err = errors.New(pfsmod.StatusOnlySupportFiles)
			} else {
				err = Download(s, arg, cmd.Dst)
			}
		} else {
			if pfsmod.IsCloudPath(cmd.Dst) {
				err = Upload(s, arg, cmd.Dst)
			} else {
				//can't do that
				err = errors.New(pfsmod.StatusOnlySupportFiles)
			}
		}

		if err != nil {
			fmt.Printf("%#v\n", err)
			return err
		}

		if ret != nil {
			results = append(results, ret...)
		}
		fmt.Println("")
	}

	return nil
}
