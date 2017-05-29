package paddlecloud

import (
	"bytes"
	//"crypto/x509"
	//"encoding/json"
	"errors"
	"fmt"
	"github.com/PaddlePaddle/cloud/go/filemanager/pfsmod"
	//"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	//"os"
	//"strconv"
	log "github.com/golang/glog"
)

type PfsSubmitter struct {
	config *submitConfig
	client *http.Client
	port   int32
}

func NewPfsCmdSubmitter(configFile string) *PfsSubmitter {
	/*
		config, err := getConfig(configFile)
		if err != nil {
			log.Fatal("LoadX509KeyPair:", err)
		}
	*/

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

	log.V(1).Infof("%#v\n", config)

	//http
	client := &http.Client{}

	return &PfsSubmitter{
		config: config,
		client: client,
		port:   8080,
	}
}

func (s *PfsSubmitter) PostFiles(cmd pfsmod.Command) ([]byte, error) {
	jsonString, err := cmd.ToJson()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s:%d/api/v1/files", s.config.ActiveConfig.Endpoint, s.port)
	log.V(1).Info("target url: " + url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonString))
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

	if resp.Status != HTTPOK {
		return nil, errors.New("http server returned non-200 status: " + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *PfsSubmitter) get(cmd pfsmod.Command, path string) ([]byte, error) {
	url := fmt.Sprintf("%s:%d/%s?%s",
		s.config.ActiveConfig.Endpoint,
		s.port,
		path,
		cmd.ToUrlParam())
	log.V(1).Info("target url: " + url)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status != HTTPOK {
		return nil, errors.New("http server returned non-200 status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *PfsSubmitter) GetFiles(cmd pfsmod.Command) ([]byte, error) {
	return s.get(cmd, "api/v1/files")
}
func (s *PfsSubmitter) GetChunkMeta(cmd pfsmod.Command) ([]byte, error) {
	return s.get(cmd, "api/v1/chunks")
}

func (s *PfsSubmitter) GetChunkData(cmd *pfsmod.ChunkCmd) error {
	url := fmt.Sprintf("%s:%d/api/v1/storage/chunks?%s",
		s.config.ActiveConfig.Endpoint,
		s.port,
		cmd.ToUrlParam())

	log.V(1).Info("target url: " + url)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.Status != HTTPOK {
		return errors.New("http server returned non-200 status: " + resp.Status)
	}

	partReader := multipart.NewReader(resp.Body, pfsmod.DefaultMultiPartBoundary)
	for {
		part, error := partReader.NextPart()
		if error == io.EOF {
			break
		}

		if part.FormName() == "chunk" {
			recvCmd, status := pfsmod.NewChunkCmdFromUrlParam(part.FileName())
			if err != nil {
				return errors.New(http.StatusText(status))
			}

			if err := recvCmd.GetChunkData(part); err != nil {
				return err
			}
		}
	}

	return nil
}

func newPostChunkDataRequest(cmd *pfsmod.ChunkCmd, url string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.SetBoundary(pfsmod.DefaultMultiPartBoundary)

	fileName := cmd.ToUrlParam()
	part, err := writer.CreateFormFile("chunk", fileName)
	if err != nil {
		return nil, err
	}

	if err := cmd.WriteChunkData(part); err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func (s *PfsSubmitter) PostChunkData(cmd *pfsmod.ChunkCmd) ([]byte, error) {
	url := fmt.Sprintf("%s:%d/api/v1/storage/chunks",
		s.config.ActiveConfig.Endpoint, s.port)
	log.V(1).Info("target url: " + url)

	req, err := newPostChunkDataRequest(cmd, url)
	if err != nil {
		return nil, nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Status != HTTPOK {
		return nil, errors.New("http server returned non-200 status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}
