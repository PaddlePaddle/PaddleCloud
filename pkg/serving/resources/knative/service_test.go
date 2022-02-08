package knative

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"

	elasticservingv1 "github.com/paddlepaddle/paddlecloud/pkg/apis/serving/v1"
	"github.com/paddlepaddle/paddlecloud/pkg/serving/constants"
)

const (
	image                      = "hub.baidubce.com/paddlepaddle/serving"
	port                       = 9292
	tag                        = "latest"
	actualTestServiceName      = "paddlesvc"
	paddleServiceDefaultCPU    = "0.1"
	paddleServiceDefaultMemory = "128Mi"
	paddleServiceName          = "paddlesvc"
	paddleServiceNamespace     = "default"
	runtimeVersion             = "latest"
)

var (
	command              = []string{"/bin/bash", "-c"}
	args                 = []string{""}
	containerConcurrency = int64(0)
	timeoutSeconds       = int64(300)

	readinessInitialDelaySeconds = 60
	readinessFailureThreshold    = 3
	readinessPeriodSeconds       = 10
	readinessTimeoutSeconds      = 180
	livenessInitialDelaySeconds  = 60
	livenessFailureThreshold     = 3
	livenessPeriodSeconds        = 10

	defaultTrafficPercent int64 = 50

	setLastRevision bool = false
)

var defaultResources = core.ResourceList{
	core.ResourceCPU:    resource.MustParse(paddleServiceDefaultCPU),
	core.ResourceMemory: resource.MustParse(paddleServiceDefaultMemory),
}

var annotations = map[string]string{
	"autoscaling.knative.dev/class":                       "kpa.autoscaling.knative.dev",
	"autoscaling.knative.dev/maxScale":                    "0",
	"autoscaling.knative.dev/metric":                      "concurrency",
	"autoscaling.knative.dev/minScale":                    "0",
	"autoscaling.knative.dev/panicThresholdPercentage":    "200",
	"autoscaling.knative.dev/panicWindowPercentage":       "10",
	"autoscaling.knative.dev/target":                      "100",
	"autoscaling.knative.dev/targetUtilizationPercentage": "70",
	"autoscaling.knative.dev/window":                      "60s",
}

var paddlesvc = elasticservingv1.PaddleService{
	ObjectMeta: metav1.ObjectMeta{
		Name:      paddleServiceName,
		Namespace: paddleServiceNamespace,
	},
	Spec: elasticservingv1.PaddleServiceSpec{
		RuntimeVersion: runtimeVersion,
		Resources: core.ResourceRequirements{
			Requests: defaultResources,
			Limits:   defaultResources,
		},
		Default: &elasticservingv1.EndpointSpec{
			ContainerImage: image,
			Tag:            tag,
			Port:           port,
		},
	},
}

var paddlesvcCanaryWithSameConfig = elasticservingv1.PaddleService{
	ObjectMeta: metav1.ObjectMeta{
		Name:      paddleServiceName,
		Namespace: paddleServiceNamespace,
	},
	Spec: elasticservingv1.PaddleServiceSpec{
		RuntimeVersion: runtimeVersion,
		Resources: core.ResourceRequirements{
			Requests: defaultResources,
			Limits:   defaultResources,
		},
		Default: &elasticservingv1.EndpointSpec{
			ContainerImage: image,
			Tag:            tag,
			Port:           port,
		},
		Canary: &elasticservingv1.EndpointSpec{
			ContainerImage: image,
			Tag:            tag,
			Port:           port,
		},
	},
}

