package main

import (
	"context"
	"flag"
	"os"

	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
	"github.com/PaddlePaddle/cloud/go/paddlecloud"
	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&paddlecloud.SubmitCmd{}, "")
	subcommands.Register(&paddlecloud.LogsCommand{}, "")
	subcommands.Register(&paddlecloud.GetCommand{}, "")
	subcommands.Register(&paddlecloud.KillCommand{}, "")
	subcommands.Register(&paddlecloud.SimpleFileCmd{}, "")
	subcommands.Register(&pfsmod.LsCmd{}, "")
	subcommands.Register(&pfsmod.CpCmd{}, "")
	subcommands.Register(&pfsmod.RmCmd{}, "")
	subcommands.Register(&pfsmod.MkdirCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
