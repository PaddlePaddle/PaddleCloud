package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"

	log "github.com/golang/glog"
)

// HTTPOK is ok status of http api call.
const HTTPOK = "200 OK"

var HttpClient = &http.Client{Transport: &http.Transport{}}

func MakeRequest(uri string, method string, body io.Reader,
	contentType string, query url.Values,
	authHeader map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		log.Errorf("new request %v\n", err)
		return nil, err
	}

	// default contentType is application/json.
	if len(contentType) == 0 {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range authHeader {
		req.Header.Set(k, v)
	}

	if query != nil {
		req.URL.RawQuery = query.Encode()
	}
	return req, nil
}

// MakeRequestToken use client token to make a authorized request.
func MakeRequestToken(uri string, method string, body io.Reader,
	contentType string, query url.Values) (*http.Request, error) {
	// get client token
	token, err := token()
	if err != nil {
		return nil, err
	}
	authHeader := make(map[string]string)
	authHeader["Authorization"] = "Token " + token
	return MakeRequest(uri, method, body, contentType, query, authHeader)
}

// NOTE: add other request makers if we need other auth methods.

func GetResponse(req *http.Request) ([]byte, error) {
	resp, err := HttpClient.Do(req)
	if err != nil {
		log.Errorf("HttpClient do error %v\n", err)
		return []byte{}, err
	}
	defer resp.Body.Close()
	if resp.Status != HTTPOK {
		return []byte{}, errors.New("server error: " + resp.Status)
	}
	// FIXME: add more resp.Status checks.
	return ioutil.ReadAll(resp.Body)
}

// GetCall make a GET call to targetURL with query.
func GetCall(targetURL string, query url.Values) ([]byte, error) {
	req, err := MakeRequestToken(targetURL, "GET", nil, "", query)
	if err != nil {
		return []byte{}, err
	}
	return GetResponse(req)
}

// PostCall make a POST call to targetURL with a json body.
func PostCall(targetURL string, jsonString []byte) ([]byte, error) {
	req, err := MakeRequestToken(targetURL, "POST", bytes.NewBuffer(jsonString), "", nil)
	if err != nil {
		return []byte{}, err
	}
	return GetResponse(req)
}

// DeleteCall make a DELETE call to targetURL with a json body.
func DeleteCall(targetURL string, jsonString []byte) ([]byte, error) {
	req, err := MakeRequestToken(targetURL, "DELETE", bytes.NewBuffer(jsonString), "", nil)
	if err != nil {
		return []byte{}, err
	}
	return GetResponse(req)
}

// PostFile make a POST call to HTTP server to upload a file.
func PostFile(targetURL string, filename string, query url.Values) ([]byte, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing to buffer: %v\n", err)
		return []byte{}, err
	}
	fh, err := os.Open(filename)
	defer fh.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		return []byte{}, err
	}
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return []byte{}, err
	}

	contentType := bodyWriter.FormDataContentType()
	if err = bodyWriter.Close(); err != nil {
		return []byte{}, err
	}

	req, err := MakeRequestToken(targetURL, "POST", bodyBuf, contentType, query)
	if err != nil {
		return []byte{}, err
	}
	return GetResponse(req)
}

// PostChunk makes a POST call to HTTP server to upload chunkdata.
func PostChunk(targetURL string,
	chunkName string, reader io.Reader, len int64, boundary string) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.SetBoundary(boundary); err != nil {
		return nil, err
	}

	log.V(4).Infoln(chunkName)
	part, err := writer.CreateFormFile("chunk", chunkName)
	if err != nil {
		return nil, err
	}

	_, err = io.CopyN(part, reader, len)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	contentType := writer.FormDataContentType()

	req, err := MakeRequestToken(targetURL, "POST", body, contentType, nil)
	if err != nil {
		return nil, err
	}

	return GetResponse(req)
}

// GetChunk makes a GET call to HTTP server to download chunk data.
func GetChunk(targetURL string,
	query url.Values) (*http.Response, error) {
	req, err := MakeRequestToken(targetURL, "GET", nil, "", query)
	if err != nil {
		return nil, err
	}

	return HttpClient.Do(req)
}
