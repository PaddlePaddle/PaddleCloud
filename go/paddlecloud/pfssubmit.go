package main

import (
	"bytes"
	"fmt"
	//"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	pfsmod "github.com/cloud/go/file_manager/pfsmodules"
	log "github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

func getConfig(file string) (*submitConfig, error) {
	buf, err := ioutil.ReadFile(file)
	config := submitConfig{}
	if err != nil {
		fmt.Printf("error loading config file: %s, %v\n", file, err)
		return nil, err
	}

	if err := yaml.Unmarshal(buf, &config); err != nil {
		return nil, err
	}

	// put active config
	for _, item := range config.DC {
		if item.Active {
			config.ActiveConfig = &item
		}
	}

	fmt.Printf("config: %v\n", config.ActiveConfig)
	return &config, err
}

func loadCA(caFile string) *x509.CertPool {
	pool := x509.NewCertPool()

	if ca, e := ioutil.ReadFile(caFile); e != nil {
		log.Fatal("ReadFile: ", e)
	} else {
		pool.AppendCertsFromPEM(ca)
	}
	return pool
}

// Submitter submit cmd to cloud
type CmdSubmitter struct {
	//cmd    *pfsmod.Cmd
	config *submitConfig
	client *http.Client
}

func NewCmdSubmitter(configFile string) *CmdSubmitter {
	config, err := getConfig(configFile)
	if err != nil {
		log.Fatal("LoadX509KeyPair:", err)
	}

	/*https
	pair, e := tls.LoadX509KeyPair(config.ActiveConfig.Usercert,
		config.ActiveConfig.Userkey)

	if e != nil {
		log.Fatal("LoadX509KeyPair:", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      loadCA(config.ActiveConfig.CAcert),
				Certificates: []tls.Certificate{pair},
			},
		}}
	*/

	//http
	client := &http.Client{}

	return &CmdSubmitter{
		//cmd:    pfscmd,
		config: config,
		client: client}
}

func (s *CmdSubmitter) SubmitCmdReqeust(
	httpMethod string,
	restPath string,
	port uint32,
	cmd pfsmod.Command) ([]byte, error) {

	jsonString, err := json.Marshal(cmd.GetCmdAttr())
	if err != nil {
		return nil, err
	}

	//fmt.Println(jsonString[:len(jsonString)])
	fmt.Println(string(jsonString))

	//targetURL := s.config.ActiveConfig.Endpoint + port + ""
	targetURL := fmt.Sprintf("%s:%d/%s", s.config.ActiveConfig.Endpoint, port, restPath)
	//targetURL := "http://localhost:8080"
	fmt.Println(targetURL)
	req, err := http.NewRequest(httpMethod, targetURL, bytes.NewBuffer(jsonString))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := s.client
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	/*
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
		return body, err
	*/
	if resp.Status != HTTPOK {
		return nil, errors.New("http server returned non-200 status: " + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s\n\n", body)

	cmdResp := cmd.GetResponse()
	if err := json.Unmarshal(body, cmdResp); err != nil {
		cmdResp.SetErr(err.Error())
		return nil, err
	}
	return body, err
}
