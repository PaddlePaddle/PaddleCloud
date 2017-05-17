package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	pfsmod "github.com/cloud/go/file_manager/pfsmodules"
	log "github.com/golang/glog"
	"github.com/google/subcommands"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

type lsCmd struct {
	r bool
}

func (*lsCmd) Name() string     { return "ls" }
func (*lsCmd) Synopsis() string { return "List files on PaddlePaddle Cloud" }
func (*lsCmd) Usage() string {
	return `ls [-r] <pfspath>:
	List files on PaddlePaddleCloud
	Options:
`
}

func (p *lsCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.r, "r", false, "list files recursively")
}

func (p *lsCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		f.Usage()
		return subcommands.ExitFailure
	}

	cmd := NewCmd(p.Name(), f)

	s := NewCmdSubmitter(UserHomeDir() + "/.paddle/config")
	/*
		if err != nil {
			fmt.Printf("error NewPfsCommand: %v\n", err)
			return subcommands.ExitFailure
		}
	*/

	body, err := s.SubmitCmdReqeust(*cmd, "GET")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return subcommands.ExitFailure
	}

	fmt.Println(body)

	return subcommands.ExitSuccess
}

func NewCmd(cmdName string, f *flag.FlagSet) *pfsmod.Cmd {
	cmd := pfsmod.Cmd{}

	cmd.Method = cmdName
	cmd.Options = make([]pfsmod.Option, f.NFlag())
	cmd.Args = make([]string, f.NArg())

	f.Visit(func(flag *flag.Flag) {
		option := pfsmod.Option{}
		option.Name = flag.Name
		option.Value = flag.Value.String()

		cmd.Options = append(cmd.Options, option)
	})

	for _, arg := range f.Args() {
		fmt.Printf("%s\n", arg)
		cmd.Args = append(cmd.Args, arg)
	}

	return &cmd
}

// Submitter submit cmd to cloud
type cmdSubmitter struct {
	//cmd    *pfsmod.Cmd
	config *submitConfig
	client *http.Client
}

func GetConfig(file string) (*submitConfig, error) {
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

func NewCmdSubmitter(configFile string) *cmdSubmitter {
	config, err := GetConfig(configFile)
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
