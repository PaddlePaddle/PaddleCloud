package config

import (
	"io/ioutil"
	"net/url"
	"path/filepath"
	"regexp"

	"github.com/PaddlePaddle/cloud/go/utils/pathutil"
	"github.com/golang/glog"

	yaml "gopkg.in/yaml.v2"
)

// SubmitConfigDataCenter is inner conf for paddlecloud
type SubmitConfigDataCenter struct {
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Usercert string `yaml:"usercert"`
	Userkey  string `yaml:"userkey"`
	Endpoint string `yaml:"endpoint"`
}

// SubmitConfig is configuration load from user config yaml files
type SubmitConfig struct {
	DC                []SubmitConfigDataCenter `yaml:"datacenters"`
	ActiveConfig      *SubmitConfigDataCenter
	CurrentDatacenter string `yaml:"current-datacenter"`
}

// DefaultConfigFile returns the path of paddlecloud default config file path
func DefaultConfigFile() string {
	return filepath.Join(pathutil.UserHomeDir(), ".paddle", "config")
}

// ParseDefaultConfig returns default parsed config struct in ~/.paddle/config
func ParseDefaultConfig() *SubmitConfig {
	return ParseConfig(DefaultConfigFile())
}

func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	return true
}

// ParseConfig parse paddlecloud config to a struct
func ParseConfig(configFile string) *SubmitConfig {
	// ------------------- load paddle config -------------------
	buf, err := ioutil.ReadFile(configFile)
	config := SubmitConfig{}
	if err == nil {
		yamlErr := yaml.Unmarshal(buf, &config)
		if yamlErr != nil {
			glog.Errorf("load config %s error: %v\n", configFile, yamlErr)
			return nil
		}

		var re = regexp.MustCompile(`(/|\\)*$`)
		for _, t := range config.DC {
			if !isValidURL(t.Endpoint) {
				glog.Errorf("DC:%v Endpoint:%v is not a valid URL\n", config.DC, t.Endpoint)
				return nil
			}
			t.Endpoint = re.ReplaceAllString(t.Endpoint, "")
		}

		// put active config
		config.ActiveConfig = nil
		for _, item := range config.DC {
			if item.Name == config.CurrentDatacenter {
				config.ActiveConfig = &item
				break
			}
		}
		return &config
	}
	glog.Errorf("config %s error: %v\n", configFile, err)
	return nil
}
