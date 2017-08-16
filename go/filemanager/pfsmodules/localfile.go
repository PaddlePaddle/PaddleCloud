package pfsmodules

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
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
}

// Open opens a file.
func (f *FileHandle) Open(path string, flag int, size int64) error {
	if flag == WriteOnly ||
		flag == ReadAndWrite {

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

// ReadChunk loads a chunk at offset with len.
func (f *FileHandle) ReadChunk(offset int64, len int64) (*Chunk, error) {
	if err := f.checkOffset(offset); err != nil {
		return nil, err
	}

	c := NewChunk(len)

	n, err := f.F.Read(c.Data)
	log.V(2).Infof("expect %d read %d\n", len, n)

	if err != nil && err != io.EOF {
		return nil, err
	}
	f.Offset += int64(n)

	c.Offset = offset
	c.Len = int64(n)
	sum := md5.Sum(c.Data[:n])
	c.Checksum = hex.EncodeToString(sum[:])

	return c, err
}

// Read loads filedata to io.Writer.
func (f *FileHandle) Read(w io.Writer, offset, len int64) error {
	if err := f.checkOffset(offset); err != nil {
		return err
	}

	n, err := io.CopyN(w, f.F, len)
	log.V(2).Infof("expect %d read %d\n", len, n)

	if err != nil && err != io.EOF {
		return err
	}
	f.Offset += int64(n)

	return err
}

// WriteChunk writes data to file.
func (f *FileHandle) WriteChunk(c *Chunk) error {
	return f.Write(bytes.NewReader(c.Data), c.Offset, c.Len)
}

// Write writes data from io.Reader.
func (f *FileHandle) Write(r io.Reader, offset int64, len int64) error {
	if err := f.checkOffset(offset); err != nil {
		return err
	}

	n, err := io.CopyN(f.F, r, len)
	log.V(2).Infof("expect write %d writen:%d\n", len, n)
	if err == nil {
		f.Offset += n
	}

	return err
}
