package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/golang/glog"

	yaml "gopkg.in/yaml.v2"
)

type submitConfigDataCenter struct {
	Active   bool   `yaml:"active"`
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Usercert string `yaml:"usercert"`
	Userkey  string `yaml:"userkey"`
	Endpoint string `yaml:"endpoint"`
}

// Configuration load from user config yaml files
type submitConfig struct {
	DC           []submitConfigDataCenter `yaml:"datacenters"`
	ActiveConfig *submitConfigDataCenter
}

// UserHomeDir get user home dierctory on different platforms
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func token() (string, error) {
	tokenbytes, err := ioutil.ReadFile(UserHomeDir() + "/.paddle/token_cache")
	if err != nil {
		fmt.Fprintf(os.Stderr, "previous token not found, fetching a new one...")
		// Authenticate to the cloud endpoint
		authJSON := map[string]string{}
		authJSON["username"] = config.ActiveConfig.Username
		authJSON["password"] = config.ActiveConfig.Password
		authStr, _ := json.Marshal(authJSON)
		body, err := postCall(authStr, config.ActiveConfig.Endpoint+"/api-token-auth/", "")
		if err != nil {
			return "", err
		}
		var respObj interface{}
		if errJSON := json.Unmarshal(body, &respObj); errJSON != nil {
			return "", errJSON
		}
		tokenStr := respObj.(map[string]interface{})["token"].(string)
		err = ioutil.WriteFile(UserHomeDir()+"/.paddle/token_cache", []byte(tokenStr), 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "write cache token file error: %v", err)
		}
		// Ignore write token error, fetch a new one next time
		return tokenStr, nil
	}
	return string(tokenbytes), nil
}

func parseConfig(configFile string) *submitConfig {
	// ------------------- load paddle config -------------------
	buf, err := ioutil.ReadFile(configFile)
	config := submitConfig{}
	if err == nil {
		yamlErr := yaml.Unmarshal(buf, &config)
		if yamlErr != nil {
			glog.Fatalf("load config %s error: %v", configFile, err)
		}
		// put active config
		for _, item := range config.DC {
			if item.Active {
				config.ActiveConfig = &item
			}
		}
		return &config
	}
	glog.Fatalf("config %s error: %v", configFile, err)
	return nil
}

var config = parseConfig(UserHomeDir() + "/.paddle/config")
