package pfsmodules

import (
	"flag"
	"reflect"
	"strings"
	"testing"
)

func TestNewLsCmdFromURLParam(t *testing.T) {
	s := LsCmd{
		Method: "ls",
		R:      false,
		Args:   []string{"/pfs/test1/", "/pfs/test2/"},
	}

	path := "arg=%2Fpfs%2Ftest1%2F&arg=%2Fpfs%2Ftest2%2F&method=ls&r=false"

	d, err := NewLsCmdFromURLParam(path)
	if err != nil {
		t.Error(err.Error())
	}

	if s.Method != d.Method {
		t.Error(d.Method)
	}

	if s.R != d.R {
		t.Error(d.R)
	}

	eq := reflect.DeepEqual(s.Args, d.Args)
	if !eq {
		t.Error(s.Args)
		t.Error(d.Args)
	}
}

func TestNewLsCmdFromFlag(t *testing.T) {
	cmdLine := "ls -r /pfs/datacenter/home/user1/1.txt /pfs/datacenter/home/user1/"
	a := strings.Split(cmdLine, " ")

	flag := flag.NewFlagSet("ls", flag.ExitOnError)
	flag.Bool("r", false, "")

	if err := flag.Parse(a[1:]); err != nil {
		t.Error(err.Error())
	}

	d, err := newLsCmdFromFlag(flag)
	if err != nil {
		t.Error(err.Error())
	}

	if d.Method != "ls" {
		t.Error(d.Method)
	}

	if !d.R {
		t.Error(d.R)
	}

	if len(d.Args) != 2 {
		t.Error(len(d.Args))
	}

	for _, s := range d.Args {
		if s != "/pfs/datacenter/home/user1/1.txt" &&
			s != "/pfs/datacenter/home/user1/" {
			t.Error(s)
		}
	}
}
