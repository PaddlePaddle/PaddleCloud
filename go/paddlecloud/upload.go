package paddlecloud

import (
	//"context"
	"errors"
	//"flag"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	//"github.com/google/subcommands"
	"log"
	//"os"
	"encoding/json"
	"path/filepath"
)

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
	dest string,
	diffMeta []pfsmod.ChunkMeta) error {
	if len(diffMeta) == 0 {
		log.Printf("srcfile:%s and destfile:%s are same\n", src, dest)
		return nil
	}

	for _, meta := range diffMeta {
		log.Printf("diffMeta:%v\n", meta)
		//cmdAttr := pfsmod.FromArgs("postchunkdata", dest, meta.Offset, meta.Len)
		//err := s.PostChunkData(8080, cmdAttr, src)
		body, err := s.PostChunkData(pfsmod.NewChunkCmd(src, meta.Offset, meta.Len))
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

	if err := RemoteTouch(s, pfsmod.NewTouchCmd(dst, srcFileSize)); err != nil {
		return err
	}

	dstMeta, err := RemoteChunkMeta(s, dst, pfsmod.DefaultChunkSize)
	if err != nil {
		return err
	}
	log.Printf("dst %s chunkMeta:%v\n", dst, dstMeta)

	log.Printf("src:%s dst:%s\n", src, dst)
	srcMeta, err := pfsmod.GetChunkMeta(src, pfsmod.DefaultChunkSize)
	if err != nil {
		return err
	}
	log.Printf("src %s chunkMeta:%v\n", src, srcMeta)

	diffMeta, err := pfsmod.GetDiffChunkMeta(srcMeta, dstMeta)
	if err != nil {
		return err
	}

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

	dstMetas, err := RemoteLs(s, pfsmod.NewLsCmd(false, src))
	if err != nil && !pfsmod.IsNotExist(err) {
		return err
	}

	srcMetas := srcRet.([]pfsmod.LsResult)
	//files must save under directories
	if len(srcMetas) > 1 && !pfsmod.IsDir(dstMetas) {
		return errors.New(pfsmod.StatusText(pfsmod.StatusDestShouldBeDirectory))
	}

	//results := make([]pfsmod.CpCmdResult, 0, 100)

	for _, srcMeta := range srcMetas {
		if srcMeta.IsDir {
			//return results, errors.New(pfsmod.StatusText(pfsmod.StatusOnlySupportFiles))
			return errors.New(pfsmod.StatusText(pfsmod.StatusOnlySupportFiles))
		}

		realSrc := srcMeta.Path
		realDst := dst

		_, file := filepath.Split(srcMeta.Path)
		if pfsmod.IsDir(dstMetas) {
			realDst = dst + "/" + file
		}

		fmt.Printf("upload src_path:%s dst_path:%s\n", realSrc, realDst)
		if err := UploadFile(s, realSrc, realDst, srcMeta.Size); err != nil {
			return err
		}

		//results = append(results, ret...)
	}

	return nil
}
