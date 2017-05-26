package pfsmod

import (
	"encoding/json"
	"errors"
	"log"
)

type CpCommand struct {
	Method string
	V      bool
	Src    []string
	Des    string
}

func (p *CpCommand) ToUrl() string {
}

func (p *CpCommand) ToJson() []byte {
}

func (p *cpCommand) Run() interface{} {
}
