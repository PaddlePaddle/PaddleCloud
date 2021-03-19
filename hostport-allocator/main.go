package main

import (
	"flag"
	"net/url"
	"time"

	"hostport-manager/pkg/config"
	"hostport-manager/pkg/core"
	"hostport-manager/pkg/signals"

	paddle "github.com/PaddlePaddle/cloud/go/api"
	"github.com/golang/glog"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	kube_flag "k8s.io/apiserver/pkg/util/flag"
	kubeinformers "k8s.io/client-go/informers"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// DefaultServiceNodePortRange is the default port range for NodePort services.
	DefaultServiceNodePortRange = utilnet.PortRange{Base: 30000, Size: 2768}
	hostPortRange               utilnet.PortRange

	address            = flag.String("address", ":8085", "The address to expose prometheus metrics.")
	kubeConfigFile     = flag.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")
	kubernetes         = flag.String("kubernetes", "", "Kuberentes master location. Leave blank for default")
	useServiceNodePort = flag.Bool("use-service-nodeport", true, "If true, will create a service with nodeport and need to stop kube-proxy first.")
)

func main() {
	flag.Var(&hostPortRange, "hostport-range", "A port range to reserve for hostport, Example: '30000-32767'.")

	kube_flag.InitFlags()
	glog.Infof("Hostport Manager %s", "0.0.1")
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	kubeClient, restClient := createKubeClient()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	opts := createHostPortManagerOptions()
	hortportManager := core.NewHostPortManager(opts, kubeClient, restClient, kubeInformerFactory, stopCh)

	go kubeInformerFactory.Start(stopCh)

	if err := hortportManager.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

func createKubeClient() (kube.Interface, *rest.RESTClient) {
	if *kubeConfigFile != "" {
		glog.Infof("Using kubeconfig file: %s", *kubeConfigFile)
		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", *kubeConfigFile)
		if err != nil {
			glog.Fatalf("Failed to build config: %v", err)
		}
		paddle.ConfigureClient(config)
		clientset, err := kube.NewForConfig(config)
		if err != nil {
			glog.Fatalf("Create clientset error: %v", err)
		}
		client, err := rest.RESTClientFor(config)
		if err != nil {
			glog.Fatalf("Create rest clientset error: %v", err)
		}
		return clientset, client
	}
	url, err := url.Parse(*kubernetes)
	if err != nil {
		glog.Fatalf("Failed to parse Kuberentes url: %v", err)
	}

	kubeConfig, err := config.GetKubeClientConfig(url)
	if err != nil {
		glog.Fatalf("Failed to build Kuberentes client configuration: %v", err)
	}
	paddle.ConfigureClient(kubeConfig)
	client, err := rest.RESTClientFor(kubeConfig)
	if err != nil {
		panic(err)
	}
	return kube.NewForConfigOrDie(kubeConfig), client
}

func createHostPortManagerOptions() core.HostPortManagerOptions {
	if hostPortRange.Size == 0 {
		hostPortRange = DefaultServiceNodePortRange
	}
	autoscalingOpts := core.HostPortManagerOptions{
		HostNodePortRange:  hostPortRange,
		UseServiceNodePort: *useServiceNodePort,
	}
	return autoscalingOpts
}
