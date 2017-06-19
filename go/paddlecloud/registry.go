package paddlecloud

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/PaddlePaddle/cloud/go/utils"
	"github.com/golang/glog"
	"github.com/google/subcommands"
)

const (
	// RegistryCmdName is subcommand name
	RegistryCmdName = "registry"
	RegistryPrefix  = "pcloud-registry"
)

// RegistryCmd is Docker registry secret information
type RegistryCmd struct {
	SecretName string `json:"name"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Server     string `json:"server"`
}

// Name is the subcommand name
func (r *RegistryCmd) Name() string { return RegistryCmdName }

// Synopsis is the subcommand's synopsis
func (r *RegistryCmd) Synopsis() string { return "Add registry secret on paddlecloud." }

// Usage is the subcommand's usage
func (r *RegistryCmd) Usage() string {
	return `registry <options> [add|del]:
`
}

// SetFlags registers subcommands flags.
func (r *RegistryCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.SecretName, "name", "", "registry secret name")
	f.StringVar(&r.Username, "username", "", "your Docker registry username")
	f.StringVar(&r.Password, "password", "", "your Docker registry password")
	f.StringVar(&r.Server, "server", "", "your Docker registry Server")
}
func (r *RegistryCmd) addRegistrySecret() error {
	jsonString, err := json.Marshal(r)
	if err != nil {
		return err
	}
	glog.V(10).Infof("Add registry secret: %s to %s\n", jsonString, utils.Config.ActiveConfig.Endpoint+"/api/v1/registry/")
	respBody, err := utils.PostCall(utils.Config.ActiveConfig.Endpoint+"/api/v1/registry/", jsonString)
	if err != nil {
		return err
	}
	var respObj interface{}
	if err = json.Unmarshal(respBody, &respObj); err != nil {
		return err
	}
	// FIXME: Return an error if error message is not empty. Use response code instead
	errMsg := respObj.(map[string]interface{})["msg"].(string)
	if len(errMsg) > 0 {
		return errors.New(errMsg)
	}
	return nil
}

// Delete the specify registry
func (r *RegistryCmd) Delete() error {
	jsonString, err := json.Marshal(r)
	if err != nil {
		return err
	}
	glog.V(10).Infof("Delete registry secret: %s to %s\n", jsonString, utils.Config.ActiveConfig.Endpoint+"/api/v1/registry/")
	respBody, err := utils.DeleteCall(utils.Config.ActiveConfig.Endpoint+"/api/v1/registry/", jsonString)
	if err != nil {
		return err
	}

	var respObj interface{}
	if err = json.Unmarshal(respBody, &respObj); err != nil {
		return err
	}
	// FIXME: Return an error if error message is not empty. Use response code instead
	errMsg := respObj.(map[string]interface{})["msg"].(string)
	if len(errMsg) > 0 {
		return errors.New(errMsg)
	}
	return nil
}
func (r *RegistryCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if r.SecretName == "" || r.Username == "" || r.Password == "" || r.Server == "" {
		f.Usage()
		return subcommands.ExitFailure
	}
	r.SecretName = strings.Join([]string{RegistryPrefix, r.SecretName}, "-")
	err := r.addRegistrySecret()
	if err != nil {
		fmt.Fprintf(os.Stderr, "add registry secret failed: %s\n", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

// KubeRegistryName add a prefix for the name
func KubeRegistryName(name string) string {
	return RegistryPrefix + "-" + name
}

// RegistryName is registry secret name for PaddleCloud
func RegistryName(name string) string {
	if strings.HasPrefix(name, RegistryPrefix) {
		return name[len(RegistryPrefix)+1 : len(name)]
	}
	return ""
}
