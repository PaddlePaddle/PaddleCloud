package paddlecloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	log "github.com/golang/glog"
	"os"
	"path/filepath"
)

func RemoteChunkMeta(s *PfsSubmitter, path string, chunkSize int64) ([]pfsmod.ChunkMeta, error) {

	cmd := pfsmod.NewChunkMetaCmd(path, chunkSize)

	ret, err := s.GetChunkMeta(cmd)
	if err != nil {
		return nil, err
	}

	resp := pfsmod.ChunkMetaResponse{}
	if err := json.Unmarshal(ret, &resp); err != nil {
		return nil, err
	}

	if len(resp.Err) == 0 {
		return resp.Results, nil
	}

	return resp.Results, errors.New(resp.Err)
}

func DownloadChunks(s *PfsSubmitter, src string, dst string, diffMeta []pfsmod.ChunkMeta) error {
	if len(diffMeta) == 0 {
		log.V(1).Infof("srcfile:%s and dstfile:%s are already same\n", src, dst)
		fmt.Printf("download ok\n")
		return nil
	}

	for _, meta := range diffMeta {
		err := s.GetChunkData(pfsmod.NewChunkCmd(src, meta.Offset, meta.Len), dst)
		if err != nil {
			return err
		}
	}

	return nil
}

func DownloadFile(s *PfsSubmitter, src string, srcFileSize int64, dst string) error {
	srcMeta, err := RemoteChunkMeta(s, src, pfsmod.DefaultChunkSize)
	if err != nil {
		return err
	}
	log.V(2).Infof("srcMeta:%#v\n", srcMeta)

	dstMeta, err := pfsmod.GetChunkMeta(dst, pfsmod.DefaultChunkSize)
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

	err = DownloadChunks(s, src, dst, diffMeta)
	if err != nil {
		return err
	}

	return nil
}

// Download files to dst
func Download(s *PfsSubmitter, src, dst string) error {
	lsRet, err := RemoteLs(s, pfsmod.NewLsCmd(true, src))
	if err != nil {
		return err
	}

	if len(lsRet) > 1 {
		fi, err := os.Stat(dst)
		if err != nil {
			if err == os.ErrNotExist {
				os.MkdirAll(dst, 0755)
			} else {
				return err
			}
		}

		if !fi.IsDir() {
			return errors.New(pfsmod.StatusText(pfsmod.StatusDestShouldBeDirectory))
		}
	}

	for _, attr := range lsRet {
		if attr.IsDir {
			return errors.New(pfsmod.StatusText(pfsmod.StatusOnlySupportFiles))
		}

		realSrc := attr.Path
		_, file := filepath.Split(attr.Path)
		realDst := dst + "/" + file

		fmt.Printf("download src_path:%s dst_path:%s\n", realSrc, realDst)
		if err := DownloadFile(s, realSrc, attr.Size, realDst); err != nil {
			return err
		}

	}

	return nil
}
