package main

import (
	"flag"
	"fmt"
	"github.com/cloud/paddlecloud/go/file_manager/pfsserver"
	"log"
	"net/http"
)

func main() {

	router := pfsserver.NewRouter()

	portPtr := flag.Int("port", 8080, "listen port")
	flag.Parse()

	addr := fmt.Sprintf("0.0.0.0:%d", *portPtr)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("server on:%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
