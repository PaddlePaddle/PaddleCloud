package pfsserver

import (
	"encoding/json"
	"errors"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	sjson "github.com/bitly/go-simplejson"
	log "github.com/golang/glog"
	"io"
	"io/ioutil"
	"net/http"
)

func cmdHandler(w http.ResponseWriter, req string, cmd pfsmod.Command) {
	resp := pfsmod.JsonResponse{}

	if err := cmd.CloudCheck(); err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, req, http.StatusOK, &resp)
		return
	}

	result, err := cmd.Run()
	if err != nil {
		resp.Err = err.Error()
		resp.Results = result
		writeJsonResponse(w, req, http.StatusOK, &resp)
		return
	}

	resp.Results = result
	writeJsonResponse(w, req, http.StatusOK, &resp)
}

func lsHandler(w http.ResponseWriter, r *http.Request) {
	cmd, err := pfsmod.NewLsCmdFromUrlParam(r.URL.RawQuery)

	resp := pfsmod.JsonResponse{}
	if err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, r.URL.RawQuery, http.StatusOK, &resp)
		return
	}

	cmdHandler(w, r.URL.RawQuery, cmd)
}

func statHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Info("begin stathandler")
	cmd, err := pfsmod.NewStatCmdFromUrlParam(r.URL.RawQuery)

	resp := pfsmod.JsonResponse{}
	if err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, r.URL.RawQuery, http.StatusOK, &resp)
		return
	}

	cmdHandler(w, r.URL.RawQuery, cmd)
}

/*
func md5sumHandler(w http.ResponseWriter, r *http.Request) {
	cmd, err := pfsmod.NewMd5CmdFromUrlParam(r.URL.RawQuery)
	if err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	cmdHandler(w, r, cmd)
}
*/

func writeJsonResponse(w http.ResponseWriter,
	req string,
	httpStatus int,
	resp *pfsmod.JsonResponse) {

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

	/*
		ret, _ := json.Marshal(&resp)
		log.V(2).Info(string(ret[:]))
	*/
}

func GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")

	switch method {
	case "ls":
		lsHandler(w, r)
	case "md5sum":
		//err := md5Handler(w, r)
	case "stat":
		statHandler(w, r)
	default:
		resp := pfsmod.JsonResponse{}
		writeJsonResponse(w, r.URL.RawQuery,
			http.StatusMethodNotAllowed, &resp)
	}
}

func rmHandler(w http.ResponseWriter, body []byte) {
	return
}

func touchHandler(w http.ResponseWriter, body []byte) {
	log.V(1).Infof("begin proc touch\n")
	cmd := pfsmod.TouchCmd{}

	resp := pfsmod.JsonResponse{}
	if err := json.Unmarshal(body, &cmd); err != nil {
		writeJsonResponse(w, string(body[:]), http.StatusOK, &resp)
		return
	}

	log.V(1).Infof("request :%#v\n", cmd)

	cmdHandler(w, string(body[:]), &cmd)
	log.V(1).Infof("end proc touch\n")
}

func getBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, pfsmod.MaxJsonRequestSize))
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
		return "", errors.New(pfsmod.StatusText(pfsmod.StatusJsonErr))
	}

	j := o.Get("method")
	if j == nil {
		return "", errors.New(pfsmod.StatusText(pfsmod.StatusJsonErr))
	}

	method, _ := j.String()
	if err != nil {
		return "", errors.New(pfsmod.StatusText(pfsmod.StatusJsonErr))
	}

	return method, nil
}

func PostFilesHandler(w http.ResponseWriter, r *http.Request) {

	//get body
	resp := pfsmod.JsonResponse{}
	body, err := getBody(r)
	if err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, string(body[:]), http.StatusOK, &resp)
		return
	}

	method, err := getMethod(body)
	if err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, string(body[:]), http.StatusOK, &resp)
		return
	}

	switch method {
	case "rm":
		rmHandler(w, body)
	case "touch":
		touchHandler(w, body)
	default:
		resp := pfsmod.JsonResponse{}
		writeJsonResponse(w, string(body[:]), http.StatusMethodNotAllowed, &resp)
	}
}

func getChunkMetaHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Infof("begin proc getChunkMeta\n")
	cmd, err := pfsmod.NewChunkMetaCmdFromUrl(r)
	resp := pfsmod.JsonResponse{}

	if err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, r.URL.RawQuery, http.StatusOK, &resp)
		return
	}

	cmdHandler(w, r.URL.RawQuery, cmd)
	log.V(1).Infof("end proc getChunkMeta\n")
}

func GetChunkMetaHandler(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")

	switch method {
	case "GetChunkMeta":
		getChunkMetaHandler(w, r)
	default:
		writeJsonResponse(w, r.URL.RawQuery, http.StatusMethodNotAllowed, &pfsmod.JsonResponse{})
	}
}

func GetChunkData(w http.ResponseWriter, r *http.Request) {
	cmd, status := pfsmod.NewChunkCmdFromUrlParam(r.URL.RawQuery)
	if status != http.StatusOK {
		writeJsonResponse(w, r.URL.RawQuery, status, &pfsmod.JsonResponse{})
		return
	}

	if err := cmd.WriteChunkData(w); err != nil {
		return
	}

	log.Infof("proc GetChunkDat ok\n")
}

func GetChunkHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Infof("begin proc %s\n", r.URL.RawQuery)

	method := r.URL.Query().Get("method")
	resp := pfsmod.JsonResponse{}

	switch method {
	case "GetChunkData":
		GetChunkData(w, r)
	default:
		writeJsonResponse(w, r.URL.RawQuery, http.StatusMethodNotAllowed, &resp)
	}

	return
}

func PostChunkHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Infof("begin proc PostChunksHandler\n")

	resp := pfsmod.JsonResponse{}
	partReader, err := r.MultipartReader()
	if err != nil {
		writeJsonResponse(w, "ChunkHandler", http.StatusBadRequest, &resp)
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

		cmd, status := pfsmod.NewChunkCmdFromUrlParam(part.FileName())
		if status != http.StatusOK {
			writeJsonResponse(w, part.FileName(), status, &resp)
			break
		}

		if err := cmd.GetChunkData(part); err != nil {
			resp.Err = err.Error()
			writeJsonResponse(w, part.FileName(), http.StatusOK, &resp)
			break
		}

		resp.Err = ""
		writeJsonResponse(w, part.FileName(), http.StatusOK, &resp)
	}

	log.V(1).Infof("end proc PostChunksHandler\n")
}
