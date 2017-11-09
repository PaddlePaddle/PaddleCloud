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

	// TODO(gongwb): add these commands.
	// subcommands.Register(&paddlecloud.LogsCommand{}, "")
	// subcommands.Register(&paddlecloud.GetCommand{}, "")
	// subcommands.Register(&paddlecloud.KillCommand{}, "")
	// subcommands.Register(&paddlecloud.SimpleFileCmd{}, "")
	// subcommands.Register(&paddlecloud.RegistryCmd{}, "")
	// subcommands.Register(&paddlecloud.DeleteCommand{}, "")
	// subcommands.Register(&paddlecloud.PublishCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
