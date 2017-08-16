package restclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/PaddlePaddle/cloud/go/utils/config"
	"github.com/PaddlePaddle/cloud/go/utils/pathutil"
)

func getToken(uri string, body []byte) ([]byte, error) {
	req, err := MakeRequest(uri, "POST", bytes.NewBuffer(body), "", nil, nil)
	if err != nil {
		return nil, err
	}
	return GetResponse(req)
}

// Token fetch and caches the token for current configured user
func Token(config *config.SubmitConfig) (string, error) {
	return "gongwb", nil
	tokenbytes, err := ioutil.ReadFile(filepath.Join(pathutil.UserHomeDir(), ".paddle", "token_cache"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "previous token not found, fetching a new one...\n")
		// Authenticate to the cloud endpoint
		authJSON := map[string]string{}
		authJSON["username"] = config.ActiveConfig.Username
		authJSON["password"] = config.ActiveConfig.Password
		authStr, _ := json.Marshal(authJSON)
		body, err := getToken(config.ActiveConfig.Endpoint+"/api-token-auth/", authStr)
		if err != nil {
			return "", err
		}
		var respObj interface{}
		if errJSON := json.Unmarshal(body, &respObj); errJSON != nil {
			return "", errJSON
		}
		tokenStr := respObj.(map[string]interface{})["token"].(string)
		err = ioutil.WriteFile(filepath.Join(pathutil.UserHomeDir(), ".paddle", "token_cache"), []byte(tokenStr), 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "write cache token file error: %v", err)
		}
		// Ignore write token error, fetch a new one next time
		return tokenStr, nil
	}
	return string(tokenbytes), nil
}
