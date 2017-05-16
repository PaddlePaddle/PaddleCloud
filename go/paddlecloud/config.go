package main

type submitConfigDataCenter struct {
	Active   bool   `yaml:"active"`
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Usercert string `yaml:"usercert"`
	Userkey  string `yaml:"userkey"`
	Endpoint string `yaml:"endpoint"`
}

// Configuration load from user config yaml files
type submitConfig struct {
	DC           []submitConfigDataCenter `yaml:"datacenters"`
	ActiveConfig *submitConfigDataCenter
}
