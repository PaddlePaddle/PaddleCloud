package paddlecloud

import (
	"context"
	"flag"
	"fmt"
	"net/url"

	"github.com/PaddlePaddle/cloud/go/utils/restclient"
	"github.com/google/subcommands"
)

// PublishCmd used for publish file for download and list published files.
type PublishCmd struct {
}

// Name is subcommands name.
func (*PublishCmd) Name() string { return "publish" }

// Synopsis is subcommands synopsis.
func (*PublishCmd) Synopsis() string {
	return "publish file for download and list published files."
}

// Usage is subcommands Usage.
func (*PublishCmd) Usage() string {
	return `publish [path]
	path must be like /pfs/[datacenter]/home/[username]
	if path not specified, will return a list of current published files.
`
}

// SetFlags registers subcommands flags.
func (p *PublishCmd) SetFlags(f *flag.FlagSet) {
}

// Execute publish ops.
func (p *PublishCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() > 1 {
		f.Usage()
		return subcommands.ExitFailure
	}
	if f.NArg() == 0 {
		queries := url.Values{}
		ret, err := restclient.GetCall(Config.ActiveConfig.Endpoint+"/api/v1/publish/", queries)
		if err != nil {
			return subcommands.ExitFailure
		}
		fmt.Printf("%s\n", ret)
	} else if f.NArg() == 1 {
		queries := url.Values{}
		queries.Set("path", f.Arg(0))
		restclient.PostCall(Config.ActiveConfig.Endpoint+"/api/v1/publish/", []byte("{\"path\": \""+f.Arg(0)+"\"}"))
	}

	return subcommands.ExitSuccess
}
