package pfsmod

import (
	"encoding/json"
	"errors"
	"log"
)

type CpCommandResult struct {
	StatusCode int32  `json:"StatusCode"`
	StatusText string `json:"StatusText"`
	Src        string `json:"Path"`
	Dst        string `json:"Dst"`
}

type CpCommand struct {
	Method string
	V      bool
	Src    []string
	Dst    string
}

func (p *CpCommand) ToUrl() string {
}

func (p *CpCommand) ToJson() []byte {
}

func (p *cpCommand) Run() interface{} {
}

func NewCpCommand(f *flag.FlagSet) (*CpCommand, error) {
	cmd := CpCommand{}

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

func (p *CpCommand) ToString(src, dst string) string {
	if p.V {
		return fmt.Printf("cp -v %s %s\n", src, dst)
	}
	return fmt.Printf("cp %s %s\n", src, dst)
}
