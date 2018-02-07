package main

import (
	"context"
	"flag"
	"os"

	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
	"github.com/PaddlePaddle/cloud/go/paddlecloud"
	"github.com/PaddlePaddle/cloud/go/paddlectl"
	"github.com/PaddlePaddle/cloud/go/utils/config"
	"github.com/google/subcommands"
)

func main() {
	pfsmod.Config = config.ParseDefaultConfig()
	paddlecloud.Config = pfsmod.Config

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&paddlectl.SubmitCmd{}, "")

	// TODO(gongwb): add more commands.
	subcommands.Register(&paddlectl.SimpleFileCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
