package pfsmodules

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/golang/glog"
)

func downloadChunks(src string,
	dst string) error {
	/*
		if len(diffMeta) == 0 {
			log.V(1).Infof("srcfile:%s and dstfile:%s are already same\n", src, dst)
			fmt.Printf("download ok\n")
			return nil
		}

		t := fmt.Sprintf("%s/api/v1/pfs/storage/chunks", Config.ActiveConfig.Endpoint)
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
	*/

	return nil
}

func downloadFile(src string, srcFileSize int64, dst string) error {
	/*
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
	*/
	return nil
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
