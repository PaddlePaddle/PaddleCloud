package paddlecoud

func DownloadChunks(src string, dest string, diffMeta []pfsmodules.ChunkMeta) error {
	if len(diffMeta) == 0 {
		log.Printf("srcfile:%s and destfile:%s are same\n", src, dest)
		return nil
	}

	s := NewCmdSubmitter(UserHomeDir() + "/.paddle/config")

	for _, meta := range diffMeta {
		cmdAttr := pfsmodules.FromArgs("getchunkdata", src, meta.Offset, meta.Len)
		err := s.GetChunkData(8080, cmdAttr, dest)
		if err != nil {
			log.Printf("download chunk error:%v\n", err)
			return err
		}
	}

	return nil
}

func DownloadFile(src string, srcFileSize int64, dest string, chunkSize int64) error {
	srcMeta, err := GetRemoteChunksMeta(src, chunkSize)
	if err != nil {
		return err
	}

	destMeta, err := pfsmodules.GetChunksMeta(dest, chunkSize)
	//log.Printf("GetChunkMeta %v\n", dest, err)
	if err != nil {
		if os.IsNotExist(err) {
			if err := pfsmodules.CreateSizedFile(dest, srcFileSize); err != nil {
				return err
			}
		} else {
			return err
		}
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

// Download files to dst
func Download(s *PfsSubmitter, src, dst string) ([]pfsmod.CpCommandResult, error) {
	lsRet, err := RemoteLs(s, NewLsCommand(true, src))
	if err != nil {
		return nil, err
	}

	if len(lsRet.StatusCode) != 0 {
		return nil, errors.New(lsRet.StatusText)
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
