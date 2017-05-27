package paddlecloud

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	"github.com/google/subcommands"
)

type CpCommand struct {
	cmd pfsmod.CpCmd
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

	cmd := pfsmod.NewCpCmdFromFlag(f)

	s := NewPfsCmdSubmitter(UserHomeDir() + "/.paddle/config")

	results, err := RunCp(s, cmd)
	if err != nil {
		return subcommands.ExitFailure
	}

	fmt.Println(results)
	return subcommands.ExitSuccess
}

// Run cp command, return err when meet any error
func RunCp(s *PfsSubmitter, cmd *pfsmod.CpCmd) ([]pfsmod.CpCmdResult, error) {

	var results []pfsmod.CpCmdResult

	for _, arg := range cmd.Src {
		fmt.Println(cmd.PartToString(arg, cmd.Dst))

		var ret []pfsmod.CpCmdResult
		var err error

		if pfsmod.IsCloudPath(arg) {
			if pfsmod.IsCloudPath(cmd.Dst) {
				err := errors.New(pfsmod.StatusText(pfsmod.StatusOnlySupportFiles))
			} else {
				ret, err = Download(s, arg, cmd.Dst)
			}
		} else {
			if pfsmod.IsCloudPath(cmd.Dst) {
				ret, err = Upload(s, arg, cmd.Dst)
			} else {
				//can't do that
				err := errors.New(pfsmod.StatusText(pfsmod.StatusOnlySupportFiles))
			}
		}

		if err != nil {
			fmt.Printf("%v\n", err)
			return results, err
		}

		if ret != nil {
			results = append(results, ret...)
		}
	}

	return results, nil
}
