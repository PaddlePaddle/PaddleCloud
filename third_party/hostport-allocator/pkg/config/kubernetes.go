package config

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime/schema"
	kube_rest "k8s.io/client-go/rest"
	kube_client_cmd "k8s.io/client-go/tools/clientcmd"
	kube_client_cmd_api "k8s.io/client-go/tools/clientcmd/api"
)

const (
	// APIVersion to be used.
	APIVersion = "v1"

	defaultUseServiceAccount  = false
	defaultServiceAccountFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultInClusterConfig    = true
)

func getConfigOverrides(uri *url.URL) (*kube_client_cmd.ConfigOverrides, error) {
	kubeConfigOverride := kube_client_cmd.ConfigOverrides{
		ClusterInfo: kube_client_cmd_api.Cluster{
			APIVersion: APIVersion,
		},
	}
	if len(uri.Scheme) != 0 && len(uri.Host) != 0 {
		kubeConfigOverride.ClusterInfo.Server = fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)
	}

	opts := uri.Query()

	if len(opts["apiVersion"]) >= 1 {
		kubeConfigOverride.ClusterInfo.APIVersion = opts["apiVersion"][0]
	}

	if len(opts["insecure"]) > 0 {
		insecure, err := strconv.ParseBool(opts["insecure"][0])
		if err != nil {
			return nil, err
		}
		kubeConfigOverride.ClusterInfo.InsecureSkipTLSVerify = insecure
	}

	return &kubeConfigOverride, nil
}

// GetKubeClientConfig returns rest client configuration based on the passed url.
func GetKubeClientConfig(uri *url.URL) (*kube_rest.Config, error) {
	var (
		kubeConfig *kube_rest.Config
		err        error
	)

	opts := uri.Query()
	configOverrides, err := getConfigOverrides(uri)
	if err != nil {
		return nil, err
	}

	inClusterConfig := defaultInClusterConfig
	if len(opts["inClusterConfig"]) > 0 {
		inClusterConfig, err = strconv.ParseBool(opts["inClusterConfig"][0])
		if err != nil {
			return nil, err
		}
	}

	if inClusterConfig {
		kubeConfig, err = kube_rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		if configOverrides.ClusterInfo.Server != "" {
			kubeConfig.Host = configOverrides.ClusterInfo.Server
		}
		kubeConfig.GroupVersion = &schema.GroupVersion{Version: configOverrides.ClusterInfo.APIVersion}
		kubeConfig.Insecure = configOverrides.ClusterInfo.InsecureSkipTLSVerify
		if configOverrides.ClusterInfo.InsecureSkipTLSVerify {
			kubeConfig.TLSClientConfig.CAFile = ""
		}
	} else {
		authFile := ""
		if len(opts["auth"]) > 0 {
			authFile = opts["auth"][0]
		}

		if authFile != "" {
			if kubeConfig, err = kube_client_cmd.NewNonInteractiveDeferredLoadingClientConfig(
				&kube_client_cmd.ClientConfigLoadingRules{ExplicitPath: authFile},
				configOverrides).ClientConfig(); err != nil {
				return nil, err
			}
		} else {
			kubeConfig = &kube_rest.Config{
				Host: configOverrides.ClusterInfo.Server,
				TLSClientConfig: kube_rest.TLSClientConfig{
					Insecure: configOverrides.ClusterInfo.InsecureSkipTLSVerify,
				},
			}
			kubeConfig.GroupVersion = &schema.GroupVersion{Version: configOverrides.ClusterInfo.APIVersion}
		}
	}
	if len(kubeConfig.Host) == 0 {
		return nil, fmt.Errorf("invalid kubernetes url specified")
	}
	if len(kubeConfig.GroupVersion.Version) == 0 {
		return nil, fmt.Errorf("invalid kubernetes API version specified")
	}

	useServiceAccount := defaultUseServiceAccount
	if len(opts["useServiceAccount"]) >= 1 {
		useServiceAccount, err = strconv.ParseBool(opts["useServiceAccount"][0])
		if err != nil {
			return nil, err
		}
	}

	if useServiceAccount {
		// If a readable service account token exists, then use it
		if contents, err := ioutil.ReadFile(defaultServiceAccountFile); err == nil {
			kubeConfig.BearerToken = string(contents)
		}
	}

	return kubeConfig, nil
}
