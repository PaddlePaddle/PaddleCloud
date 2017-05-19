package pfsserver

import (
	//"encoding/json"
	//"github.com/cloud/go/file_manager/pfscommon"
	"github.com/cloud/go/file_manager/pfsmodules"
	"log"
	"net/http"
)

func GetFilesHandler(w http.ResponseWriter, r *http.Request) {

	resp := pfsmodules.LsCmdResponse{}

	req, err := pfsmodules.GetJsonRequestCmdAttr(r)
	if err != nil {
		resp.SetErr(err.Error())
		pfsmodules.WriteCmdJsonResponse(w, &resp, 422)
		return
	}

	if req.Method != "ls" {
		resp.SetErr("not surported method:" + req.Method)
		pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusMethodNotAllowed)
		return
	}

	if len(req.Args) == 0 {
		resp.SetErr("no args")
		pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusExpectationFailed)
		return

	}

	log.Print(req)

	lsCmd := pfsmodules.NewLsCmd(req, &resp)
	lsCmd.RunAndResponse(w)
	/*
		WriteCmdJsonResponse(w, lsCmd.GetResponse, http.StatusAccepted)

		t, _ := pfsmodules.LsPaths(req.FilesPath, false)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(t); err != nil {
			panic(err)
		}
	*/
}

func PostFilesHandler(w http.ResponseWriter, r *http.Request) {
	/*
			req := pfsmodules.PostFilesReq{}
			rep := pfsmodules.PostFilesResponse{}

			if err := pfscommon.Body2Json(w, r, &req, &rep); err != nil {
				return
			}

		req, err := pfsmodules.GetCmdJsonRequest(r)
		if err != nil {
			resp.SetErr("Not surported method:" + req.Method)
			WriteCmdJsonResponse(w, &resp, 422)
			return
		}

			switch req.Method {
			case "mkdir":
				mkdirCmd := NewLsCmd
				CreateDirs(req.Metas)
			case "touch":
				CreateFiles(req.Metas)
			default:
				rep.SetErr("not surpported method")
				pfscommon.MakeResponse(w, &rep, 422)
			}
	*/
}
