package pfsmodules

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"

	log "github.com/golang/glog"
	"github.com/google/subcommands"
)

const (
	cpCmdName              = "cp"
	defaultChunkSize int64 = 2 * 1024 * 1024
)

// CpCmdResult means the copy-command's result.
type CpCmdResult struct {
	Src string `json:"Path"`
	Dst string `json:"Dst"`
}

// CpCmd means copy-command.
type CpCmd struct {
	Method string
	V      bool
	Src    []string
	Dst    string
}

func newCpCmdFromFlag(f *flag.FlagSet) (*CpCmd, error) {
	cmd := CpCmd{}

	cmd.Method = cpCmdName
	cmd.Src = make([]string, 0, f.NArg())

	var err error
	f.Visit(func(flag *flag.Flag) {
		if flag.Name == "v" {
			cmd.V, err = strconv.ParseBool(flag.Value.String())
			if err != nil {
				log.Errorln("meets error when parsing argument v")
				return
			}
		}
	})

	if err != nil {
		return nil, err
	}

	for i, arg := range f.Args() {
		if i >= len(f.Args())-1 {
			break
		}
		cmd.Src = append(cmd.Src, arg)
	}

	cmd.Dst = f.Args()[len(f.Args())-1]

	return &cmd, nil
}

// PartToString prints command's info.
func (p *CpCmd) PartToString(src, dst string) string {
	return fmt.Sprintf("cp %s %s", src, dst)
}

// Name returns CpCmd's name.
func (*CpCmd) Name() string { return "cp" }

// Synopsis returns synopsis of CpCmd.
func (*CpCmd) Synopsis() string { return "upload or download files" }

// Usage returns usage of CpCmd.
func (*CpCmd) Usage() string {
	return `cp [-v] <src> <dst>
	upload or downlod files, does't support directories this version
	Options:
	`
}

// SetFlags sets CpCmd's parameter.
func (p *CpCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.V, "v", false, "Cause cp to be verbose, showing files after they are copied.")
}

// Execute runs CpCmd.
func (p *CpCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 2 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd, err := newCpCmdFromFlag(f)
	if err != nil {
		return subcommands.ExitSuccess
	}

	if err := RunCp(cmd); err != nil {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

// RunCp runs CpCmd.
func RunCp(cmd *CpCmd) error {
	var results []CpCmdResult

	for _, arg := range cmd.Src {
		var ret []CpCmdResult
		var err error

		if IsCloudPath(arg) {
			if IsCloudPath(cmd.Dst) {
				err = errors.New(StatusOnlySupportFiles)
			} else {
				err = download(arg, cmd.Dst)
			}
		} else {
			if IsCloudPath(cmd.Dst) {
				err = upload(arg, cmd.Dst)
			} else {
				//can't do that
				err = errors.New(StatusOnlySupportFiles)
			}
		}

		if err != nil {
			return err
		}

		if ret != nil {
			results = append(results, ret...)
		}
	}

	return nil
}
