package pfsserver

import (
	//"encoding/json"
	//"github.com/cloud/go/file_manager/pfscommon"
	//"fmt"
	//"github.com/cloud/go/file_manager/pfscommon"
	"github.com/cloud/go/file_manager/pfsmodules"
	"io"
	"log"
	//"mime/multipart"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

func lsCmdHandler(w http.ResponseWriter, req *pfsmodules.CmdAttr) {
	resp := pfsmodules.LsCmdResponse{}

	log.Print(req)

	cmd := pfsmodules.NewLsCmd(req, &resp)
	cmd.RunAndResponse(w)

	return
}

func MD5SumCmdHandler(w http.ResponseWriter, req *pfsmodules.CmdAttr) {
	resp := pfsmodules.MD5SumResponse{}
	log.Print(req)

	cmd := pfsmodules.NewMD5SumCmd(req, &resp)
	cmd.RunAndResponse(w)
}

func GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	resp := pfsmodules.LsCmdResponse{}
	req, err := pfsmodules.GetJsonRequestCmdAttr(r)
	if err != nil {
		resp.SetErr(err.Error())
		pfsmodules.WriteCmdJsonResponse(w, &resp, 422)
		return
	}

	if len(req.Args) == 0 {
		resp.SetErr("no args")
		pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusExpectationFailed)
		return

	}

	switch req.Method {
	case "ls":
		lsCmdHandler(w, req)
	case "md5sum":
		MD5SumCmdHandler(w, req)
	default:
		resp.SetErr(http.StatusText(http.StatusMethodNotAllowed))
		pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusMethodNotAllowed)
	}

	log.Print(req)
}

func rmCmdHandler(w http.ResponseWriter, req *pfsmodules.CmdAttr) {
	resp := pfsmodules.RmCmdResponse{}

	log.Print(req)

	cmd := pfsmodules.NewRmCmd(req, &resp)
	cmd.RunAndResponse(w)

	return
}

func touchHandler(w http.ResponseWriter, req *pfsmodules.CmdAttr) {
	resp := pfsmodules.TouchCmdResponse{}

	//log.Print(req)

	cmd := pfsmodules.NewTouchCmd(req, &resp)
	cmd.RunAndResponse(w)

	return
}

func PostFilesHandler(w http.ResponseWriter, r *http.Request) {
	resp := pfsmodules.JsonResponse{}
	req, err := pfsmodules.GetJsonRequestCmdAttr(r)
	if err != nil {
		resp.SetErr(err.Error())
		pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusExpectationFailed)
		return
	}

	if len(req.Args) == 0 {
		resp.SetErr("no args")
		pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusExpectationFailed)
		return

	}

	log.Print(req)

	switch req.Method {
	case "rm":
		rmCmdHandler(w, req)
	case "touch":
		if len(req.Args) != 1 {
			resp.SetErr("please create only one file")
			pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusExpectationFailed)
			return
		}
		touchHandler(w, req)
	default:
		resp.SetErr(http.StatusText(http.StatusMethodNotAllowed))
		pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusMethodNotAllowed)
	}
}

func GetChunksMetaHandler(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")

	log.Println(r.URL.String())

	resp := pfsmodules.JsonResponse{}
	switch method {
	case "getchunkmeta":
		cmd := pfsmodules.GetChunkMetaCmd(w, r)
		if cmd == nil {
			return
		}
		cmd.RunAndResponse(w)
	default:
		resp.SetErr(http.StatusText(http.StatusMethodNotAllowed))
		pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusMethodNotAllowed)
	}
}

/*
// Streams upload directly from file -> mime/multipart -> pipe -> http-request
func streamingUploadFile(fileName string, offset int64, len int64, w *io.PipeWriter, file *os.File) error {
	defer file.Close()
	defer w.Close()
	writer := multipart.NewWriter(w)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	_, err = io.CopyN(part, file, len)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}
	return nil
}

func getChunkData(path string, offset int64, len int64, w *http.ResponseWriter) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	fileName := pfsmodules.GetFileNameParam(path, offset, len)
	log.Printf("filename param %s", fileName)

	go streamingUploadFile(fileName, offset, len, writer, file)

	_, err = http.NewRequest("POST", uri, reader)
	return err
}
*/

func writeStreamChunkData(path string, offset int64, len int64, w http.ResponseWriter) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	writer := multipart.NewWriter(w)
	defer writer.Close()

	fileName := pfsmodules.GetFileNameParam(path, offset, len)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}

	_, err = io.CopyN(part, file, len)
	if err != nil {
		return err
	}
	return nil
}

func GetChunksHandler(w http.ResponseWriter, r *http.Request) {
	resp := pfsmodules.JsonResponse{}
	req, err := pfsmodules.NewChunkCmdAttr(r)
	if err != nil {
		resp.SetErr(err.Error())
		pfsmodules.WriteCmdJsonResponse(w, &resp, 422)
		return
	}

	switch req.Method {
	case "getchunkdata":
		if err := writeStreamChunkData(req.Path, req.Offset, int64(req.ChunkSize), w); err != nil {
			resp.SetErr(err.Error())
			pfsmodules.WriteCmdJsonResponse(w, &resp, 422)
			return
		}
	default:
		resp.SetErr(http.StatusText(http.StatusMethodNotAllowed))
		pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusMethodNotAllowed)
	}

	return
}

func PatchChunksHandler(w http.ResponseWriter, r *http.Request) {
}

func PostChunksHandler(w http.ResponseWriter, r *http.Request) {
	resp := pfsmodules.JsonResponse{}
	partReader, err := r.MultipartReader()

	if err != nil {
		resp.SetErr("error:" + err.Error())
		pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusBadRequest)
		return
	}

	for {
		part, error := partReader.NextPart()
		if error == io.EOF {
			break
		}

		if part.FormName() == "chunk" {
			chunkCmdAttr, err := pfsmodules.ParseFileNameParam(part.FileName())
			if err != nil {
				resp.SetErr("error:" + err.Error())
				pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusInternalServerError)
				break
			}

			f, err := pfsmodules.GetChunkWriter(chunkCmdAttr.Path, chunkCmdAttr.Offset)
			if err != nil {
				resp.SetErr("open " + chunkCmdAttr.Path + "error:" + err.Error())
				pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusInternalServerError)
				//return err
				break
			}
			defer f.Close()

			writen, err := io.Copy(f, part)
			if err != nil || writen != int64(chunkCmdAttr.ChunkSize) {
				resp.SetErr("read " + strconv.FormatInt(writen, 10) + "error:" + err.Error())
				pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusBadRequest)
				//return err
				break
			}
		}
	}
}
