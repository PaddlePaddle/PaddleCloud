package pfsmodules

import (
	"flag"
	"fmt"
	"strconv"

	log "github.com/golang/glog"
)

const (
	cpCmdName = "cp"
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

// NewCpCmdFromFlag returns a new CpCmd from parsed flags.
func NewCpCmdFromFlag(f *flag.FlagSet) (*CpCmd, error) {
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
	if p.V {
		return fmt.Sprintf("cp -v %s %s\n", src, dst)
	}
	return fmt.Sprintf("cp %s %s\n", src, dst)
}
