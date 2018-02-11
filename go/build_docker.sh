#!/bin/bash
env GOOS=linux GOARCH=amd64 go build cmd/edl
env GOOS=linux GOARCH=amd64 go build cmd/pserver

docker build . -t paddlepaddle/paddlecloud-go-server