var defaultService = &knservingv1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      paddlesvc.Name,
		Namespace: paddleServiceNamespace,
	},
	Spec: knservingv1.ServiceSpec{
		ConfigurationSpec: knservingv1.ConfigurationSpec{
			Template: knservingv1.RevisionTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: paddlesvc.Name + "-default",
					Labels: map[string]string{
						"PaddleService": paddlesvc.Name,
					},
					Annotations: annotations,
				},
				Spec: knservingv1.RevisionSpec{
					ContainerConcurrency: &containerConcurrency,
					TimeoutSeconds:       &timeoutSeconds,
					PodSpec: core.PodSpec{
						Containers: []core.Container{
							{
								ImagePullPolicy: core.PullAlways,
								Name:            paddlesvc.Spec.RuntimeVersion,
								Image:           image + ":" + tag,
								Ports: []core.ContainerPort{
									{ContainerPort: port, Name: "http1", Protocol: "TCP"},
								},
								Command: command,
								Args:    args,
								ReadinessProbe: &core.Probe{
									InitialDelaySeconds: int32(readinessInitialDelaySeconds),
									FailureThreshold:    int32(readinessFailureThreshold),
									PeriodSeconds:       int32(readinessPeriodSeconds),
									TimeoutSeconds:      int32(readinessTimeoutSeconds),
									SuccessThreshold:    int32(1),
									Handler: core.Handler{
										TCPSocket: &core.TCPSocketAction{
											Port: intstr.FromInt(0),
										},
									},
								},
								LivenessProbe: &core.Probe{
									InitialDelaySeconds: int32(livenessInitialDelaySeconds),
									FailureThreshold:    int32(livenessFailureThreshold),
									PeriodSeconds:       int32(livenessPeriodSeconds),
									Handler: core.Handler{
										TCPSocket: &core.TCPSocketAction{
											Port: intstr.FromInt(0),
										},
									},
								},
								Resources: paddlesvc.Spec.Resources,
							},
						},
					},
				},
			},
		},
	},
}

var canaryServiceWithSameConfig = &knservingv1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      paddlesvc.Name,
		Namespace: paddleServiceNamespace,
	},
	Spec: knservingv1.ServiceSpec{
		ConfigurationSpec: knservingv1.ConfigurationSpec{
			Template: knservingv1.RevisionTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: paddlesvc.Name + "-canary",
					Labels: map[string]string{
						"PaddleService": paddlesvc.Name,
					},
					Annotations: annotations,
				},
				Spec: knservingv1.RevisionSpec{
					ContainerConcurrency: &containerConcurrency,
					TimeoutSeconds:       &timeoutSeconds,
					PodSpec: core.PodSpec{
						Containers: []core.Container{
							{
								ImagePullPolicy: core.PullAlways,
								Name:            paddlesvc.Spec.RuntimeVersion,
								Image:           image + ":" + tag,
								Ports: []core.ContainerPort{
									{ContainerPort: port, Name: "http1", Protocol: "TCP"},
								},
								Command: command,
								Args:    args,
								ReadinessProbe: &core.Probe{
									InitialDelaySeconds: int32(readinessInitialDelaySeconds),
									FailureThreshold:    int32(readinessFailureThreshold),
									PeriodSeconds:       int32(readinessPeriodSeconds),
									TimeoutSeconds:      int32(readinessTimeoutSeconds),
									SuccessThreshold:    int32(1),
									Handler: core.Handler{
										TCPSocket: &core.TCPSocketAction{
											Port: intstr.FromInt(0),
										},
									},
								},
								LivenessProbe: &core.Probe{
									InitialDelaySeconds: int32(livenessInitialDelaySeconds),
									FailureThreshold:    int32(livenessFailureThreshold),
									PeriodSeconds:       int32(livenessPeriodSeconds),
									Handler: core.Handler{
										TCPSocket: &core.TCPSocketAction{
											Port: intstr.FromInt(0),
										},
									},
								},
								Resources: paddlesvc.Spec.Resources,
							},
						},
					},
				},
			},
		},
		RouteSpec: knservingv1.RouteSpec{
			Traffic: []knservingv1.TrafficTarget{
				{
					RevisionName:   paddlesvc.Name + "-default",
					LatestRevision: &setLastRevision,
					Percent:        &defaultTrafficPercent,
				},
				{
					RevisionName:   paddlesvc.Name + "-canary",
					LatestRevision: &setLastRevision,
					Percent:        &defaultTrafficPercent,
				},
			},
		},
	},
}

