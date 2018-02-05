package main

import (
	"context"
	"flag"

	log "github.com/inconshreveable/log15"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/PaddlePaddle/cloud/go/edl"
	edlresource "github.com/PaddlePaddle/cloud/go/edl/resource"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	logLevel := flag.String("log_level", "info", "Log level can be debug, info, warn, error, crit.")
	maxLoadDesired := flag.Float64("max_load_desired", 0.97, `Keep the cluster max resource usage around
		this value, jobs will scale down if total request is over this level.`)
	flag.Parse()

	lvl, err := log.LvlFromString(*logLevel)
	if err != nil {
		panic(err)
	}

	log.Root().SetHandler(
		log.LvlFilterHandler(lvl, log.CallerStackHandler("%+v", log.StderrHandler)),
	)

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	var cfg *rest.Config
	if *kubeconfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(err)
	}

	// setup some optional configuration
	edlresource.RegisterTrainingJob(cfg)

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	client, err := rest.RESTClientFor(cfg)
	if err != nil {
		panic(err)
	}

	controller, err := edl.New(client, clientset, *maxLoadDesired)
	if err != nil {
		panic(err)
	}

	controller.Run(context.Background())
}
