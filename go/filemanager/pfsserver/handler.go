package pfsserver

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	sjson "github.com/bitly/go-simplejson"
	log "github.com/golang/glog"
)

func cmdHandler(w http.ResponseWriter, req string, cmd pfsmod.Command) {
	resp := pfsmod.JSONResponse{}

	if err := cmd.CloudCheck(); err != nil {
		resp.Err = err.Error()
		writeJSONResponse(w, req, http.StatusOK, &resp)
		return
	}

	result, err := cmd.Run()
	if err != nil {
		resp.Err = err.Error()
		resp.Results = result
		writeJSONResponse(w, req, http.StatusOK, &resp)
		return
	}

	resp.Results = result
	writeJSONResponse(w, req, http.StatusOK, &resp)
}

func lsHandler(w http.ResponseWriter, r *http.Request) {
	cmd, err := pfsmod.NewLsCmdFromURLParam(r.URL.RawQuery)

	resp := pfsmod.JSONResponse{}
	if err != nil {
		resp.Err = err.Error()
		writeJSONResponse(w, r.URL.RawQuery, http.StatusOK, &resp)
		return
	}

	cmdHandler(w, r.URL.RawQuery, cmd)
}

func statHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Info("begin stathandler")
	cmd, err := pfsmod.NewStatCmdFromURLParam(r.URL.RawQuery)

	resp := pfsmod.JSONResponse{}
	if err != nil {
		resp.Err = err.Error()
		writeJSONResponse(w, r.URL.RawQuery, http.StatusOK, &resp)
		return
	}

	cmdHandler(w, r.URL.RawQuery, cmd)
}

func writeJSONResponse(w http.ResponseWriter,
	req string,
	httpStatus int,
	resp *pfsmod.JSONResponse) {

	if httpStatus != http.StatusOK || len(resp.Err) > 0 {
		log.Errorf("%s httpStatus:%d resp:=%v\n",
			req, httpStatus, resp.Err)
	} else {
		log.Infof("%s httpStatus:%d\n",
			req, httpStatus)

		log.V(1).Infof("%s httpStatus:%d resp:%#v\n",
			req, httpStatus, resp)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatus)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error(err)
	}
}

// GetFilesHandler processes files's GET request
func GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")

	switch method {
	case "ls":
		lsHandler(w, r)
	case "md5sum":
		// TODO
		// err := md5Handler(w, r)
	case "stat":
		statHandler(w, r)
	default:
		resp := pfsmod.JSONResponse{}
		writeJSONResponse(w, r.URL.RawQuery,
			http.StatusMethodNotAllowed, &resp)
	}
}

func rmHandler(w http.ResponseWriter, body []byte) {
	log.V(1).Infof("begin proc rmHandler\n")
	cmd := pfsmod.RmCmd{}

	resp := pfsmod.JSONResponse{}
	if err := json.Unmarshal(body, &cmd); err != nil {
		writeJSONResponse(w, string(body[:]), http.StatusOK, &resp)
		return
	}

	log.V(1).Infof("request :%#v\n", cmd)

	cmdHandler(w, string(body[:]), &cmd)
	log.V(1).Infof("end proc handler\n")

}

func mkdirHandler(w http.ResponseWriter, body []byte) {
	log.V(1).Infof("begin proc mkdir\n")
	cmd := pfsmod.MkdirCmd{}

	resp := pfsmod.JSONResponse{}
	if err := json.Unmarshal(body, &cmd); err != nil {
		writeJSONResponse(w, string(body[:]), http.StatusOK, &resp)
		return
	}

	log.V(1).Infof("request :%#v\n", cmd)

	cmdHandler(w, string(body[:]), &cmd)
	log.V(1).Infof("end proc mkdir\n")

}

func touchHandler(w http.ResponseWriter, body []byte) {
	log.V(1).Infof("begin proc touch\n")
	cmd := pfsmod.TouchCmd{}

	resp := pfsmod.JSONResponse{}
	if err := json.Unmarshal(body, &cmd); err != nil {
		writeJSONResponse(w, string(body[:]), http.StatusOK, &resp)
		return
	}

	log.V(1).Infof("request :%#v\n", cmd)

	cmdHandler(w, string(body[:]), &cmd)
	log.V(1).Infof("end proc touch\n")
}

func getBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, pfsmod.MaxJSONRequestSize))
	if err != nil {
		return nil, err
	}

	if err := r.Body.Close(); err != nil {
		return nil, err
	}

	return body, nil
}

