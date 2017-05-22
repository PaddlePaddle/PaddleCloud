package pfsmodules

import (
	//"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cloud/go/file_manager/pfscommon"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

/*
const (
	ErrFileNotFound = 1
	ErrNoAuth       = 2
)
*/

type Response interface {
	//GetErrCode() int32
	GetErr() string
	//SetErr(errcode int32, err string)
	SetErr(err string)
}

type JsonResponse struct {
	Err     string `json:"err"`
	Results []string
}

func (p *JsonResponse) GetErr() string {
	return p.Err
}

func (p *JsonResponse) SetErr(err string) {
	p.Err = err
}

const (
	FileNotFound          = "no such file or directory"
	DirectoryNotAFile     = "should be a file not a directory"
	NotSourportedMethod   = "not surported method"
	CopyFromLocalToLocal  = "don't support copy local to local"
	DestShouldBeDirectory = "dest should be a directory"
)

const (
	defaultMaxCreateFileSize = int64(4 * 1024 * 1024 * 1024 * 1024)
)

type Command interface {
	GetCmdAttr() *CmdAttr
	GetResponse() Response
	//PushRequest()
	//RushResponse()
	/*
		SetCmdAttr(cmd *CmdAttr)
		SetResponse(resp *Response)
	*/
}

type Option struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

/*
type CmdAttrFace interface {
	Name() string
}
*/

type CmdAttr struct {
	Method  string   `json:"method"`
	Options []Option `json:"options"`
	Args    []string `json:"args"`
}

func (c *CmdAttr) Name() string {
	return c.Method
}

func NewCmdAttr(cmdName string, f *flag.FlagSet) *CmdAttr {
	cmd := CmdAttr{}

	cmd.Method = cmdName
	cmd.Options = make([]Option, f.NFlag())
	cmd.Args = make([]string, 0, f.NArg())

	f.Visit(func(flag *flag.Flag) {
		option := Option{}
		option.Name = flag.Name
		option.Value = flag.Value.String()

		cmd.Options = append(cmd.Options, option)
	})

	for _, arg := range f.Args() {
		fmt.Println("arg:" + arg)
		cmd.Args = append(cmd.Args, arg)
	}

	return &cmd
}

const (
	MaxJsonRequestSize = 2048
)

func GetJsonRequestCmdAttr(r *http.Request) (*CmdAttr, error) {

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, MaxJsonRequestSize))
	if err != nil {
		return nil, err
	}

	if err := r.Body.Close(); err != nil {
		return nil, err
	}

	c := &CmdAttr{}
	if err := json.Unmarshal(body, c); err != nil {
		return nil, err
	}
	return c, nil
}

func WriteCmdJsonResponse(w http.ResponseWriter, r Response, status int) error {

	log.SetFlags(log.LstdFlags)

	if len(r.GetErr()) != 0 {
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

type UpdateFilesResult struct {
	CmdAttr string `json:"cmd"`
	Err     string `json:"err"`
	Path    string `json:"path"`
}

type UpdateFilesResponse struct {
	Err    string              `json:"Err"`
	Result []UpdateFilesResult `json:"result"`
}

func IsRemotePath(path string) bool {
	return strings.HasPrefix(path, "/pfs/")
}
