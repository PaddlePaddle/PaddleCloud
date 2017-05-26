package pfsmod

import (
	"encoding/json"
	"errors"
	"log"
)

type CpCmdResult struct {
	Src string `json:"Path"`
	Dst string `json:"Dst"`
}

type CpCmd struct {
	Method string
	V      bool
	Src    []string
	Dst    string
}

func (p *CpCmd) ToUrlParam() string {
	return ""
}

func (p *CpCmd) ToJson() ([]byte, error) {
	return nil, nil
}

func (p *cpCmd) Run() (interface{}, error) {
	return nil, nil
}

func NewCpCmdFromFlag(f *flag.FlagSet) (*CpCmd, error) {
	cmd := CpCmd{}

	cmd.Method = "cp"
	cmd.Src = make([]string, 0, f.NArg())

	f.Visit(func(flag *flag.Flag) {
		if flag.Name == "v" {
			cmd.v = flag.Value.(bool)
		}
	})

	for _, arg := range f.Args() {
		cmd.Args = append(cmd.Args, arg)
	}

	for i < len(f.Args())-1 {
	}

	return &cmd, nil
}

func (p *CpCmd) PartToString(src, dst string) string {
	if p.V {
		return fmt.Printf("cp -v %s %s\n", src, dst)
	}
	return fmt.Printf("cp %s %s\n", src, dst)
}
