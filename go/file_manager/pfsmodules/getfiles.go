package pfsmodules

import (
	"log"
	"os"
	"path/filepath"
)

type GetFilesReq struct {
	Method    string   `json:"Method"`
	Options   []string `json:"Options"`
	FilesPath []string `json:"FilesPath"`
}

type FileMeta struct {
	Err     string `json:"Err"`
	Path    string `json:"Path"`
	ModTime string `json:"ModTime"`
	Size    int64  `json:"Size"`
	IsDir   bool   `json:IsDir`
}

type GetFilesResponse struct {
	Err          string     `json:"Err"`
	Metas        []FileMeta `json:"Metas"`
	TotalObjects int        `json:"TotalObjects"`
	TotalSize    int64      `json:"TotalSize"`
}

func NewGetFilesResponse() *GetFilesResponse {
	return &GetFilesResponse{
		Err:          "",
		Metas:        make([]FileMeta, 10),
		TotalObjects: 0,
		TotalSize:    0,
	}
}

func (o *GetFilesResponse) SetErr(err string) {
	o.Err = err
}

func (o *GetFilesResponse) GetErr() string {
	return o.Err
}

func LsPath(path string, r bool) ([]FileMeta, error) {

	metas := make([]FileMeta, 0, 100)

	list, err := filepath.Glob(path)
	if err != nil {
		m := FileMeta{}
		m.Err = err.Error()
		metas = append(metas, m)

		log.Printf("glob path:%s error:%s", path, m.Err)
		return metas, err
	}

	if len(list) == 0 {
		m := FileMeta{}
		m.Err = "file or directory not exist"
		metas = append(metas, m)

		log.Printf("glob path:%s error:%s", path, m.Err)
		return metas, err
	}

	for _, v := range list {
		log.Println("v:\t" + v)
		filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
			//log.Println("path:\t" + path)

			m := FileMeta{}
			m.Path = info.Name()
			m.Size = info.Size()
			m.ModTime = info.ModTime().Format("2006-01-02 15:04:05")
			m.IsDir = info.IsDir()
			metas = append(metas, m)

			if info.IsDir() && !r && v != path {
				return filepath.SkipDir
			}

			//log.Println(len(metas))
			return nil
		})

	}

	return metas, nil
}

func LsPaths(paths []string, r bool) ([]FileMeta, error) {

	metas := make([]FileMeta, 0, 100)

	for _, path := range paths {
		m, err := LsPath(path, r)

		if err != nil {
			metas = append(metas, m...)
			continue
		}

		metas = append(metas, m...)
	}

	return metas, nil
}
