package modules

import "time"

type GetFilesReq struct {
	method    string   `json:"method"`
	options   []string `json:"options"`
	filePaths []string `json:"filesPaths"`
}

type FileAttr struct {
	creationTime string
	size         int64
	filePath     string
}

type GetFilesResponse struct {
	fileAttr     []FileAttr `json:"FileAttr"`
	totalObjects int        `json:"totalObjects:"`
	totalSize    int64      `json:"totalSize"`
}
