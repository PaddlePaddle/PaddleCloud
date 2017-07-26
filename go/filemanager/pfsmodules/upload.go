package pfsmodules

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PaddlePaddle/cloud/go/utils/config"
	"github.com/PaddlePaddle/cloud/go/utils/restclient"
	log "github.com/golang/glog"
)

// Config is global config object for pfs commandline
var Config = config.ParseDefaultConfig()

func remoteStat(cmd *StatCmd) (*LsResult, error) {
	t := fmt.Sprintf("%s/api/v1/pfs/files", Config.ActiveConfig.Endpoint)
	log.V(3).Infoln(t)
	body, err := restclient.GetCall(t, cmd.ToURLParam())
	if err != nil {
		return nil, err
	}

	type statResponse struct {
		Err     string   `json:"err"`
		Results LsResult `json:"results"`
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

func remoteTouch(cmd *TouchCmd) error {
	j, err := cmd.ToJSON()
	if err != nil {
		return err
	}

	t := fmt.Sprintf("%s/api/v1/pfs/files", Config.ActiveConfig.Endpoint)
	body, err := restclient.PostCall(t, j)
	if err != nil {
		return err
	}

	type touchResponse struct {
		Err     string      `json:"err"`
		Results TouchResult `json:"results"`
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
		Close(f)
		return nil, err
	}

	return f, nil
}

func getDstParam(src *Chunk, dst string) string {
	cmd := Chunk{
		Path:   dst,
		Offset: src.Offset,
		Size:   src.Size,
	}

	return cmd.ToURLParam().Encode()
}

func postChunk(src *Chunk, dst string) ([]byte, error) {
	f, err := getChunkReader(src.Path, src.Offset)
	if err != nil {
		return nil, err
	}
	defer Close(f)

	t := fmt.Sprintf("%s/api/v1/pfs/storage/chunks", Config.ActiveConfig.Endpoint)
	log.V(4).Infoln(t)

	return restclient.PostChunk(t, getDstParam(src, dst),
		f, src.Size, DefaultMultiPartBoundary)
}

func uploadChunks(src string,
	dst string,
	diffMeta []ChunkMeta) error {
	if len(diffMeta) == 0 {
		log.V(1).Infof("srcfile:%s and destfile:%s are same\n", src, dst)
		return nil
	}

	for _, meta := range diffMeta {
		log.V(3).Infof("diffMeta:%v\n", meta)

		chunk := Chunk{
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

	cmd := TouchCmd{
		Method:   TouchCmdName,
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

	srcMeta, err := GetChunkMeta(src, defaultChunkSize)
	if err != nil {
		return err
	}
	log.V(2).Infof("src %s chunkMeta:%#v\n", src, srcMeta)

	diffMeta, err := GetDiffChunkMeta(srcMeta, dstMeta)
	if err != nil {
		return err
	}
	log.V(2).Infof("diff chunkMeta:%#v\n", diffMeta)

	return uploadChunks(src, dst, diffMeta)
}

func upload(src, dst string) error {
	lsCmd := NewLsCmd(true, src)
	srcRet, err := lsCmd.Run()
	if err != nil {
		return err
	}
	log.V(1).Infof("ls src:%s result:%#v\n", src, srcRet)

	dstMeta, err := remoteStat(&StatCmd{Path: dst, Method: StatCmdName})
	if err != nil && !strings.Contains(err.Error(), StatusFileNotFound) {
		return err
	}
	log.V(1).Infof("stat dst:%s result:%#v\n", dst, dstMeta)

	srcMetas := srcRet.([]LsResult)

	for _, srcMeta := range srcMetas {
		if srcMeta.IsDir {
			return errors.New(StatusOnlySupportFiles)
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
