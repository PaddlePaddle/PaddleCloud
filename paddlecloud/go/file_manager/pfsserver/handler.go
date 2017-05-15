package pfsserver

import (
	"encoding/json"
	//"fmt"
	"github.com/cloud/paddlecloud/go/file_manager/pfscommon"
	"github.com/cloud/paddlecloud/go/file_manager/pfsmodules"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	//"strconv"
	//"github.com/gorilla/mux"
)

/*
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

func GetFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(todos); err != nil {
		panic(err)
	}
}

func TodoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var todoId int
	var err error
	if todoId, err = strconv.Atoi(vars["todoId"]); err != nil {
		panic(err)
	}
	todo := RepoFindTodo(todoId)
	if todo.Id > 0 {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(todo); err != nil {
			panic(err)
		}
		return
	}

	// If we didn't find it, 404
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	if err := json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Not Found"}); err != nil {
		panic(err)
	}

}
*/

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

func lsFiles(paths []string, r bool) ([]pfsmodules.FileMeta, error) {

	ret := make([]pfsmodules.FileMeta, 0, 100)

	for _, t := range paths {

		list, err := filepath.Glob(t)
		if err != nil {
			log.Printf("glob path:%s error:%s", t, err)
			return nil, err
		}

		for _, v := range list {
			filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() && r {
					return filepath.SkipDir
				}
				m := pfsmodules.FileMeta{}
				m.Path = info.Name()
				m.Size = info.Size()
				m.ModTime = info.ModTime().Format("2006-01-02 15:04:05")

				ret = append(ret, m)
				return nil
			})
		}
	}

	return ret, nil
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

	//t := RepoCreateTodo(todo)
	log.Print(req)
	t, err := lsFiles(req.FilesPath, false)
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
