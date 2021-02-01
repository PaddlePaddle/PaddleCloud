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
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	//batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pdv1 "github.com/paddleflow/paddle-operator/api/v1"
)

var (
	ctrlRefKey = ".metadata.controller"
	apiGVStr   = pdv1.GroupVersion.String()
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

	// List all associated pods
	var childPods corev1.PodList
	if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingFields{ctrlRefKey: req.Name}); err != nil {
		log.Error(err, "unable to list child Jobs")
		return ctrl.Result{}, err
	}

	// Initialize Status
	if pdj.Status.Phase == "" {
		pdj.Status.Phase = pdv1.Starting
	}
	if pdj.Status.Mode == "" {
		pdj.Status.Mode = getPaddleJobMode(&pdj)
	}
	pdj.Status.PS = pdv1.ResourceStatus{}
	pdj.Status.Worker = pdv1.ResourceStatus{}

    // Sync status by pods
	fillStatusByChildren := func(ss *pdv1.ResourceStatus, pod *corev1.Pod) {
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
		}
		ss.Refs = append(ss.Refs, *pref)
	}
	for _, pod := range childPods.Items {
		resType := pod.Annotations[pdv1.ResourceAnnotation]
		if resType == pdv1.ResourcePS {
			fillStatusByChildren(&pdj.Status.PS, &pod)
		} else if resType == pdv1.ResourceWorker {
			fillStatusByChildren(&pdj.Status.Worker, &pod)
		}
	}
	if pdj.Spec.PS.Replicas == pdj.Status.PS.Running && pdj.Spec.Worker.Replicas == pdj.Status.Worker.Running {
		pdj.Status.Phase = pdv1.Running
	} else if pdj.Status.PS.Failed > 0 || pdj.Status.Worker.Failed > 0 {
		pdj.Status.Phase = pdv1.Failed
	} else if pdj.Spec.PS.Replicas >= pdj.Status.PS.Succeeded && pdj.Spec.Worker.Replicas == pdj.Status.Worker.Succeeded {
		pdj.Status.Phase = pdv1.Completed
	}
    // more phase HERE

    // split line ---------------------------------
    // Important action : sync status above, sync task below
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

	// Ensure PS resource ready
	if len(pdj.Status.PS.Refs) < pdj.Spec.PS.Replicas {
		pod, err := constructPS4PaddleJob(&pdj, len(pdj.Status.PS.Refs))
		err = ctrl.SetControllerReference(&pdj, pod, r.Scheme)
		err = r.Create(ctx, pod)
		log.V(1).Info("create ps", "error", err)
		return ctrl.Result{}, nil
	}

	if pdj.Status.PS.Running == pdj.Spec.PS.Replicas && len(pdj.Status.Worker.Refs) < pdj.Spec.Worker.Replicas {
		pod, err := constructWorker4PaddleJob(&pdj, len(pdj.Status.Worker.Refs))
		err = ctrl.SetControllerReference(&pdj, pod, r.Scheme)
		err = r.Create(ctx, pod)
		log.V(1).Info("create worker", "error", err)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PaddleJobReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, ctrlRefKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*corev1.Pod)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a PaddleJob...
		if owner.APIVersion != apiGVStr || owner.Kind != pdv1.KIND {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&pdv1.PaddleJob{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
