package pfsmodules

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PaddlePaddle/cloud/go/utils/config"
	"github.com/fatih/color"
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

func uploadFile(src, dst string, srcFileSize int64, verbose bool, chunkSize int64) error {
	r := FileHandle{}
	if err := r.Open(src, os.O_RDONLY, 0); err != nil {
		return err
	}
	defer r.Close()

	w := RemoteFile{}
	if err := w.Open(dst, os.O_RDWR, srcFileSize); err != nil {
		return err
	}
	defer w.Close()

	// upload chunks.
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}

	offset := int64(0)
	for {
		start := time.Now()
		c, errc := r.ReadChunk(offset, chunkSize)
		if errc != nil && errc != io.EOF {
			return errc
		}
		log.V(2).Infoln("local chunk info:" + c.String())

		m, errm := w.GetChunkMeta(offset, chunkSize)
		if errm != nil && errm != io.EOF {
			return errm
		}
		log.V(2).Infoln("remote chunk info:" + m.String())
		offset += c.Len

		if c.Checksum == m.Checksum {
			if verbose {
				used := time.Since(start).Nanoseconds() / time.Millisecond.Nanoseconds()
				ColorInfoOverWrite("%s upload   %d%% %dKB/s", src, offset*100/srcFileSize, c.Len/used)
			}
			log.V(2).Infof("remote and local chunk are same chunk info:%s\n", c.String())
			if errm == io.EOF || errc == io.EOF {
				break
			}
			continue
		}

		if verbose {
			used := time.Since(start).Nanoseconds() / time.Millisecond.Nanoseconds()
			ColorInfoOverWrite("%s upload   %d%% %dKB/s", src, offset*100/srcFileSize, c.Len/used)
		}

		if err := w.WriteChunk(c); err != nil {
			return err
		}
		log.V(2).Infof("upload chunk:%s ok\n\n", c.String())
		if errm == io.EOF || errc == io.EOF {
			break
		}
	}

	if offset != srcFileSize {
		return fmt.Errorf("expect %d upload %d", srcFileSize, offset)
	}

	return nil
}

// ColorError print red ERROR before message.
func ColorError(format string, a ...interface{}) {
	color.New(color.FgRed).Printf("[ERROR]  ")
	fmt.Printf(format, a...)
}

// ColorInfo print green INFO before message.
func ColorInfo(format string, a ...interface{}) {
	color.New(color.FgGreen).Printf("[INFO]  ")
	fmt.Printf(format, a...)
}

// ColorInfoOverWrite print green INFO before message overwrite last line.
func ColorInfoOverWrite(format string, a ...interface{}) {
	fmt.Printf("\r%c[2K", 27)
	color.New(color.FgGreen).Printf("[INFO]  ")
	fmt.Printf(format, a...)
}

func upload(src, dst string, verbose bool, chunkSize int64) error {
	lsCmd := NewLsCmd(true, src)
	srcRet, err := lsCmd.Run()
	if err != nil {
		return err
	}
	log.V(3).Infof("ls src:%s result:%#v\n", src, srcRet)

	dstMeta, err := remoteStat(&StatCmd{Path: dst, Method: StatCmdName})
	if err != nil && !strings.Contains(err.Error(), StatusFileNotFound) {
		ColorError("Upload %s to %s error info:%s\n", src, dst, err)
		return err
	}
	log.V(3).Infof("stat dst:%s result:%#v\n", dst, dstMeta)

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

		if err := uploadFile(realSrc, realDst, srcMeta.Size, verbose, chunkSize); err != nil {
			ColorError("Upload %s to %s error:%v\n", realSrc, realDst, err)
			return err
		}
		ColorInfoOverWrite("Uploaded %s\n", realSrc)
	}

	return nil
}
