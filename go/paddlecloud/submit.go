package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/google/subcommands"
)

type submitCmd struct {
	Jobpackage string
	Parallism  int
	CPU        int
	GPU        int
	Memory     string
	Pservers   int
	PSCPU      int
	PSMemory   string
	Entry      string
	Topology   string
}

func (*submitCmd) Name() string     { return "submit" }
func (*submitCmd) Synopsis() string { return "Submit job to PaddlePaddle Cloud." }
func (*submitCmd) Usage() string {
	return `submit [options] <package path>:
	Submit job to PaddlePaddle Cloud.
	Options:
`
}

func (p *submitCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&p.Parallism, "parallism", 1, "Number of parrallel trainers. Defaults to 1.")
	f.IntVar(&p.CPU, "cpu", 1, "CPU resource each trainer will use. Defaults to 1.")
	f.IntVar(&p.GPU, "gpu", 0, "GPU resource each trainer will use. Defaults to 0.")
	f.StringVar(&p.Memory, "memory", "1Gi", " Memory resource each trainer will use. Defaults to 1Gi.")
	f.IntVar(&p.Pservers, "pservers", 0, "Number of parameter servers. Defaults equal to -p")
	f.IntVar(&p.PSCPU, "pscpu", 1, "Parameter server CPU resource. Defaults to 1.")
	f.StringVar(&p.PSMemory, "psmemory", "1Gi", "Parameter server momory resource. Defaults to 1Gi.")
	f.StringVar(&p.Entry, "entry", "paddle train", "Command of starting trainer process. Defaults to paddle train")
	f.StringVar(&p.Topology, "topology", "", "Will Be Deprecated .py file contains paddle v1 job configs")
}

func (p *submitCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return subcommands.ExitFailure
	}
	// default pservers count equals to trainers count
	if p.Pservers == 0 {
		p.Pservers = p.Parallism
	}
	p.Jobpackage = f.Arg(0)

	s, err := NewSubmitter(p, "/Users/baidu/.paddle/config")
	if err != nil {
		fmt.Printf("error submitting: %v\n", err)
		return subcommands.ExitFailure
	}
	s.Submit(f.Arg(0))
	return subcommands.ExitSuccess
}

// Submitter submit job to cloud
type Submitter struct {
	args   *submitCmd
	config *submitConfig
}

// NewSubmitter returns a submitter object
func NewSubmitter(cmd *submitCmd, configFile string) (*Submitter, error) {
	buf, err := ioutil.ReadFile(configFile)
	config := submitConfig{}
	if err == nil {
		// load paddle config
		yamlErr := yaml.Unmarshal(buf, &config)
		if yamlErr != nil {
			return nil, yamlErr
		}
		// put active config
		for _, item := range config.DC {
			if item.Active {
				config.ActiveConfig = &item
			}
		}
		s := Submitter{cmd, &config}
		return &s, nil
	}
	fmt.Printf("error loading config file: %s, %v\n", configFile, err)

	return nil, err
}

// Submit current job
func (s *Submitter) Submit(jobPackage string) error {
	// 1. upload user job package to pfs
	filepath.Walk(jobPackage, func(path string, info os.FileInfo, err error) error {
		fmt.Printf("Uploading %s...\n", path)
		return nil
		//return postFile(path, s.config.activeConfig.endpoint+"/api/v1/files")
	})
	// 2. call paddlecloud server to create kubernetes job
	jsonString, err := json.Marshal(s.args)
	if err != nil {
		return err
	}
	fmt.Printf("Submitting job: %s to %s\n", jsonString, s.config.ActiveConfig.Endpoint+"/api/v1/jobs")
	//return postCall(jsonString, s.config.activeConfig.endpoint+"/api/v1/jobs")
	return nil
}

func postCall(jsonString []byte, targetURL string) error {
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonString))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	return err
}

func postFile(filename string, targetURL string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
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
