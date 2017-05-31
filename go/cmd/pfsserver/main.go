package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/PaddlePaddle/cloud/go/filemanager/pfsserver"
	log "github.com/golang/glog"
)

func main() {

	port := flag.Int("port", 8080, "port of server")
	ip := flag.String("ip", "0.0.0.0", "ip of server")
	flag.Parse()

	router := pfsserver.NewRouter()
	addr := fmt.Sprintf("%s:%d", *ip, *port)

	log.Infof("server on:%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
