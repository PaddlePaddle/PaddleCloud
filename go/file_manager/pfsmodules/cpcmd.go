package pfsmodules

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type CpCmdResult struct {
	Err  string `json:"err"`
	Src  string `json:"path"`
	Dest string `json:"Dest"`
}

type CpCmdAttr struct {
	Method  string   `json:"method"`
	Options []Option `json:"options"`
	Args    []string `json:"args"`
}

func (p *CpCmdAttr) Check() error {
	if len(p.Args) < 2 {
		return errors.New("too few arguments")
	}
	return nil
}

func (p *CpCmdAtr) GetSrc() []string {
	if err := p.check(); err != nil {
		return err
	}

	src := p.Args[:len(p.Args)-1]
	return &src
}

func (p *CpCmdAttr) GetDest() []string {
	if err := p.check(); err != nil {
		return err
	}

	dest := p.Args[len(p.Args)-1 : len(p.Args)]
	return &dest
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

func CopyDir(src, dest string) []CpCmdResult {
	log.Printf("src %s dest %s", src, dest)

	//results := CpCmdResults{}
	Results := make([]CpCmdResult, 0, 100)

	filepath.Walk(src, func(subpath string, info os.FileInfo, err error) error {
		//log.Println("path:\t" + path)

		m := CpCmdResult{}
		m.Src = subpath + info.Name()
		m.Dest = dest + info.Name()

		if info.IsFile() {
			m.err = CopyFile(m.Src, m.Dest)
			Results = append(Results, m)
			continue
		}

		if subpath != src {
			m.err = os.MkdirAll(m.Dest, 0666)
			if err != nil {
				Results = append(Results, m)
				//return filepath.
				goto Last
			}
			return filepath.SkipDir
		}

		//log.Println(len(Results))
		return nil
	})

Last:
	//results.Results = Results
	return &Results
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
		log.Printf("glob path:%s error:%s", arg, m.Err)
		return results, err
	}

	if len(list) == 0 {
		m.Err = FileNotFound
		results = append(results, m)
		log.Printf("%s error:%s", arg, m.Err)
		return results, errors.New(m.Err)
	}

	for _, path := range list {
		m := CpCmdResult{}
		m.Src = path
		m.Dest = Dest

		fi, err := os.Stat(path)
		if err != nil {
			m.Err = err.Error()
			results = append(results, m)
			log.Printf("%s error:%s", arg, m.Err)
			continue
		}

		if fi.IsDir() {
			ret := CopyDir(m.Src, m.Dest)
			results = append(results, ret)
		} else {
			err := CopyFile(path, dest)
			if err != nil {
				m.Err = err.Error()
				results = append(results, m)
				log.Printf("%s error:%s", arg, m.Err)
			} else {
				results = append(results, m)
				log.Printf("%s error:%s", arg, m.Err)
			}
		}
	}

	return results, nil
}

func (p *CpCmdAttr) Run([]CpCmdResult, error) {
	//log.Println(p.cmd.Args)
	src, err := p.getSrc()
	if err != nil {
		return nil, err
	}

	dest, err := p.getDest()
	if err != nil {
		return nil, err
	}

	results := make([]CpCmdResult, 0, 100)

	for _, arg := range src {
		log.Printf("ls %s\n", arg)

		m := CopyGlobPath(arg, dest)
		results = append(results, m)
	}

	return results, nil
}
