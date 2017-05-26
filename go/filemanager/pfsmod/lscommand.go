package pfsmod

import (
	"encoding/json"
	"errors"
	"log"
)

type FileMeta struct {
	//Path    string `json:"Path"`
	ModTime string `json:"ModTime"`
	Size    int64  `json:"Size"`
	IsDir   bool   `json:IsDir`
}

type FileAttr struct {
	Path       string   `json:"Path"`
	StatusCode int32    `json:"StatusCode"`
	StatusText string   `json:"StatusText"`
	Meta       FileMeta `json:"Meta"`
}

type LsCommandResult struct {
	StatusCode int32      `json:"StatusCode"`
	StatusText string     `json:"StatusText"`
	Attr       []FileAttr `json:"Attr"`
}

type LsCommand struct {
	Method string
	R      bool
	Args   []string
}

func (p *LsCommand) ToUrl() string {
}

func (p *LsCommand) ToJson() []byte {
}

func (p *LsCommand) Run() interface{} {
}

func NewLsCommand(f *flag.FlagSet) (*LsCommand, error) {
	cmd = LsCommand{}

	cmd.Method = "ls"
	cmd.Args = make([]string, 0, f.NArg())

	f.Visit(func(flag *flag.Flag) {
		if flag.Name == "r" {
			cmd.R = flag.Value.(bool)
		}
	})

	for _, arg := range f.Args() {
		cmd.Args = append(cmd.Args, arg)
	}

	return &cmd, nil
}
