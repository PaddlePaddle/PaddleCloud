package knative

import (
	"ElasticServing/pkg/constants"
	"ElasticServing/pkg/controllers/elasticserving/resources/knative"
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	elasticservingv1 "ElasticServing/pkg/apis/serving/v1"
)

var log = logf.Log.WithName("ServiceReconciler")

type ServiceReconciler struct {
	client         client.Client
	scheme         *runtime.Scheme
	serviceBuilder *knative.ServiceBuilder
}

func NewServiceReconciler(client client.Client, scheme *runtime.Scheme, paddlesvc *elasticservingv1.PaddleService) *ServiceReconciler {
	return &ServiceReconciler{
		client:         client,
		scheme:         scheme,
		serviceBuilder: knative.NewServiceBuilder(paddlesvc),
	}
}

// +kubebuilder:rbac:groups=serving.knative.dev,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serving.knative.dev,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=serving.knative.dev,resources=revisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serving.knative.dev,resources=revisions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;create;
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;create;
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;create;
// +kubebuilder:rbac:groups="",resources=services,verbs=*
// +kubebuilder:rbac:groups="",resources=pods,verbs=*
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=serving.paddlepaddle.org,resources=paddleservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serving.paddlepaddle.org,resources=paddleservices/status,verbs=get;update;patch

func (r *ServiceReconciler) Reconcile(paddlesvc *elasticservingv1.PaddleService) error {
	var service *knservingv1.Service
	var serviceWithCanary *knservingv1.Service
	var err error
	serviceName := paddlesvc.Name
	service, err = r.serviceBuilder.CreateService(serviceName, paddlesvc, false)
	if err != nil {
		return err
	}

	if service == nil {
		if err = r.finalizeService(serviceName, paddlesvc.Namespace); err != nil {
			return err
		}
		// TODO: Modify status
		// paddlesvc.Status.PropagateStatus(nil)
		return nil
	}

	if _, err := r.reconcileDefaultEndpoint(paddlesvc, service); err != nil {
		return err
	} else {
		// TODO: Modify status
		// paddlesvc.Status.PropagateStatus(status)
	}

	serviceWithCanary, err = r.serviceBuilder.CreateService(serviceName, paddlesvc, true)
	if err != nil {
		return err
	}
	if serviceWithCanary == nil {
		if err = r.finalizeCanaryEndpoint(serviceName, paddlesvc.Namespace, service.Spec); err != nil {
			return err
		}
		return nil
	}

	if _, err := r.reconcileCanaryEndpoint(paddlesvc, serviceWithCanary, service.Spec); err != nil {
		return err
	} else {
		// TODO: Modify status
		// paddlesvc.Status.PropagateStatus(status)
	}

	return nil
}

func (r *ServiceReconciler) finalizeService(serviceName, namespace string) error {
	existing := &knservingv1.Service{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: namespace}, existing); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		log.Info("Deleting Knative Service", "namespace", namespace, "name", serviceName)
		if err := r.client.Delete(context.TODO(), existing, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
		}
	}
	return nil
}

func (r *ServiceReconciler) finalizeCanaryEndpoint(serviceName, namespace string, serviceSpec knservingv1.ServiceSpec) error {
	existing := &knservingv1.Service{}
	existingRevision := &knservingv1.Revision{}
	canaryServiceName := constants.CanaryServiceName(serviceName)
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: canaryServiceName, Namespace: namespace}, existingRevision); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: namespace}, existing); err != nil {
			return err
		}

		existing.Spec = serviceSpec
		if err := r.client.Update(context.TODO(), existing); err != nil {
			return err
		}

		log.Info("Deleting Knative Canary Endpoint", "namespace", namespace, "name", serviceName)
		if err := r.client.Delete(context.TODO(), existingRevision, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
		}
	}
	return nil
}

