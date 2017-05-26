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

type LsCmdResult struct {
	StatusCode int32      `json:"StatusCode"`
	StatusText string     `json:"StatusText"`
	Attrs      []FileAttr `json:"Attr"`
}

type LsCmd struct {
	Method string
	R      bool
	Args   []string
}

func (p *LsCmd) ToUrl() string {
}

func (p *LsCmd) ToJson() []byte {
}

func (p *LsCmd) Run() interface{} {
}

func NewLsCmdFromFlag(f *flag.FlagSet) (*LsCmd, error) {
	cmd := LsCmd{}

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

func NewLsCmd(r bool, path string) (*LsCmd, error) {
	return &LsCmd{
		R:    r,
		Args: []string{arg},
	}

}

func lsPath(path string, r bool) []FileAttr {
	//log.Println("path:\t" + path)
	ret := make([]FileAttr, 0, 100)

	filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		//log.Println("path:\t" + path)

		m := FileAttr{}
		m.Path = subpath
		m.Meta.Size = info.Size()
		m.Meta.ModTime = info.ModTime().Format("2006-01-02 15:04:05")
		m.Meta.IsDir = info.IsDir()
		ret = append(ret, m)

		if info.IsDir() && !r && subpath != path {
			return filepath.SkipDir
		}

		//log.Println(len(metas))
		return nil
	})

	return metas
}
