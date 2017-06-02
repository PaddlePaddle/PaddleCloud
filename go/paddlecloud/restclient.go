package paddlecloud

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

// HTTPOK is ok status of http api call
const HTTPOK = "200 OK"

var httpTransport = &http.Transport{}

func makeRequest(uri string, method string, body io.Reader,
	contentType string, query map[string]string,
	authHeader map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}
	// default contentType is application/json
	if len(contentType) == 0 {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range authHeader {
		req.Header.Set(k, v)
	}

	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
}

// makeRequestToken use client token to make a authorized request
func makeRequestToken(uri string, method string, body io.Reader,
	contentType string, query map[string]string) (*http.Request, error) {
	// get client token
	token, err := token()
	if err != nil {
		return nil, err
	}
	authHeader := make(map[string]string)
	authHeader["Authorization"] = "Token " + token
	return makeRequest(uri, method, body, contentType, query, authHeader)
}

// NOTE: add other request makers if we need other auth methods

func getResponse(req *http.Request) ([]byte, error) {
	client := &http.Client{Transport: httpTransport}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	if resp.Status != HTTPOK {
		return []byte{}, errors.New("server error: " + resp.Status)
	}
	// FIXME: add more resp.Status checks
	return ioutil.ReadAll(resp.Body)
}

// GetCall make a GET call to targetURL with k-v params of query
func GetCall(targetURL string, query map[string]string) ([]byte, error) {
	req, err := makeRequestToken(targetURL, "GET", nil, "", query)
	if err != nil {
		return []byte{}, err
	}
	return getResponse(req)
}

// PostCall make a POST call to targetURL with a json body
func PostCall(targetURL string, jsonString []byte) ([]byte, error) {
	req, err := makeRequestToken(targetURL, "POST", bytes.NewBuffer(jsonString), "", nil)
	if err != nil {
		return []byte{}, err
	}
	return getResponse(req)
}

// DeleteCall make a DELETE call to targetURL with a json body
func DeleteCall(targetURL string, jsonString []byte) ([]byte, error) {
	req, err := makeRequestToken(targetURL, "DELETE", bytes.NewBuffer(jsonString), "", nil)
	if err != nil {
		return []byte{}, err
	}
	return getResponse(req)
}

// PostFile make a POST call to HTTP server to upload a file
func PostFile(targetURL string, filename string) ([]byte, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing to buffer: %v\n", err)
		return []byte{}, err
	}

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		return []byte{}, err
	}

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return []byte{}, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	req, err := makeRequestToken(targetURL, "POST", bodyBuf, contentType, nil)
	if err != nil {
		return []byte{}, err
	}
	return getResponse(req)
}
