package main

import (
	"flag"
	"time"

	"github.com/golang/glog"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	paddleclientset "github.com/PaddlePaddle/cloud/go/pkg/client/clientset/versioned"
	paddleinformers "github.com/PaddlePaddle/cloud/go/pkg/client/informers/externalversions"
	paddlecontroller "github.com/PaddlePaddle/cloud/go/pkg/controller"
	"github.com/PaddlePaddle/cloud/go/pkg/signals"
)

func init() {

}

func main() {
	masterURL := flag.String("master", "", "Address of a kube master.")
	kubeConfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeConfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	extapiClient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes extension api clientset: %s", err.Error())
	}

	paddleClient, err := paddleclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building paddle clientset: %s", err.Error())
	}

	paddleInformer := paddleinformers.NewSharedInformerFactory(paddleClient, time.Second*10)

	controller := paddlecontroller.New(kubeClient, extapiClient, paddleClient, paddleInformer)

	go paddleInformer.Start(stopCh)

	if controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running paddle trainingjob controller: %s", err.Error())
	}
}
