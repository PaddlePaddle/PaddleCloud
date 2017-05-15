package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
)

type submitCmd struct {
	jobpackage string
	parallism  int
	cpu        int
	gpu        int
	memory     string
	pservers   int
	pscpu      int
	psmemory   string
	entry      string
	topology   string
}

func (*submitCmd) Name() string     { return "submit" }
func (*submitCmd) Synopsis() string { return "Submit job to PaddlePaddle Cloud." }
func (*submitCmd) Usage() string {
	return `submit [options] <package path>:
	Submit job to PaddlePaddle Cloud.
	Options:
`
}

func (p *submitCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&p.parallism, "parallism", 1, "Number of parrallel trainers. Defaults to 1.")
	f.IntVar(&p.cpu, "cpu", 1, "CPU resource each trainer will use. Defaults to 1.")
	f.IntVar(&p.gpu, "gpu", 0, "GPU resource each trainer will use. Defaults to 0.")
	f.StringVar(&p.memory, "memory", "1Gi", " Memory resource each trainer will use. Defaults to 1Gi.")
	f.IntVar(&p.pservers, "pservers", 0, "Number of parameter servers. Defaults equal to -p")
	f.IntVar(&p.pscpu, "pscpu", 1, "Parameter server CPU resource. Defaults to 1.")
	f.StringVar(&p.psmemory, "psmemory", "1Gi", "Parameter server momory resource. Defaults to 1Gi.")
	f.StringVar(&p.entry, "entry", "paddle train", "Command of starting trainer process. Defaults to paddle train")
	f.StringVar(&p.topology, "topology", "", "Will Be Deprecated .py file contains paddle v1 job configs")
}

func (p *submitCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	for _, arg := range f.Args() {
		if p.pservers == 0 {
			p.pservers = p.parallism
		}
		fmt.Printf("%s ", arg)
	}
	fmt.Println()
	return subcommands.ExitSuccess
}

type jobsCommand struct {
	a bool
}

func (*jobsCommand) Name() string     { return "jobs" }
func (*jobsCommand) Synopsis() string { return "List jobs. List only running jobs if no -a specified." }
func (*jobsCommand) Usage() string {
	return `jobs [-a]:
	List jobs. List only running jobs if no -a specified.
	Options:
`
}

func (p *jobsCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.a, "a", false, "List all jobs.")
}

func (p *jobsCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	fmt.Println("Listing jobs...")
	return subcommands.ExitSuccess
}

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

type quotaCommand struct {
}

func (*quotaCommand) Name() string     { return "quota" }
func (*quotaCommand) Synopsis() string { return "Show user's quota usages." }
func (*quotaCommand) Usage() string {
	return `quota:
	Show user's quota usages.
`
}

func (p *quotaCommand) SetFlags(f *flag.FlagSet) {
}

func (p *quotaCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	fmt.Println("Printing quota...")
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&submitCmd{}, "")
	subcommands.Register(&jobsCommand{}, "")
	subcommands.Register(&logsCommand{}, "")
	subcommands.Register(&quotaCommand{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
