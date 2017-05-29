package pfsserver

import (
	"encoding/json"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	log "github.com/golang/glog"
	"io"
	"io/ioutil"
	"net/http"
)

func cmdHandler(w http.ResponseWriter, r *http.Request, cmd pfsmod.Command) {
	resp := pfsmod.JsonResponse{}

	cmd, err := pfsmod.NewLsCmdFromUrlParam(r.URL.RawQuery)
	if err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	if err := cmd.Check(); err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	result, err := cmd.Run()
	if err != nil {
		resp.Err = err.Error()
		resp.Results = result
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	resp.Results = result
	writeJsonResponse(w, r, http.StatusOK, &resp)

}

func lsHandler(w http.ResponseWriter, r *http.Request) {
	resp := pfsmod.JsonResponse{}

	cmd, err := pfsmod.NewLsCmdFromUrlParam(r.URL.RawQuery)
	if err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	if err := cmd.Check(); err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	result, err := cmd.Run()
	if err != nil {
		resp.Err = err.Error()
		resp.Results = result
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	resp.Results = result
	writeJsonResponse(w, r, http.StatusOK, &resp)
}

/*
func MD5SumCmdHandler(w http.ResponseWriter, req *pfsmod.CmdAttr) {
	resp := pfsmod.MD5SumResponse{}
	log.Print(req)

	cmd := pfsmod.NewMD5SumCmd(req, &resp)
	cmd.RunAndResponse(w)
}
*/

func writeJsonResponse(w http.ResponseWriter,
	r *http.Request,
	httpStatus int,
	resp *pfsmod.JsonResponse) {

	if httpStatus != http.StatusOK || len(resp.Err) > 0 {
		log.Errorf("%s httpStatus:%d resp:=%v\n",
			r.URL.RawQuery, httpStatus, resp.Err)
	} else {
		log.Infof("%s httpStatus:%d\n",
			r.URL.RawQuery, httpStatus)

		log.V(1).Infof("%s httpStatus:%d resp:%#v\n",
			r.URL.RawQuery, httpStatus, resp)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatus)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error(err)
	}
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
		writeJsonResponse(w, r,
			http.StatusMethodNotAllowed, &resp)
	}
}

func rmCmdHandler(w http.ResponseWriter, r *http.Request) {
	return
}

func parseJson(r *http.Request, cmd interface{}) error {

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, pfsmod.MaxJsonRequestSize))
	if err != nil {
		return err
	}

	if err := r.Body.Close(); err != nil {
		return err
	}

	if err := json.Unmarshal(body, cmd); err != nil {
		return err
	}

	return nil
}

func touchHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Infof("begin proc touchHandler\n")
	resp := pfsmod.JsonResponse{}

	cmd := pfsmod.TouchCmd{}
	if err := parseJson(r, &cmd); err != nil {
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	if err := cmd.Check(); err != nil {
		resp.Err = err.Error()
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	result, err := cmd.Run()
	if err != nil {
		resp.Results = result
		resp.Err = err.Error()
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	resp.Results = result

	writeJsonResponse(w, r, http.StatusOK, &resp)
}

func PostFilesHandler(w http.ResponseWriter, r *http.Request) {

	log.V(1).Infof("begin proc PostFilesHandler\n")

	/*
		resp := pfsmod.JsonResponse{}

		switch req.Method {
		case "rm":
			//rm
		case "touch":
			touchHandler(w, r)
		default:
			resp := pfsmod.JsonResponse{}
			writeJsonResponse(w, r, http.StatusMethodNotAllowed)
		}
	*/
}

func getChunkMetaHandler(w http.ResponseWriter, r *http.Request) {
	cmd, status := pfsmod.NewChunkMetaCmdFromUrl(r)
	if status != http.StatusOK {
		writeJsonResponse(w, r, status, &pfsmod.JsonResponse{})
		return
	}

	resp := pfsmod.JsonResponse{}
	result, err := cmd.Run()
	if err != nil {
		resp.Results = result
		resp.Err = err.Error()
		writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	resp.Results = result
	writeJsonResponse(w, r, http.StatusOK, &pfsmod.JsonResponse{})

	log.V(1).Infof("proc %s ok\n", r.URL.RawQuery)
}

func GetChunksMetaHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Infof("begin proc %s\n", r.URL.RawQuery)
	method := r.URL.Query().Get("method")

	switch method {
	case "GetChunkMeta":
		getChunkMetaHandler(w, r)
	default:
		writeJsonResponse(w, r, http.StatusMethodNotAllowed, &pfsmod.JsonResponse{})
	}
}

func GetChunkData(w http.ResponseWriter, r *http.Request) {
	//log.V(1).Infof("begin proc %s\n", r.URL.RawQuery)

	cmd, status := pfsmod.NewChunkCmdFromUrlParam(r.URL.RawQuery)
	if status != http.StatusOK {
		writeJsonResponse(w, r, status, &pfsmod.JsonResponse{})
		return
	}
	if err := cmd.WriteChunkData(w); err != nil {
		//resp.Err = err.Error()
		//writeJsonResponse(w, r, http.StatusOK, &resp)
		return
	}

	//writeJsonResponse(w, r, http.StatusOK, &resp)
	log.Infof("proc %s ok\n", r.URL.RawQuery)
}

func GetChunksHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Infof("begin proc %s\n", r.URL.RawQuery)

	method := r.URL.Query().Get("method")
	resp := pfsmod.JsonResponse{}

	switch method {
	case "GetChunkData":
		GetChunkData(w, r)
	default:
		writeJsonResponse(w, r, http.StatusMethodNotAllowed, &resp)
	}

	return
}

func PostChunksHandler(w http.ResponseWriter, r *http.Request) {
	log.V(1).Infof("begin proc PostChunksHandler\n")

	resp := pfsmod.JsonResponse{}
	partReader, err := r.MultipartReader()
	if err != nil {
		writeJsonResponse(w, r, http.StatusBadRequest, &resp)
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
			writeJsonResponse(w, r, status, &resp)
			break
		}

		if err := cmd.GetChunkData(part); err != nil {
			resp.Err = err.Error()
			writeJsonResponse(w, r, http.StatusOK, &resp)
			break
		}

		writeJsonResponse(w, r, http.StatusOK, &resp)
	}

	log.V(1).Infof("proc PostChunksHandler ok\n")
}
