package pfsmodules

type PostFileMeta struct {
	Path string `json:"Path"`
	Size int64  `json:"Size"`
}

type PostFilesReq struct {
	Method string         `json:"Method"`
	Meta   []PostFilesReq `json:"Meta"`
}

type ActResult struct {
	Err  string       `json:Err`
	Meta PostFilesReq `json:Path`
}

type PostFilesResponse struct {
	Results []ActResult `json:"ActResult"`
}
