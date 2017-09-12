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
	subcommands.Register(&paddlecloud.RegistryCmd{}, "")
	subcommands.Register(&paddlecloud.DeleteCommand{}, "")
	subcommands.Register(&paddlecloud.PublishCmd{}, "")
	subcommands.Register(&pfsmod.LsCmd{}, "PFS")
	subcommands.Register(&pfsmod.CpCmd{}, "PFS")
	subcommands.Register(&pfsmod.RmCmd{}, "PFS")
	subcommands.Register(&pfsmod.MkdirCmd{}, "PFS")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
