package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/PaddlePaddle/cloud/go/controller"
	k8scontroller "github.com/PaddlePaddle/cloud/go/controller/k8s"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(*kubeconfig)
	if err != nil {
		panic(err)
	}

	// setup some optional configuration
	controller.ConfigureClient(config)

	// start a controller on instances of our custom resource
	controller, err := k8scontroller.NewController(config)
	if err != nil {
		panic(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go controller.Run(ctx)

	for {
		time.Sleep(time.Second)
		fmt.Println("tick.")
	}
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
