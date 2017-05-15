package pfsmodules

type GetFilesReq struct {
	Method    string   `json:"Method"`
	Options   []string `json:"Options"`
	FilePaths []string `json:"FilesPaths"`
}

type FileMeta struct {
	Path         string `json:"Path"`
	Err          string `json:"Err"`
	CreationTime string `json:"CreationTime"`
	Size         int64  `json:"Size"`
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
