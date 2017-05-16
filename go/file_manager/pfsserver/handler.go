package pfsserver

import (
	"encoding/json"
	"github.com/cloud/paddlecloud/go/file_manager/pfscommon"
	"github.com/cloud/paddlecloud/go/file_manager/pfsmodules"
	"log"
	"net/http"
)

func GetFiles(w http.ResponseWriter, r *http.Request) {
	req := pfsmodules.GetFilesReq{}
	rep := pfsmodules.GetFilesResponse{}

	if err := pfscommon.Body2Json(w, r, &req, &rep); err != nil {
		return
	}

	if req.Method != "ls" {
		rep.SetErr("Not surported method:" + req.Method)
		pfscommon.MakeResponse(w, &rep, 422)
		return
	}

	log.Print(req)
	t, _ := pfsmodules.LsPaths(req.FilesPath, false)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(t); err != nil {
		panic(err)
	}
}

func PostFiles(w http.ResponseWriter, r *http.Request) {
	req := pfsmodules.PostFilesReq{}
	rep := pfsmodules.PostFilesResponse{}

	if err := pfscommon.Body2Json(w, r, &req, &rep); err != nil {
		return
	}

	switch req.Method {
	case "mkdir":
		CreateDirs(req.Metas)
	case "touch":
		CreateFiles(req.Metas)
	default:
		rep.SetErr("not surpported method")
		pfscommon.MakeResponse(w, &rep, 422)
	}
}

func CreateDirs(Metas []pfsmodules.PostFileMeta) error {
	return nil
}

func CreateFiles(Metas []pfsmodules.PostFileMeta) error {
	return nil
}

func PatchFiles(w http.ResponseWriter, r *http.Request) error {
	return nil
}
