package pfsmodules

import (
	//"encoding/json"
	"errors"
	"io"
	"log"
	//"net/http"
	"flag"
	"os"
	"path/filepath"
)

type CpCmdResult struct {
	Err  string `json:"err"`
	Dest string `json:"Dest"`
	Src  string `json:"path"`
}

/*
type CpCmdAttr struct {
	Method  string   `json:"method"`
	Options []Option `json:"options"`
	Args    []string `json:"args"`
}
*/

type CpCmdAttr CmdAttr

func NewCpCmdAttr(cmdName string, f *flag.FlagSet) *CpCmdAttr {
	attr := NewCmdAttr(cmdName, f)

	return &CpCmdAttr{
		Method:  attr.Method,
		Options: attr.Options,
		Args:    attr.Args,
	}
}

func (p *CpCmdAttr) GetNewCmdAttr() *CmdAttr {
	return &CmdAttr{
		Method:  p.Method,
		Options: p.Options,
		Args:    p.Args,
	}
}

func (p *CpCmdAttr) Check() error {
	if len(p.Args) < 2 {
		return errors.New("too few arguments")
	}
	return nil
}

func (p *CpCmdAttr) GetSrc() ([]string, error) {
	if err := p.Check(); err != nil {
		return nil, err
	}

	src := p.Args[:len(p.Args)-1]
	return src, nil
}

func (p *CpCmdAttr) GetDest() (string, error) {
	if err := p.Check(); err != nil {
		return "", err
	}

	dest := p.Args[len(p.Args)-1]
	return dest, nil
}

// Copies file source to destination dest.
func CopyFile(source string, dest string) (err error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	if err == nil {
		return err
	}

	return nil
}

func CopyDir(src, dest string) ([]CpCmdResult, error) {
	log.Printf("src %s dest %s", src, dest)

	//results := CpCmdResults{}
	results := make([]CpCmdResult, 0, 100)

	filepath.Walk(src, func(subpath string, info os.FileInfo, err error) error {
		//log.Println("path:\t" + path)

		m := CpCmdResult{}
		m.Src = subpath + info.Name()
		m.Dest = dest + info.Name()

		if !info.IsDir() {
			err := CopyFile(m.Src, m.Dest)
			if err != nil {
				m.Err = err.Error()
				results = append(results, m)
				return err
			}
			results = append(results, m)
			return err
		}

		if subpath != src {
			err := os.MkdirAll(m.Dest, 0666)
			if err != nil {
				m.Err = err.Error()
				results = append(results, m)
				return err
			}
			return filepath.SkipDir
		}

		//log.Println(len(results))
		return nil
	})

	return results, nil
}

func CopyGlobPath(src, dest string) ([]CpCmdResult, error) {
	results := make([]CpCmdResult, 0, 100)

	m := CpCmdResult{}
	m.Src = src
	m.Dest = dest

	list, err := filepath.Glob(src)
	if err != nil {
		m.Err = err.Error()
		results = append(results, m)
		log.Printf("glob path:%s error:%s", src, m.Err)
		return results, err
	}

	if len(list) == 0 {
		m.Err = FileNotFound
		results = append(results, m)
		log.Printf("%s error:%s", src, m.Err)
		return results, errors.New(m.Err)
	}

	for _, path := range list {
		m := CpCmdResult{}
		m.Src = path
		m.Dest = dest

		fi, err := os.Stat(m.Src)
		if err != nil {
			m.Err = err.Error()
			results = append(results, m)
			log.Printf("%s to %s error:%s", m.Src, m.Dest, m.Err)
			return results, err
		}

		if fi.IsDir() {
			ret, err := CopyDir(m.Src, m.Dest)
			results = append(results, ret...)
			return results, err
		} else {
			err := CopyFile(m.Src, m.Dest)
			if err != nil {
				m.Err = err.Error()
				results = append(results, m)
				log.Printf("%s to %s error:%s", m.Src, m.Dest, m.Err)
			} else {
				results = append(results, m)
				log.Printf("%s to %s error:%s", m.Src, m.Dest, m.Err)
			}
		}
	}

	return results, nil
}
