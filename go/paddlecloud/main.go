package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	"log"
)

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
