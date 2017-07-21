package pfsmodules

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/PaddlePaddle/cloud/go/utils/restclient"
	log "github.com/golang/glog"
)

func remoteChunkMeta(path string,
	chunkSize int64) ([]ChunkMeta, error) {
	cmd := ChunkMetaCmd{
		Method:    ChunkMetaCmdName,
		FilePath:  path,
		ChunkSize: chunkSize,
	}

	t := fmt.Sprintf("%s/api/v1/chunks", Config.ActiveConfig.PfsEndpoint)
	ret, err := restclient.GetCall(t, cmd.ToURLParam())
	if err != nil {
		return nil, err
	}

	type chunkMetaResponse struct {
		Err     string      `json:"err"`
		Results []ChunkMeta `json:"results"`
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

func getChunkData(target string, chunk Chunk, dst string) error {
	log.V(1).Info("target url: " + target)

	resp, err := restclient.GetChunk(target, chunk.ToURLParam())
	if err != nil {
		return err
	}
	defer Close(resp.Body)

	if resp.Status != restclient.HTTPOK {
		return errors.New("http server returned non-200 status: " + resp.Status)
	}

	partReader := multipart.NewReader(resp.Body, DefaultMultiPartBoundary)
	for {
		part, error := partReader.NextPart()
		if error == io.EOF {
			break
		}

		if part.FormName() == "chunk" {
			recvCmd, err := ParseChunk(part.FileName())
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
	dst string, diffMeta []ChunkMeta) error {
	if len(diffMeta) == 0 {
		log.V(1).Infof("srcfile:%s and dstfile:%s are already same\n", src, dst)
		fmt.Printf("download ok\n")
		return nil
	}

	t := fmt.Sprintf("%s/api/v1/storage/chunks", Config.ActiveConfig.PfsEndpoint)
	for _, meta := range diffMeta {
		chunk := Chunk{
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

func downloadFile(src string, srcFileSize int64, dst string) error {
	srcMeta, err := remoteChunkMeta(src, defaultChunkSize)
	if err != nil {
		return err
	}
	log.V(4).Infof("srcMeta:%#v\n\n", srcMeta)

	dstMeta, err := GetChunkMeta(dst, defaultChunkSize)
	if err != nil {
		if os.IsNotExist(err) {
			if err = CreateSizedFile(dst, srcFileSize); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	log.V(4).Infof("dstMeta:%#v\n", dstMeta)

	diffMeta, err := GetDiffChunkMeta(srcMeta, dstMeta)
	if err != nil {
		return err
	}

	err = downloadChunks(src, dst, diffMeta)
	return err
}

func checkBeforeDownLoad(src []LsResult, dst string) (bool, error) {
	var bDir bool
	fi, err := os.Stat(dst)
	if err == nil {
		bDir = fi.IsDir()
		if !fi.IsDir() && len(src) > 1 {
			return bDir, errors.New(StatusDestShouldBeDirectory)
		}
	} else if os.IsNotExist(err) {
		return false, nil
	}

	return bDir, err
}

func download(src, dst string) error {
	log.V(1).Infof("download %s to %s\n", src, dst)
	lsRet, err := RemoteLs(NewLsCmd(true, src))
	if err != nil {
		return err
	}

	bDir, err := checkBeforeDownLoad(lsRet, dst)
	if err != nil {
		return err
	}

	for _, attr := range lsRet {
		if attr.IsDir {
			return errors.New(StatusOnlySupportFiles)
		}

		realSrc := attr.Path
		realDst := dst

		if bDir {
			_, file := filepath.Split(attr.Path)
			realDst = dst + "/" + file
		}

		fmt.Printf("download src_path:%s dst_path:%s\n", realSrc, realDst)
		if err := downloadFile(realSrc, attr.Size, realDst); err != nil {
			return err
		}
	}

	return nil
}
