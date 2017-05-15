package pfsmodules

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
