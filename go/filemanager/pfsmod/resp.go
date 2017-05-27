package pfsmod

/*
type Response interface {
	GetErr() string
}
*/

type JsonResponse struct {
	Err     string      `json:"err"`
	Results interface{} `json:"results"`
}

type LsResponse struct {
	Err     string     `json:"err"`
	Results []LsResult `json:"results"`
}

type ChunkMetaResponse struct {
	Err     string      `json:"err"`
	Results []ChunkMeta `json:"results"`
}

type TouchResponse struct {
	Err    string      `json:"err"`
	Result TouchResult `json:"results"`
}

type UploadChunkResponse struct {
	Err string `json:"err"`
}

func (p *JsonResponse) GetErr() string {
	return p.Err
}

/*
func (p *LsResponse) GetErr() string {
	return p.Err
}

func (p *ChunkMetaResponse) GetErr() string {
	return p.Err
}

func (p *TouchResponse) GetErr() string {
	return p.Err
}

func (p *UploadChunkResponse) GetErr() string {
	return p.Err
}
*/
