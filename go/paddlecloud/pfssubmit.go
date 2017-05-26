package paddlecloud

type PfsSubmitter struct {
	config *submitConfig
	client *http.Client
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
		client: client}
}

func (s *PfsSubmitter) GetFiles(c Command) []byte {
}

func (s *PfsSubmitter) PostFiles(c Comand) []byte {
}

func (s *PfsSubmitter) GetChunkMeta(c *ChunkMetaCommand) []byte {
}

func (s *PfsSubmitter) GetChunkData(c *ChunkCommand) {
}

func (s *PfsSubmitter) PostChunkData(c *ChunkCommand) []byte {
}
