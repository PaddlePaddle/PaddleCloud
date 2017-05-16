package pfsserver

import (
	"encoding/json"
	"github.com/cloud/paddlecloud/go/file_manager/pfscommon"
	"github.com/cloud/paddlecloud/go/file_manager/pfsmodules"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func makeResponse(w http.ResponseWriter,
	rep pfsmodules.GetFilesResponse,
	status int) {

	log.SetFlags(log.LstdFlags)

	if len(rep.Err) > 0 {
		log.Printf("%s error:%s\n", pfscommon.CallerFileLine(), rep.Err)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(rep); err != nil {
		log.Printf("encode err:%s\n", err.Error())
		panic(err)
	}

}

func GetFiles(w http.ResponseWriter, r *http.Request) {
	var req pfsmodules.GetFilesReq
	rep := pfsmodules.GetFilesResponse{}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 2048))
	if err != nil {
		panic(err)
	}

	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	if err := json.Unmarshal(body, &req); err != nil {
		rep.SetErr(err.Error())
		makeResponse(w, rep, 422)
		return
	}

	if req.Method != "ls" {
		rep.SetErr("Not surported method:" + req.Method)
		makeResponse(w, rep, 422)
		return
	}

	log.Print(req)
	t, err := pfsmodules.LsPaths(req.FilesPath, false)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(t); err != nil {
		panic(err)
	}
}

func CreateFiles(w http.ResponseWriter, r *http.Request) {
}

func PatchFiles(w http.ResponseWriter, r *http.Request) {
}
