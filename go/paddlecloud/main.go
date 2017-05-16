package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
)

type logsCommand struct {
	n int
}

func (*logsCommand) Name() string     { return "logs" }
func (*logsCommand) Synopsis() string { return "Print logs of the job." }
func (*logsCommand) Usage() string {
	return `logs <job name>:
	Print logs of the job.
	Options:
`
}

func (p *logsCommand) SetFlags(f *flag.FlagSet) {
	f.IntVar(&p.n, "n", 10, "Number of lines to print from tail.")
}

func (p *logsCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	fmt.Println("Printing logs...")
	return subcommands.ExitSuccess
}

type getCommand struct {
	a bool
}

func (*getCommand) Name() string     { return "get" }
func (*getCommand) Synopsis() string { return "Print resources" }
func (*getCommand) Usage() string {
	return `get:
	Print resources.
`
}

func (p *getCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.a, "a", false, "Get all resources.")
}

func (p *getCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	for _, arg := range f.Args() {
		fmt.Printf("Getting resource info %s ", arg)
	}
	fmt.Println()
	return subcommands.ExitSuccess
}

type killCommand struct {
	rm bool
}

func (*killCommand) Name() string     { return "kill" }
func (*killCommand) Synopsis() string { return "Stop the job. -rm will remove the job from history." }
func (*killCommand) Usage() string {
	return `kill:
	Stop the job. -rm will remove the job from history.
`
}

func (p *killCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.rm, "rm", false, "remove the job from history")
}

func (p *killCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	fmt.Println("Killing job...")
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&submitCmd{}, "")
	subcommands.Register(&logsCommand{}, "")
	subcommands.Register(&getCommand{}, "")
	subcommands.Register(&killCommand{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
