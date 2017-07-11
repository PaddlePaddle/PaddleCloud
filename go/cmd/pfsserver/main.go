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
	tokenURI := flag.String("tokenuri", "http://cloud.paddlepaddle.org", "uri of token server")
	flag.Parse()

	router := pfsserver.NewRouter()
	addr := fmt.Sprintf("%s:%d", *ip, *port)
	pfsserver.TokenURI = *tokenURI

	log.Infof("server on:%s and tokenuri:%s\n", addr, *tokenURI)
	log.Fatal(http.ListenAndServe(addr, router))
}
