package pfsmod

type Response struct {
	Err     string      `json:"err"`
	Results interface{} `json:"results"`
}

type LsResponse struct {
	Err     string        `json:"err"`
	Results []LsCmdResult `json:"results"`
}

type ChunkMetaResponse struct {
	Err     string      `json:"err"`
	Results []ChunkMeta `json:"results"`
}

type TouchResponse struct {
	Err    string      `json:"err"`
	Result TouchResult `json:"results"`
}
