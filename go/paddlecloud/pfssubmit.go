package paddlecloud

type PfsSubmitter struct {
	config *submitConfig
	client *http.Client
	port   int32
}

func NewPfsCmdSubmitter(configFile string) *PfsSubmitter {
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
		config: config,
		client: client,
		port:   8080,
	}
}

func (s *PfsSubmitter) PostFiles(cmd Command) ([]byte, error) {
	jsonString, err := c.ToJson()
	if err != nil {
		return nil, err
	}

	targetURL := fmt.Sprintf("%s:%d/%s", s.config.ActiveConfig.Endpoint, port, restPath)
	fmt.Printf("target url:%s\n", targetUrl)

	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonString))
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

func (s *PfsSubmitter) get(cmd Commnad, path string) ([]byte, error) {
	url := fmt.Sprintf("%s:%d/%s?%s",
		s.config.ActiveConfig.Endpoint,
		s.port,
		path,
		cmd.ToUrl())
	fmt.Printf("target url: " + targetURL)

	req, err := http.NewRequest("GET", targetURL, http.NoBody)
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

func (s *PfsSubmitter) GetFiles(cmd Comand) ([]byte, error) {
	return get(cmd, "api/v1/files")
}
func (s *PfsSubmitter) GetChunkMeta(cmd Comand) ([]byte, error) {
	return get(cmd, "api/v1/chunks")
}

func (s *PfsSubmitter) GetChunkData(cmd *pfsmod.ChunkCmd) ([]body, error) {
	baseUrl := fmt.Sprintf("%s:%d/api/v1/storage/chunks?%s",
		s.config.ActiveConfig.Endpoint,
		s.port,
		cmd.ToUrl())

	fmt.Printf("target url: " + targetURL)

	req, err := http.NewRequest("GET", targetURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
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
			chunkCmdAttr, err := pfsmod.ParseFileNameParam(part.FileName())
			recvCmd, err := pfsmod.NewChunkCmd
			if err != nil {
				return nil, err
			}

			f, err := pfsmod.GetChunkWriter(dest, chunkCmdAttr.Offset)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			writen, err := io.Copy(f, part)
			if err != nil || writen != int64(chunkCmdAttr.ChunkSize) {
				fmt.Errorf("read " + strconv.FormatInt(writen, 10) + "error:" + err.Error())
				return err
			}
		}
	}
	return nil
}

func newPostChunkDataRequest(cmd *ChunkCommand) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.SetBoundary(pfsmod.DefaultMultiPartBoundary)

	fileName := cmd.ToUrl()
	part, err := writer.CreateFormFile("chunk", fileName)
	if err != nil {
		return nil, err
	}

	cmd.WriteChunkData(part)

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func (s *PfsSubmitter) PostChunkData(cmd *ChunkCommand) []byte {
	targetUrl := fmt.Sprintf("%s:%d/api/v1/storage/chunks", s.config.ActiveConfig.Endpoint, port)
	fmt.Printf("chunk data target url: " + targetUrl)

	req, err := newPostChunkDataRequest(cmd)
	if err != nil {
		return nil, nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.Status != HTTPOK {
		return errors.New("http server returned non-200 status: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}
