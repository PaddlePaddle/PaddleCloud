package paddlecloud

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	"github.com/google/subcommands"
	"log"
	"os"
	"path/filepath"
)

func UploadFile(src, dest string, srcFileSize int64) error {
	if err := RemoteTouch(dest, srcFileSize); err != nil {
		return err
	}

	dstMeta, err := GetRemoteChunksMeta(dest, pfsmodules.DefaultChunkSize)
	if err != nil {
		return err
	}
	log.Printf("dest %s chunkMeta:%v\n", dest, dstMeta)

	log.Printf("src:%s dest:%s\n", src, dest)
	srcMeta, err := pfsmodules.GetChunksMeta(src, pfsmodules.DefaultChunkSize)
	if err != nil {
		return err
	}
	log.Printf("src %s chunkMeta:%v\n", src, srcMeta)

	diffMeta, err := pfsmodules.GetDiffChunksMeta(srcMeta, dstMeta)
	if err != nil {
		return err
	}

	err = UploadChunks(src, dest, diffMeta)
	if err != nil {
		return err
	}

	return nil
}

func Upload(src, dest string) ([]pfsmodules.CpCmdResult, error) {
	resp, err := localLs(src)
	if err != nil {
		return nil, err
	}

	log.Printf("dest file:%s\n", dest)
	destMeta, err := GetRemoteMeta(dest)
	if err != nil {
		return nil, err
	}

	results := make([]pfsmodules.CpCmdResult, 0, 100)

	for _, result := range resp.Results {
		m := pfsmodules.CpCmdResult{}
		m.Src = result.Path
		_, file := filepath.Split(m.Src)
		if destMeta.IsDir {
			m.Dest = dest + "/" + file
		} else {
			m.Dest = dest
		}

		if len(result.Err) > 0 {
			results = append(results, m)
			log.Printf("%s is a directory\n", m.Src)
			return results, errors.New(result.Err)
		}

		for _, meta := range result.Metas {
			m.Src = meta.Path
			_, file := filepath.Split(meta.Path)
			if destMeta.IsDir {
				m.Dest = dest + "/" + file
			} else {
				m.Dest = dest
			}

			if meta.IsDir {
				m.Err = pfsmodules.OnlySupportUploadOrDownloadFiles
				results = append(results, m)
				log.Printf("%s is a directory\n", meta.Path)
				return results, errors.New(m.Err)
			}

			log.Printf("src_path:%s dest_path:%s\n", m.Src, m.Dest)
			if err := UploadFile(m.Src, m.Dest, meta.Size); err != nil {
				m.Err = err.Error()
				results = append(results, m)
				log.Printf("upload %s  error:%s\n", meta.Path, m.Err)
				return results, errors.New(m.Err)
			}

			results = append(results, m)
		}
	}

	return nil, nil
}