func getMethod(body []byte) (string, error) {
	o, err := sjson.NewJson(body)
	if err != nil {
		return "", errors.New(pfsmod.StatusText(pfsmod.StatusJSONErr))
	}

	j := o.Get("method")
	if j == nil {
		return "", errors.New(pfsmod.StatusText(pfsmod.StatusJSONErr))
	}

	method, _ := j.String()
	if err != nil {
		return "", errors.New(pfsmod.StatusText(pfsmod.StatusJSONErr))
	}

	return method, nil
}

// PostFilesHandler processes files' POST request
func PostFilesHandler(w http.ResponseWriter, r *http.Request) {

	//get body
	resp := pfsmod.JSONResponse{}
	body, err := getBody(r)
	if err != nil {
		resp.Err = err.Error()
		writeJSONResponse(w, string(body[:]), http.StatusOK, &resp)
		return
	}

	method, err := getMethod(body)
	if err != nil {
		resp.Err = err.Error()
		writeJSONResponse(w, string(body[:]), http.StatusOK, &resp)
		return
	}

	switch method {
	case "rm":
		rmHandler(w, body)
	case "touch":
		touchHandler(w, body)
	case "mkdir":
		mkdirHandler(w, body)
	default:
		resp := pfsmod.JSONResponse{}
		writeJSONResponse(w, string(body[:]), http.StatusMethodNotAllowed, &resp)
	}
}

func getChunkMetaHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Infof("begin proc getChunkMeta\n")
	cmd, err := pfsmod.NewChunkMetaCmdFromURLParam(r)
	resp := pfsmod.JSONResponse{}

	if err != nil {
		resp.Err = err.Error()
		writeJSONResponse(w, r.URL.RawQuery, http.StatusOK, &resp)
		return
	}

	cmdHandler(w, r.URL.RawQuery, cmd)
	log.V(1).Infof("end proc getChunkMeta\n")
}

// GetChunkMetaHandler processes GET ChunkMeta  request
func GetChunkMetaHandler(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")

	switch method {
	case "GetChunkMeta":
		getChunkMetaHandler(w, r)
	default:
		writeJSONResponse(w, r.URL.RawQuery, http.StatusMethodNotAllowed, &pfsmod.JSONResponse{})
	}
}

// GetChunkHandler processes GET Chunk  request
func GetChunkHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Infof("begin proc GetChunkHandler")

	cmd, err := pfsmod.NewChunkCmdFromURLParam(r.URL.RawQuery)
	if err != nil {
		writeJSONResponse(w, r.URL.RawQuery, http.StatusOK, &pfsmod.JSONResponse{})
		return
	}

	writer := multipart.NewWriter(w)
	writer.SetBoundary(pfsmod.DefaultMultiPartBoundary)

	fileName := cmd.ToURLParam()
	part, err := writer.CreateFormFile("chunk", fileName)
	if err != nil {
		log.Error(err)
		return
	}

	if err = cmd.LoadChunkData(part); err != nil {
		log.Error(err)
		return
	}

	err = writer.Close()
	if err != nil {
		log.Error(err)
		return
	}

	log.V(1).Info("end proc GetChunkHandler")
	return
}

// PostChunkHandler processes POST Chunk request
func PostChunkHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Infof("begin proc PostChunksHandler\n")

	resp := pfsmod.JSONResponse{}
	partReader, err := r.MultipartReader()
	if err != nil {
		writeJSONResponse(w, "ChunkHandler", http.StatusBadRequest, &resp)
		return
	}

	for {
		part, err := partReader.NextPart()
		if err == io.EOF {
			break
		}

		if part.FormName() != "chunk" {
			continue
		}

		cmd, err := pfsmod.NewChunkCmdFromURLParam(part.FileName())
		if err != nil {
			resp.Err = err.Error()
			writeJSONResponse(w, part.FileName(), http.StatusOK, &resp)
			return
		}

		log.V(1).Infof("recv cmd:%#v\n", cmd)

		if err := cmd.SaveChunkData(part); err != nil {
			resp.Err = err.Error()
			writeJSONResponse(w, part.FileName(), http.StatusOK, &resp)
			return
		}

		writeJSONResponse(w, part.FileName(), http.StatusOK, &resp)
	}

	log.V(1).Infof("end proc PostChunksHandler\n")
}
