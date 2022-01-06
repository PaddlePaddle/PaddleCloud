// Copyright 2021 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	clientv3 "go.etcd.io/etcd/client/v3"
	volcano "volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	batchv1 "github.com/paddleflow/paddle-operator/api/v1"
	"github.com/paddleflow/paddle-operator/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(batchv1.AddToScheme(scheme))
	utilruntime.Must(volcano.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var namespace string
	var enableLeaderElection bool
	var scheduling string
	var initImage string
	var probeAddr string
	var hostPortRange string
	var etcdServer string
	flag.StringVar(&etcdServer, "etcd-server", "", "The etcd server endpoints.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&namespace, "namespace", "", "The namespace the controller binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&hostPortRange, "port-range", "35000,65000", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&scheduling, "scheduling", "", "The scheduler to take, e.g. volcano")
	flag.StringVar(&initImage, "initImage", "docker.io/library/busybox:1", "The image used for init container, default to busybox")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Namespace:              namespace,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "b2a304f2.paddlepaddle.org",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	hostPosts := strings.Split(hostPortRange, ",")
	portStart, err := strconv.Atoi(hostPosts[0])
	if err != nil {
		setupLog.Error(err, "port should have int type")
		os.Exit(1)
	}

	portEnd, err := strconv.Atoi(hostPosts[1])
	if err != nil {
		setupLog.Error(err, "port should have int type")
		os.Exit(1)
	}

	var etcdCli *clientv3.Client
	if etcdServer != "" {
		etcdEndpoints := strings.Split(etcdServer, ",")
		etcdCli, err = clientv3.New(clientv3.Config{
			Endpoints:   etcdEndpoints,
			DialTimeout: 2 * time.Second,
		})
		if err != nil {
			setupLog.Error(err, "etcd connect failed")
			os.Exit(1)
		}
	}

	// restClient for exec
	gvk := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	}
	restClient, err := apiutil.RESTClientForGVK(gvk, false, mgr.GetConfig(), serializer.NewCodecFactory(mgr.GetScheme()))
	if err != nil {
		setupLog.Error(err, "unable to create REST client")
	}

	if err = (&controllers.PaddleJobReconciler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("PaddleJob"),
		Scheme:     mgr.GetScheme(),
		Recorder:   mgr.GetEventRecorderFor("paddlejob-controller"),
		RESTClient: restClient,
		RESTConfig: mgr.GetConfig(),
		Scheduling: scheduling,
		InitImage:  initImage,
		HostPortMap: map[string]int{
			controllers.HOST_PORT_START: portStart,
			controllers.HOST_PORT_CUR:   portStart,
			controllers.HOST_PORT_END:   portEnd,
		},
		EtcdCli: etcdCli,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PaddleJob")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
