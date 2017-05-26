package paddlecoud

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

func GetChunkMeta(s *PfsSubmitter, path string, chunkSize int64) ([]pfsmod.ChunkMeta, error) {

	cmd := pfsmod.NewChunkMetaCmd(path, chunkSize)

	ret, err := s.GetChunkMeta(cmd)
	if err != nil {
		return nil, err
	}

	results := pfsmod.ChunkMetaCmdReult
	if err := json.Unmarshal(ret, results); err != nil {
		return nil, err
	}

	return results, nil
}

func GetFileAttr(s *PfsSubmitter, filePath string) (*pfsmod.FileAttr, error) {
	lsCmd := pfsMod.NewLsCmd(false, filePath)

	lsResp, err := RemoteLs(lsCmd)
	if err != nil {
		return nil, err
	}

	if len(lsResp.StatusCode) != 0 {
		return nil, errors.New(lsResp.StatusText)
	}

	for _, attr := range lsResp.Attrs {
		if len(attr.StatusCode) != 0 {
			return nil, errors.New(lsResp.StatusText)
		}

		return &attr, nil
	}

	return nil, errors.New("internal error")
}

func DownloadFileChunks(s *PfsSubmitter, src string, dest string, diffMeta []pfsmod.ChunkMeta) error {
	if len(diffMeta) == 0 {
		fmt.Printf("srcfile:%s and destfile:%s are already same\n", src, dest)
		return nil
	}

	for _, meta := range diffMeta {
		cmdAttr := pfsmod.FromArgs("getchunkdata", src, meta.Offset, meta.Len)
		err := s.GetChunkData(8080, cmdAttr, dest)
		if err != nil {
			log.Printf("download chunk error:%v\n", err)
			return err
		}
	}

	return nil
}

func DownloadFile(s *PfsSubmitter, src string, srcFileSize int64, dst string) error {
	srcMeta, err := pfsmod.GetChunkMeta(src, chunkSize)
	if err != nil {
		return err
	}

	destMeta, err := pfsmod.GetChunkMeta(dest, chunkSize)
	if err != nil {
		if os.IsNotExist(err) {
			if err := pfsmod.CreateSizedFile(dest, srcFileSize); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	diffMeta, err := pfsmod.GetDiffChunksMeta(srcMeta, destMeta)
	if err != nil {
		return err
	}

	err = DownloadChunks(src, dest, diffMeta)
	if err != nil {
		return err
	}

	return nil
}

// Download files to dst
func Download(s *PfsSubmitter, src, dst string) ([]pfsmod.CpCommandResult, error) {
	lsRet, err := RemoteLs(s, NewLsCommand(true, src))
	if err != nil {
		return nil, err
	}

	if len(lsRet.Attr) > 1 {
		fi, err := os.Stat(dst)
		if err != nil {
			if err == os.ErrNotExist {
				os.MkdirAll(dst, 0755)
			} else {
				return nil, err
			}
		}

		if !fi.IsDir() {
			return nil, errors.New(pfsmod.DestShouldBeDirectory)
		}
	}

	results := make([]pfsmod.CpCommandResult, 0, 100)
	for _, attr := range lsRet.Attr {
		if attr.IsDir {
			return results, errors.New(pfsmod.StatusText(pfsmod.StatusOnlySupportFiles))
		}

		realSrc = attr.Path
		_, file := filepath.Split(attr.Path)
		realDst = dst + "/" + file

		fmt.Printf("download src_path:%s dst_path:%s\n", m.Src, m.Dest)
		if err := DownloadFile(m.Src, attr.Size, m.Dest, pfsmod.DefaultChunkSize); err != nil {
			return results, err
		}

		m := pfsmod.CpCommandResult{
			StatusCode: 0,
			StatusText: "",
			Src:        realSrc,
			Dst:        realDst,
		}

		results = append(results, m)
	}

	return results, nil
}
