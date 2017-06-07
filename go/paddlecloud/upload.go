package paddlecloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
	"github.com/PaddlePaddle/cloud/go/utils"
	log "github.com/golang/glog"
)

func remoteStat(cmd *pfsmod.StatCmd) (*pfsmod.LsResult, error) {
	t := fmt.Sprintf("%s/api/v1/files", utils.Config.ActiveConfig.Endpoint)
	log.V(3).Infoln(t)
	body, err := utils.GetCall(t, cmd.ToURLParam())
	if err != nil {
		return nil, err
	}

	type statResponse struct {
		Err     string          `json:"err"`
		Results pfsmod.LsResult `json:"results"`
	}

	resp := statResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	log.V(1).Infof("result:%#v\n", resp)

	if len(resp.Err) != 0 {
		return nil, errors.New(resp.Err)
	}

	return &resp.Results, nil
}

func remoteTouch(cmd *pfsmod.TouchCmd) error {
	j, err := cmd.ToJSON()
	if err != nil {
		return err
	}

	t := fmt.Sprintf("%s/api/v1/files", utils.Config.ActiveConfig.Endpoint)
	body, err := utils.PostCall(t, j)
	if err != nil {
		return err
	}

	type touchResponse struct {
		Err     string             `json:"err"`
		Results pfsmod.TouchResult `json:"results"`
	}

	resp := touchResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if len(resp.Err) == 0 {
		return nil
	}

	return errors.New(resp.Err)
}

type uploadChunkResponse struct {
	Err string `json:"err"`
}

func getChunkReader(path string, offset int64) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	_, err = f.Seek(offset, 0)
	if err != nil {
		pfsmod.Close(f)
		return nil, err
	}

	return f, nil
}

func getDstParam(src *pfsmod.Chunk, dst string) string {
	cmd := pfsmod.Chunk{
		Path:   dst,
		Offset: src.Offset,
		Size:   src.Size,
	}

	return cmd.ToURLParam().Encode()
}

func postChunk(src *pfsmod.Chunk, dst string) ([]byte, error) {
	f, err := getChunkReader(src.Path, src.Offset)
	if err != nil {
		return nil, err
	}
	defer pfsmod.Close(f)

	t := fmt.Sprintf("%s/api/v1/storage/chunks", utils.Config.ActiveConfig.Endpoint)
	log.V(4).Infoln(t)

	return utils.PostChunk(t, getDstParam(src, dst),
		f, src.Size, pfsmod.DefaultMultiPartBoundary)
}

func uploadChunks(src string,
	dst string,
	diffMeta []pfsmod.ChunkMeta) error {
	if len(diffMeta) == 0 {
		log.V(1).Infof("srcfile:%s and destfile:%s are same\n", src, dst)
		return nil
	}

	for _, meta := range diffMeta {
		log.V(3).Infof("diffMeta:%v\n", meta)

		chunk := pfsmod.Chunk{
			Path:   src,
			Offset: meta.Offset,
			Size:   meta.Len,
		}

		body, err := postChunk(&chunk, dst)
		if err != nil {
			return err
		}

		resp := uploadChunkResponse{}
		if err := json.Unmarshal(body, &resp); err != nil {
			return err
		}

		if len(resp.Err) == 0 {
			continue
		}

		return errors.New(resp.Err)
	}

	return nil
}

func uploadFile(src, dst string, srcFileSize int64) error {

	log.V(1).Infof("touch %s size:%d\n", dst, srcFileSize)

	cmd := pfsmod.TouchCmd{
		Method:   pfsmod.TouchCmdName,
		Path:     dst,
		FileSize: srcFileSize,
	}

	if err := remoteTouch(&cmd); err != nil {
		return err
	}

	dstMeta, err := remoteChunkMeta(dst, defaultChunkSize)
	if err != nil {
		return err
	}
	log.V(2).Infof("dst %s chunkMeta:%#v\n", dst, dstMeta)

	srcMeta, err := pfsmod.GetChunkMeta(src, defaultChunkSize)
	if err != nil {
		return err
	}
	log.V(2).Infof("src %s chunkMeta:%#v\n", src, srcMeta)

	diffMeta, err := pfsmod.GetDiffChunkMeta(srcMeta, dstMeta)
	if err != nil {
		return err
	}
	log.V(2).Infof("diff chunkMeta:%#v\n", diffMeta)

	return uploadChunks(src, dst, diffMeta)
}

func upload(src, dst string) error {
	lsCmd := pfsmod.NewLsCmd(true, src)
	srcRet, err := lsCmd.Run()
	if err != nil {
		return err
	}
	log.V(1).Infof("ls src:%s result:%#v\n", src, srcRet)

	dstMeta, err := remoteStat(&pfsmod.StatCmd{Path: dst, Method: pfsmod.StatCmdName})
	if err != nil && !strings.Contains(err.Error(), pfsmod.StatusFileNotFound) {
		return err
	}
	log.V(1).Infof("stat dst:%s result:%#v\n", dst, dstMeta)

	srcMetas := srcRet.([]pfsmod.LsResult)

	for _, srcMeta := range srcMetas {
		if srcMeta.IsDir {
			return errors.New(pfsmod.StatusOnlySupportFiles)
		}

		realSrc := srcMeta.Path
		realDst := dst

		_, file := filepath.Split(srcMeta.Path)
		if dstMeta != nil && dstMeta.IsDir {
			realDst = dst + "/" + file
		}

		log.V(1).Infof("upload src_path:%s src_file_size:%d dst_path:%s\n",
			realSrc, srcMeta.Size, realDst)
		fmt.Printf("uploading %s to %s", realSrc, realDst)
		if err := uploadFile(realSrc, realDst, srcMeta.Size); err != nil {
			fmt.Printf(" error %v\n", err)
			return err
		}

		fmt.Printf(" ok!\n")
	}

	return nil
}
