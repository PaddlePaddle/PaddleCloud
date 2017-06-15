package utils

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func fakeServer() *http.Server {
	http.HandleFunc("/api-token-auth/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"token\": \"testtokenvalue\"}"))
	})
	http.HandleFunc("/fake-api/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fakeresult"))
	})
	srv := &http.Server{Addr: ":8980"}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			return
		}
	}()
	return srv
}

func TestConfigParse(t *testing.T) {
	sampleConfig := `current-datacenter: dc1
datacenters:
- name: dc1
  username: testuser
  password: 123123
  endpoint: http://127.0.0.1:8980
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
	tempconfig := parseConfig(tmpfile.Name())
	if tempconfig.ActiveConfig.Endpoint != "http://127.0.0.1:8980" {
		t.Error("config parse error")
	}

	// test token fetching
	srv := fakeServer()
	defer srv.Shutdown(nil)
	Config = tempconfig
	os.Remove(filepath.Join(UserHomeDir(), ".paddle", "token_cache"))
	token, err := token()
	if err != nil {
		t.Errorf("get token error %v", err)
	}
	if token != "testtokenvalue" {
		t.Error("token not equal to the server: (" + token + ")")
	}

	// FIXME: separate these tests
	// test token request
	req, err := MakeRequestToken(Config.ActiveConfig.Endpoint+"/fake-api/", "GET", nil, "", nil)
	if err != nil {
		t.Errorf("make request error %v", err)
	}
	resp, err := GetResponse(req)
	if err != nil {
		t.Errorf("get request error %v", err)
	}
	if string(resp) != "fakeresult" {
		t.Error("error result fetched")
	}

	// test GetCall
	resp, err = GetCall(Config.ActiveConfig.Endpoint+"/fake-api/", nil)
	if err != nil {
		t.Errorf("GetCall error : %v", err)
	}
	if string(resp) != "fakeresult" {
		t.Error("GetCall result error")
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
	tempconfig := parseConfig(tmpfile.Name())
	if tempconfig != nil {
		t.Error("config error not return nil")
	}
}

func TestNonExistFile(t *testing.T) {
	tempconfig := parseConfig("/path/to/non/exist/file")
	if tempconfig != nil {
		t.Error("non exist file should return nil")
	}
}
