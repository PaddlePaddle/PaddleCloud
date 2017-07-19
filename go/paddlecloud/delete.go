package paddlecloud

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
)

// DeleteCommand do job killings
type DeleteCommand struct {
}

// Name is subcommands name
func (*DeleteCommand) Name() string { return "delete" }

// Synopsis is subcommands synopsis
func (*DeleteCommand) Synopsis() string { return "Delete the specify resource." }

// Usage is subcommands usage
func (*DeleteCommand) Usage() string {
	return `delete registry [registry-name]
`
}

// SetFlags registers subcommands flags
func (p *DeleteCommand) SetFlags(f *flag.FlagSet) {
}

// Execute kill command
func (p *DeleteCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 2 {
		f.Usage()
		return subcommands.ExitFailure
	}
	if f.Arg(0) == RegistryCmdName {
		name := f.Arg(1)
		r := RegistryCmd{SecretName: KubeRegistryName(name)}
		err := r.Delete()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error delete registry: %v\n", err)
			return subcommands.ExitFailure
		}
		fmt.Fprintf(os.Stdout, "registry: [%s] is deleted\n", name)
	} else {
		f.Usage()
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
