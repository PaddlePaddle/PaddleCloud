package paddlecloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	log "github.com/golang/glog"
	"path/filepath"
)

func RemoteStat(s *PfsSubmitter, cmd *pfsmod.StatCmd) (*pfsmod.LsResult, error) {
	body, err := s.GetFiles(cmd)
	if err != nil {
		return nil, err
	}

	resp := pfsmod.StatResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	log.V(1).Infof("stat %s result:%#v\n", resp)

	if len(resp.Err) != 0 {
		return nil, errors.New(resp.Err)
	}

	return &resp.Results, nil
}

func RemoteTouch(s *PfsSubmitter, cmd *pfsmod.TouchCmd) error {
	body, err := s.PostFiles(cmd)
	if err != nil {
		return err
	}

	resp := pfsmod.TouchResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if len(resp.Err) == 0 {
		return nil
	}

	return errors.New(resp.Err)
}

func UploadChunks(s *PfsSubmitter,
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
		body, err := s.PostChunkData(pfsmod.NewChunkCmd(src, meta.Offset, meta.Len), dst)
		if err != nil {
			return err
		}

		resp := pfsmod.UploadChunkResponse{}
		if err := json.Unmarshal(body, &resp); err != nil {
			return err
		}

		if len(resp.Err) == 0 {
			return nil
		}

		return errors.New(resp.Err)
	}

	return nil
}

func UploadFile(s *PfsSubmitter,
	src, dst string, srcFileSize int64) error {

	log.V(1).Infof("touch %s size:%d\n", dst, srcFileSize)
	if err := RemoteTouch(s, pfsmod.NewTouchCmd(dst, srcFileSize)); err != nil {
		return err
	}

	dstMeta, err := RemoteChunkMeta(s, dst, pfsmod.DefaultChunkSize)
	if err != nil {
		return err
	}
	log.V(2).Infof("dst %s chunkMeta:%#v\n", dst, dstMeta)

	srcMeta, err := pfsmod.GetChunkMeta(src, pfsmod.DefaultChunkSize)
	if err != nil {
		return err
	}
	log.V(2).Infof("src %s chunkMeta:%#v\n", src, srcMeta)

	diffMeta, err := pfsmod.GetDiffChunkMeta(srcMeta, dstMeta)
	if err != nil {
		return err
	}
	log.V(2).Infof("diff chunkMeta:%#v\n", diffMeta)

	err = UploadChunks(s, src, dst, diffMeta)
	if err != nil {
		return err
	}

	return nil
}

func Upload(s *PfsSubmitter, src, dst string) error {
	lsCmd := pfsmod.NewLsCmd(true, src)
	srcRet, err := lsCmd.Run()
	if err != nil {
		return err
	}
	log.V(1).Infof("ls src:%s result:%#v\n", src, srcRet)

	dstMeta, err := RemoteStat(s, pfsmod.NewStatCmd(dst))
	if err != nil {
		return err
	}
	log.V(1).Infof("stat dst:%s result:%#v\n", dst, dstMeta)

	srcMetas := srcRet.([]pfsmod.LsResult)

	for _, srcMeta := range srcMetas {
		if srcMeta.IsDir {
			return errors.New(pfsmod.StatusText(pfsmod.StatusOnlySupportFiles))
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
