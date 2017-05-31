package pfsmod

// JSONResponse is the common response from server
type JSONResponse struct {
	Err     string      `json:"err"`
	Results interface{} `json:"results"`
}

// LsResponse is the response of LsCmd
type LsResponse struct {
	Err     string     `json:"err"`
	Results []LsResult `json:"results"`
}

// ChunkMetaResponse is the response of ChunkMetaCmd
type ChunkMetaResponse struct {
	Err     string      `json:"err"`
	Results []ChunkMeta `json:"results"`
}

// TouchResponse is the response of TouchCmd
type TouchResponse struct {
	Err     string      `json:"err"`
	Results TouchResult `json:"results"`
}

// RmResponse is the response of RmCmd
type RmResponse struct {
	Err     string     `json:"err"`
	Results []RmResult `json:"path"`
}

// UploadChunkResponse is the response of UploadChunk
type UploadChunkResponse struct {
	Err string `json:"err"`
}

// StatResponse is the response of StatCmd
type StatResponse struct {
	Err     string   `json:"err"`
	Results LsResult `json:"results"`
}

// MkdirResponse is the response of MkdirCmd
type MkdirResponse struct {
	Err     string        `json:"err"`
	Results []MkdirResult `json:"results"`
}
