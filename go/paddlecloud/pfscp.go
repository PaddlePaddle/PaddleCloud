package main

import (
	"context"
	"flag"
	//"fmt"
	//"errors"
	"fmt"
	"github.com/cloud/go/file_manager/pfsmodules"
	"github.com/google/subcommands"
	"log"
)

type cpCommand struct {
	v bool
}

func (*cpCommand) Name() string     { return "cp" }
func (*cpCommand) Synopsis() string { return "copy files or directories" }
func (*cpCommand) Usage() string {
	return `cp [-v] <src> <dest>
	copy files or directories
	Options:
	`
}

func (p *cpCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.v, "v", false, "Cause cp to be verbose, showing files after they are copied.")
}

func (p *cpCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 2 {
		f.Usage()
		return subcommands.ExitFailure
	}

	attrs := pfsmodules.NewCpCmdAttr("cp", f)
	/*
		if err != nil {
			return subcommands.ExitFailure
		}
	*/

	results, err := RunCp(attrs)
	if err != nil {
		return subcommands.ExitFailure
	}

	log.Println(results)

	return subcommands.ExitSuccess
}

func RunCp(p *pfsmodules.CpCmdAttr) ([]pfsmodules.CpCmdResult, error) {
	src, err := p.GetSrc()
	if err != nil {
		return nil, err
	}

	dest, err := p.GetDest()
	if err != nil {
		return nil, err
	}

	//results := make([]CpCmdResult, 0, 100)

	var results, ret []pfsmodules.CpCmdResult
	//var err error

	for _, arg := range src {
		log.Printf("ls %s\n", arg)

		if pfsmodules.IsRemotePath(arg) {
			if pfsmodules.IsRemotePath(dest) {
				//remotecp
				ret, err = RemoteCp(p, arg, dest)
			} else {
				//download
				ret, err = Download(p, arg, dest)
			}
		} else {
			if pfsmodules.IsRemotePath(dest) {
				//upload
				ret, err = Upload(p, arg, dest)
			} else {
				//can't do that
				m := pfsmodules.CpCmdResult{}
				m.Err = pfsmodules.CopyFromLocalToLocal
				m.Src = arg
				m.Dest = dest

				ret = append(ret, m)
			}
		}

		//m := CopyGlobPath(arg, dest)
		results = append(results, ret...)
	}

	return results, nil
}

func GetRemoteChunksMeta(path string, chunkSize uint32) ([]pfsmodules.ChunkMeta, error) {
	cmdAttr := pfsmodules.ChunkMetaCmdAttr{
		Method:    "getchunkmeta",
		Path:      path,
		ChunkSize: chunkSize,
	}
	resp := pfsmodules.ChunkMetaCmdResponse{}
	s := NewCmdSubmitter(UserHomeDir() + "/.paddle/config")

	cmd := pfsmodules.NewChunkMetaCmd(&cmdAttr, &resp)
	err := s.SubmitChunkMetaRequest(8080, cmd)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return resp.Metas, err
	}

	//fmt.Println(body)
	//return subcommands.ExitSuccess
	return resp.Metas, err
}

func DownloadChunks(src string, dest string, diffMeta []pfsmodules.ChunkMeta) error {
	return nil
}

func DownloadFile(src string, dest string, chunkSize uint32) error {
	srcMeta, err := GetRemoteChunksMeta(src, chunkSize)
	if err != nil {
		return err
	}

	destMeta, err := pfsmodules.GetChunksMeta(dest, chunkSize)
	if err != nil {
		return err
	}

	diffMeta, err := pfsmodules.GetDiffChunksMeta(srcMeta, destMeta)
	if err != nil {
		return err
	}

	err = DownloadChunks(src, dest, diffMeta)
	if err != nil {
		return err
	}

	return nil
}

func Upload(cpCmdAttr *pfsmodules.CpCmdAttr, src, dest string) ([]pfsmodules.CpCmdResult, error) {
	return nil, nil
}

func Download(cpCmdAttr *pfsmodules.CpCmdAttr, src, dest string) ([]pfsmodules.CpCmdResult, error) {
	cmdAttr := cpCmdAttr.GetNewCmdAttr()

	lsResp, err := RemoteLs(cmdAttr)
	if err != nil {
		return nil, err
	}

	results := make([]pfsmodules.CpCmdResult, 0, 100)
	for _, lsResult := range lsResp.Results {
		m := pfsmodules.CpCmdResult{}
		m.Src = lsResult.Path
		m.Dest = dest

		if len(lsResult.Err) > 0 {
			fmt.Printf("%s error:%s\n", lsResult.Path, lsResult.Err)
			m.Err = lsResult.Err
			results = append(results, m)
			continue
		}

		if err := DownloadFile(lsResult.Path, dest, pfsmodules.DefaultChunkSize); err != nil {
			//fmt.Printf("%s error:%s\n", result.Path, result.Err)
			m.Err = lsResult.Err
			results = append(results, m)
			continue
		}

		results = append(results, m)
	}

	return results, nil
}

func RemoteCp(cpCmdAttr *pfsmodules.CpCmdAttr, src, dest string) ([]pfsmodules.CpCmdResult, error) {
	resp := pfsmodules.RemoteCpCmdResponse{}
	cmdAttr := cpCmdAttr.GetNewCmdAttr()
	s := NewCmdSubmitter(UserHomeDir() + "/.paddle/config")

	remoteCpCmd := pfsmodules.NewRemoteCpCmd(cmdAttr, &resp)
	_, err := s.SubmitCmdReqeust("POST", "api/v1/files", 8080, remoteCpCmd)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}

	return resp.Results, nil
}
