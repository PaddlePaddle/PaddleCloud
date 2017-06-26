package config

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

func TestConfigParse(t *testing.T) {
	port := 8000
	sampleConfig := `current-datacenter: dc1
datacenters:
- name: dc1
  username: testuser
  password: 123123
  endpoint: http://127.0.0.1:` + strconv.Itoa(port) + `
- name: dc2
  username: testuser2
  password: 123123
  endpoint: http://abc.com:8448`

	tmpfile, err := ioutil.TempFile("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up temp file
	if _, err := tmpfile.Write([]byte(sampleConfig)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	tempconfig := ParseConfig(tmpfile.Name())
	if tempconfig.ActiveConfig.Endpoint != "http://127.0.0.1:"+strconv.Itoa(port) {
		t.Error("config parse error")
	}
}

func TestErrorConfigParse(t *testing.T) {
	sampleErrorConfig := `current-datacenter: dc2
datacenters:
- name: dc1
  username:,, testuser
      password123123
  endpoint: http://cloud.paddlepaddle.org`

	tmpfile, err := ioutil.TempFile("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up temp file
	if _, err := tmpfile.Write([]byte(sampleErrorConfig)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	tempconfig := ParseConfig(tmpfile.Name())
	if tempconfig != nil {
		t.Error("config error not return nil")
	}
}

func TestNonExistFile(t *testing.T) {
	tempconfig := ParseConfig("/path/to/non/exist/file")
	if tempconfig != nil {
		t.Error("non exist file should return nil")
	}
}
