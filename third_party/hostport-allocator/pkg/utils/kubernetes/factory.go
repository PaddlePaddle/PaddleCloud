package kubernetes

import (
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes/fake"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	clientv1 "k8s.io/client-go/pkg/api/v1"
	kube_record "k8s.io/client-go/tools/record"

	"k8s.io/client-go/kubernetes"
)

// CreateEventRecorder creates an event recorder to send custom events to Kubernetes to be recorded for targeted Kubernetes objects
func CreateEventRecorder(kubeClient kubernetes.Interface) kube_record.EventRecorder {
	eventBroadcaster := kube_record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)

	if _, isfake := kubeClient.(*fake.Clientset); !isfake {
		eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubeClient.Core().RESTClient()).Events("")})
	}
	return eventBroadcaster.NewRecorder(api.Scheme, clientv1.EventSource{Component: "hostport-manager"})
}
