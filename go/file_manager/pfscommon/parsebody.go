package pfscommon

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Rep interface {
	GetErr() string
	SetErr(err string)
}

func MakeResponse(w http.ResponseWriter,
	rep Rep,
	status int) {

	log.SetFlags(log.LstdFlags)

	if len(rep.GetErr()) > 0 {
		log.Printf("%s error:%s\n", CallerFileLine(), rep.GetErr())
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(rep); err != nil {
		log.Printf("encode err:%s\n", err.Error())
		panic(err)
	}
}

func Body2Json(w http.ResponseWriter,
	r *http.Request,
	req interface{},
	rep Rep) error {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 2048))
	if err != nil {
		panic(err)
	}

	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	if err := json.Unmarshal(body, &req); err != nil {
		rep.SetErr(err.Error())
		MakeResponse(w, rep, 422)
		return err
	}
	return nil
}
