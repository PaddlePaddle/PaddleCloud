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

func RemoteTouch(s *PfsSubmitter, cmd *TouchCmd) (TouchResult, error) {
	body, err := s.PostFiles(cmd)
	if err != nil {
		return nil, err
	}

	resp := pfsmod.TouchResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if len(resp.Err) == 0 {
		return resp.Result, nil
	}

	return resp.Result, errors.New(resp.Err)
}

func UploadFile(s *PfsSubmitter,
	src, dst string, srcFileSize int64) error {

	if err := RemoteTouch(s, pfsmod.NewTouchCmd(dst, srcFileSize)); err != nil {
		return err
	}

	dstMeta, err := GetRemoteChunksMeta(s, dst, pfsmod.DefaultChunkSize)
	if err != nil {
		return err
	}
	log.Printf("dst %s chunkMeta:%v\n", dst, dstMeta)

	log.Printf("src:%s dst:%s\n", src, dst)
	srcMeta, err := pfsmod.GetChunksMeta(src, pfsmod.DefaultChunkSize)
	if err != nil {
		return err
	}
	log.Printf("src %s chunkMeta:%v\n", src, srcMeta)

	diffMeta, err := pfsmod.GetDiffChunksMeta(srcMeta, dstMeta)
	if err != nil {
		return err
	}

	err = UploadChunks(s, src, dst, diffMeta)
	if err != nil {
		return err
	}

	return nil
}

func Upload(s *PfsSubmitter, src, dst string) ([]pfsmod.CpCmdResult, error) {
	lsCmd := pfsmod.NewLsCmd(true, src)
	srcMetas, err := lsCmd.Run()
	if err != nil {
		return nil, err
	}

	dstMetas, err := RemoteLs(s, NewLsCmd(false, src))
	if err != nil && !pfsmod.IsNotExist(err) {
		return nil, err
	}

	//files must save under directories
	if len(srcMetas) > 1 && !pfsmod.IsDir(dstMeta) {
		return results, errors.New(pfsmod.StatusText(pfsmod.StatusDestShouldBeDirectory))
	}

	results := make([]pfsmod.CpCmdResult, 0, 100)

	for _, srcMeta := range srcMetas {
		if srcMeta.IsDir {
			return results, errors.New(pfsmod.StatusText[pfsmod.StatusOnlySupportFiles])
		}

		realSrc := srcMeta.Path
		realDst := dst

		_, file := filepath.Split(srcMeta.Path)
		if dstMeta.IsDir {
			realDst = dst + "/" + file
		}

		fmt.Printf("upload src_path:%s dst_path:%s\n", realSrc, realDst)
		if ret, err := UploadFile(realSrc, realDst, srcMeta.Size); err != nil {
			return results, err
		}

		results = append(results, ret...)
	}

	return nil, nil
}
