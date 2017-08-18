package pfsmodules

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"os"

	log "github.com/golang/glog"
)

// FileHandle is a local *os.File with offset.
type FileHandle struct {
	F      *os.File
	Offset int64
}

// Close closes FileHandle.
func (f *FileHandle) Close() {
	if f.F != nil {
		f.F.Close()
	}

	f.Offset = 0
}

// Open opens a file.
func (f *FileHandle) Open(path string, flag int, size int64) error {
	if flag != os.O_RDONLY &&
		flag != os.O_WRONLY &&
		flag != os.O_RDWR {
		return errors.New("only support os.O_RDONLY, os.O_WRONLY, os.O_RDWR")
	}

	if (flag == os.O_WRONLY ||
		flag == os.O_RDWR) && size >= 0 {

		cmd := TouchCmd{
			Method:   TouchCmdName,
			Path:     path,
			FileSize: size,
		}

		// attempt to create sized file if it does't exist or
		// file's size != size
		if err := localTouch(&cmd); err != nil {
			return err
		}
	}

	fd, err := os.OpenFile(path, flag, 0666)
	if err != nil {
		return err
	}

	f.F = fd
	return nil
}

func (f *FileHandle) checkOffset(offset int64) error {
	if offset != f.Offset {
		if _, err := f.F.Seek(offset, 0); err != nil {
			return err
		}
		f.Offset = offset
	}
	return nil
}

// GetChunkMeta returns ChunkMeta at offset with size.
func (f *FileHandle) GetChunkMeta(offset, size int64) (*ChunkMeta, error) {
	c, err := f.ReadChunk(offset, size)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &ChunkMeta{
		Offset:   offset,
		Checksum: c.Checksum,
		Len:      c.Len,
	}, err
}

// ReadChunk loads a chunk at offset with len.
func (f *FileHandle) ReadChunk(offset int64, size int64) (*Chunk, error) {
	if err := f.checkOffset(offset); err != nil {
		return nil, err
	}

	c := NewChunk(size)

	//n, err := f.F.Read(c.Data)
	n, err := io.ReadFull(f.F, c.Data)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	f.Offset += int64(n)

	c.Offset = offset
	c.Len = int64(n)
	sum := md5.Sum(c.Data[:n])
	c.Checksum = hex.EncodeToString(sum[:])

	log.V(3).Infof("f:%d offset:%d need offset:%d size:%d Readed Chunk:%s error_info:%v\n",
		f.F, f.Offset-int64(n), offset, size, c.String(), err)

	fi, errfi := f.F.Stat()
	log.V(3).Infof("f.stat size:%d error:%v", fi.Size(), errfi)

	if err == io.ErrUnexpectedEOF {
		err = io.EOF
	}

	return c, err
}

// CopyN loads filedata to io.Writer.
func (f *FileHandle) CopyN(w io.Writer, offset, len int64) error {
	if err := f.checkOffset(offset); err != nil {
		return err
	}

	n, err := io.CopyN(w, f.F, len)
	log.V(2).Infof("CopyN expect %d real %d\n", len, n)

	if err != nil && err != io.EOF {
		return err
	}
	f.Offset += int64(n)

	return err
}

// WriteChunk writes data to file.
func (f *FileHandle) WriteChunk(c *Chunk) error {
	if c.Len == 0 {
		return nil
	}

	return f.Write(bytes.NewReader(c.Data), c.Offset, c.Len)
}

// Write writes data from io.Reader.
func (f *FileHandle) Write(r io.Reader, offset int64, size int64) error {
	if err := f.checkOffset(offset); err != nil {
		return err
	}

	n, err := io.CopyN(f.F, r, size)
	if n > 0 {
		f.Offset += n
	}

	log.V(3).Infof("f:%d offset:%d need offset:%d size:%d writen:%d\n",
		f.F, f.Offset-n, offset, size, n)

	return err
}
