/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	//batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pdv1 "github.com/paddleflow/paddle-operator/api/v1"
)

var (
	ctrlRefKey    = ".metadata.controller"
	apiGVStr      = pdv1.GroupVersion.String()
	finalizerName = "finalizers.paddlepaddle.org"
)

// PaddleJobReconciler reconciles a PaddleJob object
type PaddleJobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=paddlejobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=paddlejobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=paddlejobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=pods/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PaddleJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *PaddleJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("paddlejob", req.NamespacedName)

	log.V(1).Info("Reconcile start -----------------------------------------------------\n")
	defer log.V(1).Info("Reconcile end -----------------------------------------------------\n")

	// Obtain the PaddleJob instance we are working on
	var pdj pdv1.PaddleJob
	if err := r.Get(ctx, req.NamespacedName, &pdj); err != nil {
		log.Error(err, "failed to fetch paddlejob")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// r.finalize(&pdj)

	// List all associated pods
	var childPods corev1.PodList
	if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingFields{ctrlRefKey: req.Name}); err != nil {
		log.Error(err, "unable to list child Jobs")
		return ctrl.Result{}, err
	}

	syncStatusByPod := func(ss *pdv1.ResourceStatus, pod *corev1.Pod) {
		switch pod.Status.Phase {
		case corev1.PodPending:
			ss.Pending++
		case corev1.PodRunning:
			ss.Running++
		case corev1.PodFailed:
			ss.Failed++
		case corev1.PodSucceeded:
			ss.Succeeded++
		}
		pref, err := ref.GetReference(r.Scheme, pod)
		if err != nil {
			log.Error(err, "get reference failed", "pod", pod)
		} else {
			ss.Refs = append(ss.Refs, *pref)
		}
	}
	var podsMap = make(map[string]*corev1.Pod)
	// Initialize Status before sync
	pdj.Status.PS = pdv1.ResourceStatus{}
	pdj.Status.Worker = pdv1.ResourceStatus{}
	for _, pod := range childPods.Items {
		log.V(1).Info("pod to map", "pod", pod.Name)
		podsMap[pod.Name] = &pod
		resType := pod.Annotations[pdv1.ResourceAnnotation]
		if resType == pdv1.ResourcePS {
			syncStatusByPod(&pdj.Status.PS, &pod)
		} else if resType == pdv1.ResourceWorker {
			syncStatusByPod(&pdj.Status.Worker, &pod)
		}
	}
	if pdj.Spec.PS.Replicas == pdj.Status.PS.Running && pdj.Spec.Worker.Replicas == pdj.Status.Worker.Running {
		pdj.Status.Phase = pdv1.Running
	} else if pdj.Status.PS.Failed > 0 || pdj.Status.Worker.Failed > 0 {
		pdj.Status.Phase = pdv1.Failed
	} else if pdj.Spec.PS.Replicas >= pdj.Status.PS.Succeeded && pdj.Spec.Worker.Replicas == pdj.Status.Worker.Succeeded {
		pdj.Status.Phase = pdv1.Completed
	} else if pdj.Status.PS.Pending > 0 || pdj.Status.Worker.Pending > 0 {
		pdj.Status.Phase = pdv1.Starting
	}
	// more phase HERE
	pdj.Status.Mode = getPaddleJobMode(&pdj)

	// Important action : sync status above, take action below
	if err := r.Status().Update(ctx, &pdj); err != nil {
		log.Error(err, "unable to update status")
		return ctrl.Result{}, err
	}

	if pdj.Status.Phase == pdv1.Failed {
		log.V(1).Info("job failed, do nothing now")
		return ctrl.Result{}, nil
	}
	if pdj.Status.Phase == pdv1.Completed {
		log.V(1).Info("job completed, may be clean")
		return ctrl.Result{}, nil
	}

	// List all associated svc
	var svcs corev1.ServiceList
	if err := r.List(ctx, &svcs, client.InNamespace(req.Namespace), client.MatchingFields{ctrlRefKey: req.Name}); err != nil {
		log.Error(err, "unable to list child Services ")
		return ctrl.Result{}, err
	}
	var svcsMap = make(map[string]*corev1.Service)
	for _, svc := range svcs.Items {
		svcsMap[svc.Name] = &svc
	}

	// Ensure service for running pod
	for _, pod := range childPods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			if svcsMap[pod.Name] != nil {
				continue
			}
			svc := constructService4Pod(&pod)
			err := ctrl.SetControllerReference(&pdj, svc, r.Scheme)
			if err != nil {
				log.Error(err, "make reference failed")
				continue
			}
			err = r.Create(ctx, svc)
			if err != nil {
				log.Error(err, "create failed")
				continue
			}
			return ctrl.Result{}, nil
		}
	}

	// Ensure PS resource ready
	if len(pdj.Status.PS.Refs) < pdj.Spec.PS.Replicas {
		for i := 0; i < pdj.Spec.PS.Replicas; i++ {
			name := genPaddleResName(pdj.Name, pdv1.ResourcePS, i)
			if podsMap[name] != nil {
				continue
			}
			pod := constructPS4PaddleJob(&pdj, i)
			err := ctrl.SetControllerReference(&pdj, pod, r.Scheme)
			if err != nil {
				log.Error(err, "make reference failed")
				continue
			}
			err = r.Create(ctx, pod)
			if err != nil {
				log.Error(err, "create failed")
				continue
			}
			return ctrl.Result{}, nil
		}
	}

	// Ensure worker resource ready
	if pdj.Status.PS.Running == pdj.Spec.PS.Replicas && len(pdj.Status.Worker.Refs) < pdj.Spec.Worker.Replicas {
		for i := 0; i < pdj.Spec.Worker.Replicas; i++ {
			name := genPaddleResName(pdj.Name, pdv1.ResourceWorker, i)
			if podsMap[name] != nil {
				continue
			}
			pod := constructWorker4PaddleJob(&pdj, i)
			err := ctrl.SetControllerReference(&pdj, pod, r.Scheme)
			if err != nil {
				log.Error(err, "make reference failed")
				continue
			}
			err = r.Create(ctx, pod)
			if err != nil {
				log.Error(err, "create failed")
				continue
			}
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

/*
func (r *PaddleJobReconciler) finalize(ctx context.Context, obj client.Object) error {
	if obj.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(obj.ObjectMeta.Finalizers, finalizerName) {
			obj.ObjectMeta.Finalizers = append(obj.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, obj); err != nil {
				return err
			}
		}
	} else {
		if containsString(obj.ObjectMeta.Finalizers, finalizerName) {

			// do before delete

			obj.ObjectMeta.Finalizers = removeString(obj.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, obj); err != nil {
				return err
			}
		}
		return nil
	}
}
*/

func (r *PaddleJobReconciler) deleteService(ctx context.Context, nn types.NamespacedName) error {
	svc := corev1.Service{}
	if err := r.Get(ctx, nn, &svc); err != nil {
		return err
	}
	if err := r.Delete(ctx, &svc); err != nil {
		return err
	}
	return nil
}

func indexerFunc(rawObj client.Object) []string {
	owner := metav1.GetControllerOf(rawObj)
	if owner == nil {
		return nil
	}
	// ...make sure it's a PaddleJob...
	if owner.APIVersion != apiGVStr || owner.Kind != pdv1.KIND {
		return nil
	}

	// ...and if so, return it
	return []string{owner.Name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PaddleJobReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// index pod
	if err := mgr.GetFieldIndexer().
		IndexField(context.Background(), &corev1.Pod{}, ctrlRefKey, indexerFunc); err != nil {
		return err
	}

	// index service
	if err := mgr.GetFieldIndexer().
		IndexField(context.Background(), &corev1.Service{}, ctrlRefKey, indexerFunc); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&pdv1.PaddleJob{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
