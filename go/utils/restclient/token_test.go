package restclient

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/PaddlePaddle/cloud/go/utils/config"
	"github.com/PaddlePaddle/cloud/go/utils/pathutil"
	"github.com/stretchr/testify/require"
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
	time.Sleep(50 * time.Millisecond)
	return srv, listener.Addr().(*net.TCPAddr).Port
}

func mkdir_p(path string) error {
	fi, err := os.Stat(path)

	if os.IsExist(err) {
		if !fi.IsDir() {
			return errors.New("exist a same name file")
		}

		return nil
	}

	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}

	return nil
}

func TestTokenParse(t *testing.T) {
	srv, port := fakeServer()
	defer srv.Shutdown(nil)

	// test token fetching
	path := filepath.Join(pathutil.UserHomeDir(), ".paddle")
	require.Nil(t, mkdir_p(path), "mkdir ", path)

	os.Remove(filepath.Join(path, "token_cache"))
	tmpconf := &config.SubmitConfig{ActiveConfig: &config.SubmitConfigDataCenter{
		Name:     "test",
		Username: "testuser",
		Password: "fff",
		Endpoint: fmt.Sprintf("http://127.0.0.1:%d", port),
	}}

	token, err := Token(tmpconf)
	require.Nil(t, err, "get token")
	require.Equal(t, "testtokenvalue", token, "token not equal to the server")

	// FIXME: separate these tests
	// test token request
	uri := fmt.Sprintf("http://127.0.0.1:%d/fake-api/", port)
	req, err := MakeRequest(uri, "GET", nil, "", nil, nil)
	require.Nil(t, err, "make request")
	resp, err := GetResponse(req)
	require.Nil(t, err, "get request")
	require.Equal(t, "fakeresult", string(resp))
}
