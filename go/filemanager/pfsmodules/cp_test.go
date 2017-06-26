package pfsmodules

import (
	"flag"
	"strings"
	"testing"
)

func TestNewCpCmdFromFlag(t *testing.T) {
	cmdLine := "cp -v /pfs/datacenter/home/user1/1.txt /pfs/datacenter/home/user1/2.txt ./user1/"
	a := strings.Split(cmdLine, " ")

	flag := flag.NewFlagSet("cp", flag.ExitOnError)
	flag.Bool("v", false, "")

	if err := flag.Parse(a[1:]); err != nil {
		t.Error(err.Error())
	}

	d, err := newCpCmdFromFlag(flag)
	if err != nil {
		t.Error(err.Error())
	}

	if d.Method != "cp" {
		t.Error(d.Method)
	}

	if !d.V {
		t.Error(d.V)
	}

	if d.Dst != "./user1/" {
		t.Error(d.Dst)
	}

	if len(d.Src) != 2 {
		t.Error(len(d.Src))
	}

	for _, s := range d.Src {
		if s != "/pfs/datacenter/home/user1/1.txt" &&
			s != "/pfs/datacenter/home/user1/2.txt" {
			t.Error(s)
		}
	}
}
