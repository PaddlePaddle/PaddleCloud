package pfsmodules

import (
	//"crypto/md5"
	"encoding/json"
	"github.com/cloud/go/file_manager/pfscommon"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Response interface {
	GetErr() string
	SetErr(err string)
}

type FileAttr struct {
	Path    string `json:"path"`
	ModTime string `json:"modtime"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:isdir`
}

type LsCmdResult struct {
	Cmd   string     `json:"cmd"`
	Err   string     `json:"err"`
	Metas []FileAttr `json:"metas"`
}

type LsResponse struct {
	Err    string        `json:"err"`
	Result []LsCmdResult `json:"result"`
}

type MD5SumResult struct {
	Cmd    string `json:"cmd"`
	Err    string `json:"err"`
	Path   string `json:"path"`
	MD5Sum []byte `json:"md5sum"`
}

type MD5SumResponse struct {
	Err    string         `json:"err"`
	Result []MD5SumResult `json:"result"`
}

type UpdateFilesResult struct {
	Cmd  string `json:"cmd"`
	Err  string `json:"err"`
	Path string `json:"path"`
}

type UpdateFilesResponse struct {
	Err    string              `json:"Err"`
	Result []UpdateFilesResult `json:"result"`
}

//func (r *Response) WriteJsonResponse(w http.ResponseWriter, r *Response
func WriteJsonResponse(w http.ResponseWriter, r Response,
	status int) error {

	log.SetFlags(log.LstdFlags)

	if len(r.GetErr()) > 0 {
		log.Printf("%s error:%s\n", pfscommon.CallerFileLine(), r.GetErr())
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(r); err != nil {
		log.Printf("encode err:%s\n", err.Error())
		return err
	}

	return nil
}

type Option struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type Cmd struct {
	Method  string   `json:"method"`
	Options []Option `json:"options"`
	Args    []string `json:"args"`
}

/*
func NewCmd() {
	return &Cmd{
		Method : ""
		Options :

	}
}
*/

const (
	MaxJsonRequestSize = 2048
)

func (c *Cmd) GetJsonRequest(w http.ResponseWriter,
	r *http.Request,
	resp Response) error {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, MaxJsonRequestSize))
	if err != nil {
		return err
	}

	if err := r.Body.Close(); err != nil {
		return err
	}

	if err := json.Unmarshal(body, c); err != nil {
		resp.SetErr(err.Error())
		WriteJsonResponse(w, resp, 422)
		return err
	}
	return nil
}
