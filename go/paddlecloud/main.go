package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"log"
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

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&submitCmd{}, "")
	subcommands.Register(&logsCommand{}, "")
	subcommands.Register(&getCommand{}, "")
	subcommands.Register(&killCommand{}, "")
	subcommands.Register(&lsCommand{}, "")
	subcommands.Register(&rmCommand{}, "")
	subcommands.Register(&cpCommand{}, "")

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
