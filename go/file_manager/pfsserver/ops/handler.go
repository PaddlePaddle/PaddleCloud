package pfsserver

import (
	//"encoding/json"
	//"github.com/cloud/go/file_manager/pfscommon"
	"github.com/cloud/go/file_manager/pfsmodules"
	"log"
	"net/http"
	//"strconv"
)

func lsCmdHandler(w http.ResponseWriter, req *pfsmodules.CmdAttr) {
	resp := pfsmodules.LsCmdResponse{}

	/*
		if req.Method != "ls" {
			resp.SetErr("not surported method:" + req.Method)
			pfsmodules.WriteCmdJsonResponse(w, &resp, http.StatusMethodNotAllowed)
			return
		}
	*/

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

	if req.Method != "ls" {
		return
	}

	log.Print(req)

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

func GetChunksHandler(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")

	log.Println(r.URL.String())

	switch method {
	case "getchunkmeta":
		cmd := pfsmodules.GetChunkMetaCmd(w, r)
		if cmd == nil {
			return
		}
		cmd.RunAndResponse(w)
	default:
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func PostChunksHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("w")
	//for k
	path := r.URL.Query().Get("path")

	log.Println(path)
}
