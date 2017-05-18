package pfsmodules

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	//pfsmod "github.com/cloud/go/file_manager/pfsmodules"
	log "github.com/golang/glog"
	"github.com/google/subcommands"
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

func NewCmdSubmitter(configFile string) *cmdSubmitter {
	config, err := getConfig(configFile)
	if err != nil {
		log.Fatal("LoadX509KeyPair:", err)
	}

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

	return &cmdSubmitter{
		//cmd:    pfscmd,
		config: config,
		client: client}
}

func (s *cmdSubmitter) SubmitCmdReqeust(
	cmd pfsmod.Cmd,
	//targetURL string,
	httpMethod string) ([]byte, error) {

	jsonString, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	targetURL := s.config.ActiveConfig.Endpoint
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

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	return body, err
}
