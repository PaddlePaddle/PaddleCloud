package pfsserver

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"reflect"
	"testing"

	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
)

func TestGetChunk(t *testing.T) {
	param := pfsmod.ChunkParam{}
	param.Path = "./testdata/test_lt_chunk.dat"
	param.Offset = 0
	param.Size = int64(4096)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	err := getChunk(w, req, &param)
	if err == nil ||
		err.Error() != pfsmod.StatusFileEOF {
		t.Error(err)
	}

	fmt.Println(w)

	partReader := multipart.NewReader(w.Body, pfsmod.DefaultMultiPartBoundary)
	for {
		part, err := partReader.NextPart()
		if err == io.EOF {
			break
		}

		if part.FormName() != "chunk" {
			continue
		}

		m1, err := pfsmod.ParseChunkParam(part.FileName())
		if err != nil {
			t.Error(err)
		}

		if reflect.DeepEqual(m1, param) {
			t.Error("m1 != param, m1:%v", m1)
		}

		data := make([]byte, m1.Size)
		n, err := io.ReadFull(part, data)
		if err != io.ErrUnexpectedEOF {
			t.Error(err)
		}

		sum := md5.Sum(data[:n])
		checksum := hex.EncodeToString(sum[:])
		if checksum != "77c58f04583c86f78c51df158e3f35e8" {
			t.Error("checksum error")
		}
	}
}
