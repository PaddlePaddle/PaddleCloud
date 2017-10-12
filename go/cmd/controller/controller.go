package main

import (
	"context"
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	"github.com/PaddlePaddle/cloud/go/controller"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	loglevel := flag.String("log_level", "info", "Log level can be debug, info, warn, error, fatal, panic.")
	flag.Parse()

	level, err := log.ParseLevel(*loglevel)
	if err != nil {
		panic(err)
	}
	log.SetLevel(level)

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(*kubeconfig)
	if err != nil {
		panic(err)
	}

	// setup some optional configuration
	paddlejob.ConfigureClient(config)

	// start a controller on instances of our custom resource
	informer, err := controller.NewController(config)
	if err != nil {
		panic(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go informer.Run(ctx)

	for {
		time.Sleep(time.Second)
	}
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