func TestDefaultPaddleServiceToKnativeService(t *testing.T) {
	scenarios := map[string]struct {
		paddleService   elasticservingv1.PaddleService
		expectedDefault *knservingv1.Service
	}{
		"Default Test": {
			paddleService:   paddlesvc,
			expectedDefault: defaultService,
		},
	}
	serviceBuilder := NewServiceBuilder(&paddlesvc)

	for name, scenario := range scenarios {
		actualDefaultService, err := serviceBuilder.CreateService(actualTestServiceName, &paddlesvc, false)
		if err != nil {
			t.Errorf("Test %q unexpected error %s", name, err.Error())
		}
		if diff := cmp.Diff(scenario.expectedDefault, actualDefaultService); diff != "" {
			t.Errorf("Test %q unexpected canary service (-want +got): %v", name, diff)
		}
	}
}

func TestCanaryPaddleServiceToKnativeService(t *testing.T) {
	scenarios := map[string]struct {
		paddleService  elasticservingv1.PaddleService
		expectedCanary *knservingv1.Service
	}{
		"Canary Test": {
			paddleService:  paddlesvc,
			expectedCanary: canaryServiceWithSameConfig,
		},
	}
	serviceBuilder := NewServiceBuilder(&paddlesvcCanaryWithSameConfig)

	for name, scenario := range scenarios {
		actualCanaryService, err := serviceBuilder.CreateService(actualTestServiceName, &paddlesvc, true)
		if err != nil {
			t.Errorf("Test %q unexpected error %s", name, err.Error())
		}
		if diff := cmp.Diff(scenario.expectedCanary, actualCanaryService); diff != "" {
			t.Errorf("Test %q unexpected canary service (-want +got): %v", name, diff)
		}
	}
}

var defaultEndpoint = &knservingv1.Revision{
	ObjectMeta: metav1.ObjectMeta{
		Name:        paddlesvc.Name,
		Namespace:   paddlesvc.Namespace,
		Labels:      paddlesvc.Labels,
		Annotations: annotations,
	},
	Spec: knservingv1.RevisionSpec{
		TimeoutSeconds:       &constants.PaddleServiceDefaultTimeout,
		ContainerConcurrency: &containerConcurrency,
		PodSpec: core.PodSpec{
			Containers: []core.Container{
				{
					ImagePullPolicy: core.PullAlways,
					Name:            paddlesvc.Spec.RuntimeVersion,
					Image:           image + ":" + tag,
					Ports: []core.ContainerPort{
						{ContainerPort: port,
							Name:     constants.PaddleServiceDefaultPodName,
							Protocol: core.ProtocolTCP,
						},
					},
					Command: command,
					Args:    args,
					ReadinessProbe: &core.Probe{
						SuccessThreshold:    constants.SuccessThreshold,
						InitialDelaySeconds: constants.ReadinessInitialDelaySeconds,
						TimeoutSeconds:      constants.ReadinessTimeoutSeconds,
						FailureThreshold:    constants.ReadinessFailureThreshold,
						PeriodSeconds:       constants.ReadinessPeriodSeconds,
						Handler: core.Handler{
							TCPSocket: &core.TCPSocketAction{
								Port: intstr.FromInt(0),
							},
						},
					},
					LivenessProbe: &core.Probe{
						InitialDelaySeconds: constants.LivenessInitialDelaySeconds,
						FailureThreshold:    constants.LivenessFailureThreshold,
						PeriodSeconds:       constants.LivenessPeriodSeconds,
						Handler: core.Handler{
							TCPSocket: &core.TCPSocketAction{
								Port: intstr.FromInt(0),
							},
						},
					},
					Resources: paddlesvc.Spec.Resources,
				},
			},
		},
	},
}

func TestPaddleEndpointToKnativeRevision(t *testing.T) {
	scenarios := map[string]struct {
		paddleService    elasticservingv1.PaddleService
		expectedRevision *knservingv1.Revision
	}{
		"Default Test": {
			paddleService:    paddlesvc,
			expectedRevision: defaultEndpoint,
		},
	}
	serviceBuilder := NewServiceBuilder(&paddlesvc)

	for name, scenario := range scenarios {
		actualDefaultEndpoint, err := serviceBuilder.CreateRevision(actualTestServiceName, &paddlesvc, false)
		if err != nil {
			t.Errorf("Test %q unexpected error %s", name, err.Error())
		}
		if diff := cmp.Diff(scenario.expectedRevision, actualDefaultEndpoint); diff != "" {
			t.Errorf("Test %q unexpected canary service (-want +got): %v", name, diff)
		}
	}
}
