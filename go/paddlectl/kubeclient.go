package paddlectl

import (
	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func kubeClient(kubeconfig string) (*rest.RESTClient, error) {
	config, err := buildConfig(kubeconfig)
	if err != nil {
		panic(err)
	}

	// setup some optional configuration
	paddlejob.ConfigureClient(config)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	client, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}

	return client, nil
}
