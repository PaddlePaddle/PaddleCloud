package pfsmod

import (
	"encoding/json"
	"errors"
	"log"
)

type Command interface {
	ToUrl() string
	ToJson() []byte
	Run() interface{}
}
