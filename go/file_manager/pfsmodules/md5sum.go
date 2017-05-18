package pfsmodules

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
