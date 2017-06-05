package paddlecloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
	log "github.com/golang/glog"
)

// RemoteChunkMeta get ChunkMeta of path from cloud.
func RemoteChunkMeta(path string,
	chunkSize int64) ([]pfsmod.ChunkMeta, error) {
	cmd := pfsmod.ChunkMetaCmd{
		Method:    pfsmod.ChunkMetaCmdName,
		FilePath:  path,
		ChunkSize: chunkSize,
	}

	t := fmt.Sprintf("%s/api/v1/chunks", config.ActiveConfig.Endpoint)
	ret, err := GetCall(t, cmd.ToURLParam())
	if err != nil {
		return nil, err
	}

	type chunkMetaResponse struct {
		Err     string             `json:"err"`
		Results []pfsmod.ChunkMeta `json:"results"`
	}

	resp := chunkMetaResponse{}
	if err := json.Unmarshal(ret, &resp); err != nil {
		return nil, err
	}

	if len(resp.Err) == 0 {
		return resp.Results, nil
	}

	return resp.Results, errors.New(resp.Err)
}

func getChunkData(target string, chunk pfsmod.Chunk, dst string) error {
	log.V(1).Info("target url: " + target)

	resp, err := GetChunk(target, chunk.ToURLParam())
	if err != nil {
		return err
	}
	defer pfsmod.Close(resp.Body)

	if resp.Status != HTTPOK {
		return errors.New("http server returned non-200 status: " + resp.Status)
	}

	partReader := multipart.NewReader(resp.Body, pfsmod.DefaultMultiPartBoundary)
	for {
		part, error := partReader.NextPart()
		if error == io.EOF {
			break
		}

		if part.FormName() == "chunk" {
			recvCmd, err := pfsmod.ParseChunk(part.FileName())
			if err != nil {
				return errors.New(err.Error())
			}

			recvCmd.Path = dst

			if err := recvCmd.SaveChunkData(part); err != nil {
				return err
			}
		}
	}

	return nil
}

func downloadChunks(src string,
	dst string, diffMeta []pfsmod.ChunkMeta) error {
	if len(diffMeta) == 0 {
		log.V(1).Infof("srcfile:%s and dstfile:%s are already same\n", src, dst)
		fmt.Printf("download ok\n")
		return nil
	}

	t := fmt.Sprintf("%s/api/v1/chunks", config.ActiveConfig.Endpoint)
	for _, meta := range diffMeta {
		chunk := pfsmod.Chunk{
			Path:   src,
			Offset: meta.Offset,
			Size:   meta.Len,
		}

		err := getChunkData(t, chunk, dst)
		if err != nil {
			return err
		}
	}

	return nil
}

// DownloadFile downloads src to dst. If dst does not exists, create it with srcFileSize.
func DownloadFile(src string, srcFileSize int64, dst string) error {
	srcMeta, err := RemoteChunkMeta(src, defaultChunkSize)
	if err != nil {
		return err
	}
	log.V(2).Infof("srcMeta:%#v\n", srcMeta)

	dstMeta, err := pfsmod.GetChunkMeta(dst, defaultChunkSize)
	if err != nil {
		if os.IsNotExist(err) {
			if err := pfsmod.CreateSizedFile(dst, srcFileSize); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	diffMeta, err := pfsmod.GetDiffChunkMeta(srcMeta, dstMeta)
	if err != nil {
		return err
	}

	err = downloadChunks(src, dst, diffMeta)
	return err
}

func checkBeforeDownLoad(src []pfsmod.LsResult, dst string) (bool, error) {
	var bDir bool
	fi, err := os.Stat(dst)
	if err == nil {
		bDir = fi.IsDir()
		if !fi.IsDir() && len(src) > 1 {
			return bDir, errors.New(pfsmod.StatusDestShouldBeDirectory)
		}
	} else if err == os.ErrNotExist {
		if err = os.MkdirAll(dst, 0700); err != nil {
			bDir = true
			return bDir, err
		}
	}

	return bDir, err
}

// Download function downloads src to dst.
func Download(src, dst string) error {
	lsRet, err := RemoteLs(pfsmod.NewLsCmd(true, src))
	if err != nil {
		return err
	}

	bDir, err := checkBeforeDownLoad(lsRet, dst)
	if err != nil {
		return err
	}

	for _, attr := range lsRet {
		if attr.IsDir {
			return errors.New(pfsmod.StatusOnlySupportFiles)
		}

		realSrc := attr.Path
		realDst := dst

		if bDir {
			_, file := filepath.Split(attr.Path)
			realDst = dst + "/" + file
		}

		fmt.Printf("download src_path:%s dst_path:%s\n", realSrc, realDst)
		if err := DownloadFile(realSrc, attr.Size, realDst); err != nil {
			return err
		}
	}

	return nil
}
