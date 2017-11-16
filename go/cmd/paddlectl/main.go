package main

import (
	"context"
	"flag"
	"os"

	"github.com/PaddlePaddle/cloud/go/paddlectl"
	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&paddlectl.SubmitCmd{}, "")

	// TODO(gongwb): add more commands.
	subcommands.Register(&paddlectl.SimpleFileCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