func (r *ServiceReconciler) reconcileDefaultEndpoint(paddlesvc *elasticservingv1.PaddleService, desired *knservingv1.Service) (*knservingv1.ServiceStatus, error) {
	// Set Paddlesvc as owner of desired service
	if err := controllerutil.SetControllerReference(paddlesvc, desired, r.scheme); err != nil {
		return nil, err
	}

	// Create service if does not exist
	existing := &knservingv1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Knative Default Endpoint", "namespace", desired.Namespace, "name", desired.Name)
			err = r.client.Create(context.TODO(), desired)
			if err != nil {
				return nil, err
			}
			for {
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)
				if err == nil || !errors.IsNotFound(err) {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			if err != nil {
				return nil, err
			}
			return &existing.Status, nil
		}
		return nil, err
	}

	existingRevision := &knservingv1.Revision{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: constants.DefaultServiceName(desired.Name), Namespace: desired.Namespace}, existingRevision)
	if err != nil {
		return nil, err
	}
	desiredRevision, err := r.serviceBuilder.CreateRevision(constants.DefaultServiceName(desired.Name), paddlesvc, false)
	if err != nil {
		return nil, err
	}

	if knativeRevisionSemanticEquals(desiredRevision, existingRevision) {
		log.Info("No differences on revision found")
		return &existing.Status, nil
	}

	// The update process includes two steps.
	// 1. Delete the default endpoint(revision)
	// 2. Update knative service whose template should be updated to desired.Spec
	err = r.client.Delete(context.TODO(), existingRevision, client.PropagationPolicy(metav1.DeletePropagationBackground))
	if err != nil {
		return nil, err
	}

	existing.Spec = desired.Spec

	if paddlesvc.Spec.Canary != nil {
		r.serviceBuilder.AddTrafficRoute(paddlesvc.Name, paddlesvc, existing)
	}

	err = r.client.Update(context.TODO(), existing)
	if err != nil {
		return nil, err
	}

	return &existing.Status, nil
}

func (r *ServiceReconciler) reconcileCanaryEndpoint(paddlesvc *elasticservingv1.PaddleService, desired *knservingv1.Service, serviceSpec knservingv1.ServiceSpec) (*knservingv1.ServiceStatus, error) {
	// Set Paddlesvc as owner of desired service
	if err := controllerutil.SetControllerReference(paddlesvc, desired, r.scheme); err != nil {
		return nil, err
	}

	existingRevision := &knservingv1.Revision{}
	existing := &knservingv1.Service{}

	//Create canary revision if does not exist
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: constants.CanaryServiceName(desired.Name), Namespace: desired.Namespace}, existingRevision)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Canary Revision", "namespace", desired.Namespace, "name", desired.Name)
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)
			if err != nil {
				return &desired.Status, err
			}

			if knativeSpecSemanticEquals(desired.Spec, existing.Spec) {
				return &existing.Status, nil
			}
			existing.Spec = desired.Spec

			err = r.client.Update(context.TODO(), existing)
			if err != nil {
				return nil, err
			}
			return &existing.Status, nil
		}
		return nil, err
	}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: constants.CanaryServiceName(desired.Name), Namespace: desired.Namespace}, existingRevision)
	if err != nil {
		return nil, err
	}

	desiredRevision, err := r.serviceBuilder.CreateRevision(constants.CanaryServiceName(desired.Name), paddlesvc, true)
	if err != nil {
		return nil, err
	}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)
	if err != nil {
		return &desired.Status, err
	}

	if knativeRevisionSemanticEquals(desiredRevision, existingRevision) &&
		knativeServiceTrafficSemanticEquals(desired, existing) {
		log.Info("No differences on revision found")
		return &existing.Status, nil
	}

	// The update process includes two steps.
	// 1. Delete the canary endpoint(revision)
	// 2. Update knative service whose template should be updated to desired.Spec
	err = r.finalizeCanaryEndpoint(paddlesvc.Name, paddlesvc.Namespace, serviceSpec)
	if err != nil {
		return nil, err
	}

	existing.Spec = desired.Spec

	err = r.client.Update(context.TODO(), existing)
	if err != nil {
		return nil, err
	}
	return &existing.Status, nil
}

func knativeSpecSemanticEquals(desired, existing interface{}) bool {
	return equality.Semantic.DeepDerivative(desired, existing)
}

func knativeServiceTrafficSemanticEquals(desired, existing *knservingv1.Service) bool {
	return equality.Semantic.DeepDerivative(desired.Spec.RouteSpec, existing.Spec.RouteSpec)
}

func knativeRevisionSemanticEquals(desired, existing *knservingv1.Revision) bool {
	return equality.Semantic.DeepDerivative(desired.ObjectMeta.Annotations, existing.ObjectMeta.Annotations) &&
		equality.Semantic.DeepDerivative(desired.Spec, existing.Spec)
}
