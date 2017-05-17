package main

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

func getCall(targetURL string, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 {
		req.Header.Set("Authorization", "Token "+token)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	if resp.Status != HTTPOK {
		return []byte{}, errors.New("http server returned non-200 status: " + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func postCall(jsonString []byte, targetURL string, token string) ([]byte, error) {
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonString))
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 {
		req.Header.Set("Authorization", "Token "+token)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.Status != HTTPOK {
		return []byte{}, errors.New("http server returned non-200 status: " + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s\n\n", body)
	return body, err
}

func deleteCall(jsonString []byte, targetURL string, token string) ([]byte, error) {
	req, err := http.NewRequest("DELETE", targetURL, bytes.NewBuffer(jsonString))
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 {
		req.Header.Set("Authorization", "Token "+token)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.Status != HTTPOK {
		return []byte{}, errors.New("http server returned non-200 status: " + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s\n\n", body)
	return body, err
}

func postFile(filename string, targetURL string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing to buffer: %v\n", err)
		return err
	}

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		return err
	}

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(targetURL, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(respBody))
	return nil
}
