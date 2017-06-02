package paddlecloud

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
)

// HTTPOK is ok status of http api call.
const HTTPOK = "200 OK"

type RestClient struct {
	client *http.Client
}

// NewRestClient returns a new RestClient struct.
func NewRestClient() *RestClient {
	client := http.Client{Transport: &http.Transport{}}
	return &RestClient{client: &client}
}

func makeRequest(uri string, method string, body io.Reader,
	contentType string, query url.Values,
	authHeader map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
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

// makeRequestToken use client token to make a authorized request.
func makeRequestToken(uri string, method string, body io.Reader,
	contentType string, query url.Values) (*http.Request, error) {
	// get client token
	token, err := token()
	if err != nil {
		return nil, err
	}
	authHeader := make(map[string]string)
	authHeader["Authorization"] = "Token " + token
	return makeRequest(uri, method, body, contentType, query, authHeader)
}

// NOTE: add other request makers if we need other auth methods.

func (p *RestClient) getResponse(req *http.Request) ([]byte, error) {
	resp, err := p.client.Do(req)
	if err != nil {
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
func (p *RestClient) GetCall(targetURL string, query url.Values) ([]byte, error) {
	req, err := makeRequestToken(targetURL, "GET", nil, "", query)
	if err != nil {
		return []byte{}, err
	}
	return p.getResponse(req)
}

// PostCall make a POST call to targetURL with a json body.
func (p *RestClient) PostCall(targetURL string, jsonString []byte) ([]byte, error) {
	req, err := makeRequestToken(targetURL, "POST", bytes.NewBuffer(jsonString), "", nil)
	if err != nil {
		return []byte{}, err
	}
	return p.getResponse(req)
}

// DeleteCall make a DELETE call to targetURL with a json body.
func (p *RestClient) DeleteCall(targetURL string, jsonString []byte) ([]byte, error) {
	req, err := makeRequestToken(targetURL, "DELETE", bytes.NewBuffer(jsonString), "", nil)
	if err != nil {
		return []byte{}, err
	}
	return p.getResponse(req)
}

// PostFile make a POST call to HTTP server to upload a file.
func (p *RestClient) PostFile(targetURL string, filename string) ([]byte, error) {
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
	return p.getResponse(req)
}

// PostChunkData makes a POST call to HTTP server to upload chunkdata.
func (p *RestClient) PostChunk(targetURL string,
	chunkName string, reader io.Reader, len int64, boundary string) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.SetBoundary(boundary); err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("chunk", chunkName)
	if err != nil {
		return nil, err
	}

	_, err = io.CopyN(part, reader, len)
	if err != nil {
		return nil, err
	}

	contentType := writer.FormDataContentType()
	writer.Close()

	req, err := makeRequestToken(targetURL, "POST", body, contentType, nil)
	if err != nil {
		return []byte{}, err
	}

	return p.getResponse(req)
}

// GetChunkData makes a GET call to HTTP server to download chunk data.
func (p *RestClient) GetChunk(targetURL string,
	query url.Values) (*http.Response, error) {
	req, err := makeRequestToken(targetURL, "GET", nil, "", query)
	if err != nil {
		return nil, err
	}

	return p.client.Do(req)
}

var DefaultClient = NewRestClient()

// GetCall makes a GET call to targetURL with k-v params of query.
func GetCall(targetURL string, query map[string]string) ([]byte, error) {
	q := url.Values{}
	for k, v := range query {
		q.Add(k, v)
	}

	return DefaultClient.GetCall(targetURL, q)
}

// PostCall makes a POST call to targetURL with a json body.
func PostCall(targetURL string, jsonString []byte) ([]byte, error) {
	return DefaultClient.PostCall(targetURL, jsonString)
}

// DeleteCall makes a DELETE call to targetURL with a json body.
func DeleteCall(targetURL string, jsonString []byte) ([]byte, error) {
	return DefaultClient.DeleteCall(targetURL, jsonString)
}
