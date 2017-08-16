package pfsmodules

import (
	"errors"
	"fmt"
	"io"
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
	r := FileHandle{}
	if err := r.Open(src, os.O_RDONLY); err != nil {
		return err
	}
	defer r.Close()

	w := RFileHandle{}
	if err := w.Open(dst, os.O_RDWR, srcFileSize); err != nil {
		return err
	}
	defer w.Close()

	// upload chunks.
	size := defaultChunkSize
	var offset int64 = 0
	for {
		c, err := r.ReadChunk(offset, size)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		offset += c.Len
		log.V(2).Infoln(c.ToString())

		m, err := w.GetChunkMeta(offset, size)
		if err != nil {
			return err
		}
		log.V(2).Infoln(m.ToString())

		if c.Checksum == m.Checksum {
			log.V(2).Infof("remote and local chunk are same chunk info:%s\n", offset, c.ToString())
			continue
		}

		if err := w.Write(c); err != nil {
			return err
		}
		log.V(2).Infoln("put chunk:" + c.ToString())
	}

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
