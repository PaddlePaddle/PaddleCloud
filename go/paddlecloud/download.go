package paddlecloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	log "github.com/golang/glog"
)

// RemoteChunkMeta get ChunkMeta of path from cloud
func RemoteChunkMeta(s *pfsSubmitter, path string, chunkSize int64) ([]pfsmod.ChunkMeta, error) {

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

func downloadChunks(s *pfsSubmitter, src string, dst string, diffMeta []pfsmod.ChunkMeta) error {
	if len(diffMeta) == 0 {
		log.V(1).Infof("srcfile:%s and dstfile:%s are already same\n", src, dst)
		fmt.Printf("download ok\n")
		return nil
	}

	for _, meta := range diffMeta {
		err := s.GetChunkData(pfsmod.NewChunk(src, meta.Offset, meta.Len), dst)
		if err != nil {
			return err
		}
	}

	return nil
}

// DownloadFile downloads src to dst. If dst does not exists, create it with srcFileSize.
func DownloadFile(s *pfsSubmitter, src string, srcFileSize int64, dst string) error {
	srcMeta, err := RemoteChunkMeta(s, src, defaultChunkSize)
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

	err = downloadChunks(s, src, dst, diffMeta)
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

// Download files to dst
func Download(s *pfsSubmitter, src, dst string) error {
	lsRet, err := RemoteLs(s, pfsmod.NewLsCmd(true, src))
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
		if err := DownloadFile(s, realSrc, attr.Size, realDst); err != nil {
			return err
		}
	}

	return nil
}
