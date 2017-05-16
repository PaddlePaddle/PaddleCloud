package pfsmodules

type PostFileMeta struct {
	Path string `json:"Path"`
	Size int64  `json:"Size"`
}

type PostFilesReq struct {
	Method string         `json:"Method"`
	Metas  []PostFileMeta `json:"Meta"`
}

type ActResult struct {
	Err  string       `json:"Err"`
	Meta PostFileMeta `json:"Path"`
}

type PostFilesResponse struct {
	Err     string      `json:"Err"`
	Results []ActResult `json:"ActResult"`
}

func (o *PostFilesResponse) SetErr(err string) {
	o.Err = err
}

func (o *PostFilesResponse) GetErr() string {
	return o.Err
}
