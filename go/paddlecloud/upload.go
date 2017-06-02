package paddlecloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
	log "github.com/golang/glog"
)

// RemoteStat gets StatCmd's result from server.
func RemoteStat(s *pfsSubmitter, cmd *pfsmod.StatCmd) (*pfsmod.LsResult, error) {
	body, err := s.GetFiles(cmd)
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

	log.V(1).Infof("stat %s result:%#v\n", resp)

	if len(resp.Err) != 0 {
		return nil, errors.New(resp.Err)
	}

	return &resp.Results, nil
}

// RemoteTouch touches a file on cloud.
func RemoteTouch(s *pfsSubmitter, cmd *pfsmod.TouchCmd) error {
	body, err := s.PostFiles(cmd)
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

func uploadChunks(s *pfsSubmitter,
	src string,
	dst string,
	diffMeta []pfsmod.ChunkMeta) error {
	if len(diffMeta) == 0 {
		log.V(1).Infof("srcfile:%s and destfile:%s are same\n", src, dst)
		fmt.Printf("upload ok!\n")
		return nil
	}

	for _, meta := range diffMeta {
		log.V(1).Infof("diffMeta:%v\n", meta)

		chunk := pfsmod.Chunk{
			Path:   src,
			Offset: meta.Offset,
			Size:   meta.Len,
		}

		body, err := s.PostChunkData(&chunk, dst)
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

// UploadFile uploads src file to dst.
func UploadFile(s *pfsSubmitter,
	src, dst string, srcFileSize int64) error {

	log.V(1).Infof("touch %s size:%d\n", dst, srcFileSize)

	cmd := pfsmod.TouchCmd{
		Method:   pfsmod.TouchCmdName,
		Path:     dst,
		FileSize: srcFileSize,
	}

	if err := RemoteTouch(s, &cmd); err != nil {
		return err
	}

	dstMeta, err := RemoteChunkMeta(s, dst, defaultChunkSize)
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

	err = uploadChunks(s, src, dst, diffMeta)

	return err
}

// Upload uploads src to dst.
func Upload(s *pfsSubmitter, src, dst string) error {
	lsCmd := pfsmod.NewLsCmd(true, src)
	srcRet, err := lsCmd.Run()
	if err != nil {
		return err
	}
	log.V(1).Infof("ls src:%s result:%#v\n", src, srcRet)

	dstMeta, err := RemoteStat(s, &pfsmod.StatCmd{Path: dst, Method: pfsmod.StatCmdName})
	if err != nil {
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
		fmt.Printf("uploading %s\n", realSrc)
		if err := UploadFile(s, realSrc, realDst, srcMeta.Size); err != nil {
			return err
		}
	}

	return nil
}

func getDstParam(src *pfsmod.Chunk, dst string) string {
	cmd := pfsmod.Chunk{
		Path:   dst,
		Offset: src.Offset,
		Size:   src.Size,
	}

	return cmd.ToURLParam()
}

func postChunkData(cmd *pfsmod.Chunk, dst string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/storage/chunks",
		s.config.ActiveConfig.Endpoint)
	log.V(1).Info("target url: " + url)

	fileName := getDstParam(cmd, dst)

	req, err := newPostChunkDataRequest(cmd, dst, url)
	if err != nil {
		return nil, nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer pfsmod.Close(resp.Body)

	if resp.Status != HTTPOK {
		return nil, errors.New("http server returned non-200 status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}
