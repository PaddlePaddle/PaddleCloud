package restclient

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/PaddlePaddle/cloud/go/utils/config"
	"github.com/PaddlePaddle/cloud/go/utils/pathutil"
)

func fakeServer() (*http.Server, int) {
	http.HandleFunc("/api-token-auth/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"token\": \"testtokenvalue\"}"))
	})
	http.HandleFunc("/fake-api/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fakeresult"))
	})
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	fmt.Println("Using port:", listener.Addr().(*net.TCPAddr).Port)

	srv := &http.Server{Addr: listener.Addr().String()}
	go func() {
		if err := srv.Serve(listener); err != nil {
			return
		}
	}()
	return srv, listener.Addr().(*net.TCPAddr).Port
}

func TestTokenParse(t *testing.T) {
	srv, port := fakeServer()
	defer srv.Shutdown(nil)

	// test token fetching
	os.Remove(filepath.Join(pathutil.UserHomeDir(), ".paddle", "token_cache"))
	tmpconf := &config.SubmitConfig{ActiveConfig: &config.SubmitConfigDataCenter{
		Name:     "test",
		Username: "testuser",
		Password: "fff",
		Endpoint: fmt.Sprintf("http://127.0.0.1:%d", port),
	}}

	token, err := Token(tmpconf)
	if err != nil {
		t.Errorf("get token error %v", err)
	}
	if token != "testtokenvalue" {
		t.Error("token not equal to the server: (" + token + ")")
	}

	// FIXME: separate these tests
	// test token request
	uri := fmt.Sprintf("http://127.0.0.1:%d/fake-api/", port)
	req, err := MakeRequest(uri, "GET", nil, "", nil, nil)
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

}
