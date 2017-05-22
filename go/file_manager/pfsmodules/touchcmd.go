package pfsmodules

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	//"path/filepath"
	"errors"
	"strconv"
)

type TouchCmdResult struct {
	Err  string `json:"err"`
	Path string `json:"path"`
}

type TouchCmdResponse struct {
	Err     string           `json:"err"`
	Results []TouchCmdResult `json:"path"`
}

func (p *TouchCmdResponse) SetErr(err string) {
	p.Err = err
}

func (p *TouchCmdResponse) GetErr() string {
	return p.Err
}

type TouchCmd struct {
	cmdAttr *CmdAttr
	resp    *TouchCmdResponse
}

func NewTouchCmd(cmdAttr *CmdAttr, resp *TouchCmdResponse) *TouchCmd {
	return &TouchCmd{
		cmdAttr: cmdAttr,
		resp:    resp,
	}
}

func (p *TouchCmd) GetCmdAttr() *CmdAttr {
	return p.cmdAttr
}

func (p *TouchCmd) GetResponse() Response {
	return p.resp
}

func CreateSizedFile(path string, size int64) error {
	log.Printf("%s %d\n", path, size)
	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	if size <= 0 {
		return nil
	}

	_, err = fd.Seek(size-1, 0)
	if err != nil {
		return err
	}

	_, err = fd.Write([]byte{0})
	if err != nil {
		return err
	}
	return nil
}

func (p *TouchCmd) Run() error {
	fileSize := int64(0)
	for _, t := range p.cmdAttr.Options {
		//log.Printf("%s:%s\n", t.Name, t.Value)
		if t.Name == "file-size" {
			inputSize, err := strconv.ParseInt(t.Value, 10, 64)
			if err != nil {
				return err
			}

			fileSize = inputSize
			if fileSize < 0 || fileSize > defaultMaxCreateFileSize {
				return errors.New("too large file size")
			}
		}
	}

	//log.Println(p.cmd.Args)
	results := make([]TouchCmdResult, 0, 100)
	for _, path := range p.cmdAttr.Args {
		log.Printf("%s %s\n", p.cmdAttr.Name(), path)
		m := TouchCmdResult{}
		m.Path = path

		fi, err := os.Stat(path)
		if os.IsExist(err) && fi.IsDir() {
			m.Err = "directory already exist"
			results = append(results, m)
			log.Printf("touch path %s error:%s", path, m.Err)
			continue
		}

		//log.Printf("%d %d\n", fi.Size(), fileSize)
		if os.IsNotExist(err) || fi.Size() != fileSize {
			if err := CreateSizedFile(path, fileSize); err != nil {
				m.Err = err.Error()
				results = append(results, m)
				log.Printf("touch path %s error:%s", path, m.Err)
				continue
			}
			results = append(results, m)
		} else {
			results = append(results, m)
		}
	}

	p.resp.Results = results
	return nil
}

func (p *TouchCmd) RunAndResponse(w http.ResponseWriter) error {
	p.Run()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(p.resp); err != nil {
		//w.WriteHeader(http.StatusInternalServerError)
		log.Printf("write response error:%v", err)
		return err
	}
	return nil
}
