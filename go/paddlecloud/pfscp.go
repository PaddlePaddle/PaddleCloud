package paddlecloud

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
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
	return `cp [-v] <src> <dest>
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

	attrs := pfsmodules.NewCpCmdAttr("cp", f)

	results, err := RunCp(attrs)
	if err != nil {
		return subcommands.ExitFailure
	}

	log.Println(results)

	return subcommands.ExitSuccess
}
