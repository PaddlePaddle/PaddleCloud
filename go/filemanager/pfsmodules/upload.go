package pfsmodules

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PaddlePaddle/cloud/go/utils/config"
	log "github.com/golang/glog"
)

// Config is global config object for pfs commandline
var Config = config.ParseDefaultConfig()

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

func uploadFile(src, dst string, srcFileSize int64) error {

	log.V(1).Infof("touch %s size:%d\n", dst, srcFileSize)

	/*
		cmd := TouchCmd{
			Method:   TouchCmdName,
			Path:     dst,
			FileSize: srcFileSize,
		}

		f := FileHandle{}
		if err := f.Open(src, os.O_RDONLY); err != nil {
			return err
		}

		// upload chunks.
			for {
				chunk, err := f.Load()
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}

				d, err := remoteChunkMeta(dst, defaultChunkSize)
				if err != nil {
					return err
				}
				log.V(2).Infof("dst %s chunkMeta:%#v\n", dst, d)

				c, err := f.Load(offset, defaultChunkSize)
				if err != nil {
					return err
				}
				log.V(2).Infof("src %s chunkMeta:%#v\n", src, srcMeta)

				if c.Sum != d.Sum {
				}
			}

			dstMeta, err := remoteChunkMeta(dst, defaultChunkSize)

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
	*/

	//return uploadChunks(src, dst, diffMeta)

	return nil
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
