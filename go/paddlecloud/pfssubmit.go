package paddlecloud

import (
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	pfsmod "github.com/PaddlePaddle/cloud/go/filemanager/pfsmodules"
	log "github.com/golang/glog"
)

type pfsSubmitter struct {
	config *submitConfig
	client *http.Client
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

// newPfsCmdSubmitter returns a new pfsSubmitter.
func newPfsCmdSubmitter(configFile string) *pfsSubmitter {
	// TODO
	/*https
	pair, err := tls.LoadX509KeyPair(config.ActiveConfig.Usercert,
		config.ActiveConfig.Userkey)
	if err != nil {
		log.Fatalf("LoadX509KeyPair:%v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      loadCA(config.ActiveConfig.CAcert),
				Certificates: []tls.Certificate{pair},
			},
		}}

	log.V(1).Infof("%#v\n", config)
	log.V(3).Infof("%#v\n", client)
	*/

	client := &http.Client{}

	return &pfsSubmitter{
		config: config,
		client: client,
	}
}

// PostFiles implements files' POST method.
func (s *pfsSubmitter) PostFiles(cmd pfsmod.Command) ([]byte, error) {
	jsonString, err := cmd.ToJSON()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/files", s.config.ActiveConfig.Endpoint)
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
	defer pfsmod.Close(resp.Body)

	if resp.Status != HTTPOK {
		return nil, errors.New("http server returned non-200 status: " + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.V(2).Info(string(body[:]))
	return body, nil
}

func (s *pfsSubmitter) get(cmd pfsmod.Command, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s?%s",
		s.config.ActiveConfig.Endpoint,
		path,
		cmd.ToURLParam())
	log.V(1).Info("target url: " + url)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer pfsmod.Close(resp.Body)

	if resp.Status != HTTPOK {
		return nil, errors.New("http server returned non-200 status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// GetFiles implements files's GET method.
func (s *pfsSubmitter) GetFiles(cmd pfsmod.Command) ([]byte, error) {
	return s.get(cmd, "api/v1/files")
}

// GetChunkMeta implements ChunkMeta's GET method.
func (s *pfsSubmitter) GetChunkMeta(cmd pfsmod.Command) ([]byte, error) {
	return s.get(cmd, "api/v1/chunks")
}

// GetChunkData implements Chunks's GET method.
func (s *pfsSubmitter) GetChunkData(cmd *pfsmod.Chunk, dst string) error {
	url := fmt.Sprintf("%s/api/v1/storage/chunks?%s",
		s.config.ActiveConfig.Endpoint,
		cmd.ToURLParam())

	log.V(1).Info("target url: " + url)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer pfsmod.Close(resp.Body)

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
			recvCmd, err := pfsmod.ParseChunk(part.FileName())
			if err != nil {
				return errors.New(err.Error())
			}

			recvCmd.Path = dst

			if err := recvCmd.SaveChunkData(part); err != nil {
				return err
			}
		}
	}

	return nil
}

func getDstParam(src *pfsmod.Chunk, dst string) string {
	cmd := pfsmod.Chunk{
		Path:   dst,
		Offset: src.Offset,
		Size:   src.Size,
	}

	return cmd.ToURLParam()
}

func newPostChunkDataRequest(cmd *pfsmod.Chunk, dst string, url string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.SetBoundary(pfsmod.DefaultMultiPartBoundary); err != nil {
		return nil, err
	}

	fileName := getDstParam(cmd, dst)
	part, err := writer.CreateFormFile("chunk", fileName)
	if err != nil {
		return nil, err
	}

	if err := cmd.LoadChunkData(part); err != nil {
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

// PostChunkData implements chunks's POST method.
func (s *pfsSubmitter) PostChunkData(cmd *pfsmod.Chunk, dst string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/storage/chunks",
		s.config.ActiveConfig.Endpoint)
	log.V(1).Info("target url: " + url)

	req, err := newPostChunkDataRequest(cmd, dst, url)
	if err != nil {
		return nil, nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer pfsmod.Close(resp.Body)

	if resp.Status != HTTPOK {
		return nil, errors.New("http server returned non-200 status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}
