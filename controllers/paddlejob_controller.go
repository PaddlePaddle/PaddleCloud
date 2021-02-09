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
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=paddlejobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=paddlejobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.paddlepaddle.org,resources=paddlejobs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods/status,verbs=get
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Reconcile function compares the state specified by
// the PaddleJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *PaddleJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("paddlejob", req.NamespacedName)

	// Obtain the PaddleJob instance we are working on
	var pdj pdv1.PaddleJob
	if err := r.Get(ctx, req.NamespacedName, &pdj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Reconcile", "version", pdj.ResourceVersion)

	//r.finalize(ctx, &pdj)

	// List all associated pods
	var childPods corev1.PodList
	if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingFields{ctrlRefKey: req.Name}); err != nil {
		return ctrl.Result{}, err
	}
	var podsMap = make(map[string]bool)
	for _, pod := range childPods.Items {
		podsMap[pod.Name] = true
	}

	newStatus := r.currentStatus(ctx, &pdj, childPods)
	if !reflect.DeepEqual(newStatus, pdj.Status) {
		pdj.Status = newStatus
		if err := r.Status().Update(ctx, &pdj); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	}

	// List all associated svc
	var svcs corev1.ServiceList
	if err := r.List(ctx, &svcs, client.InNamespace(req.Namespace), client.MatchingFields{ctrlRefKey: req.Name}); err != nil {
		return ctrl.Result{}, err
	}
	var svcsMap = make(map[string]bool)
	for _, svc := range svcs.Items {
		svcsMap[svc.Name] = true
	}

	cleanOne := func() {
		for i := range childPods.Items {
			r.deleteResource(ctx, &pdj, &childPods.Items[i])
			return
		}
		for i := range svcs.Items {
			r.deleteResource(ctx, &pdj, &svcs.Items[i])
			return
		}
	}

	if pdj.Status.Phase == pdv1.Failed {
		if pdj.Spec.CleanPolicy == pdv1.CleanAll || pdj.Spec.CleanPolicy == pdv1.CleanOnFailure {
			cleanOne()
			return ctrl.Result{}, nil
		}
	}
	if pdj.Status.Phase == pdv1.Completed {
		if pdj.Spec.CleanPolicy == "" || pdj.Spec.CleanPolicy == pdv1.CleanAll || pdj.Spec.CleanPolicy == pdv1.CleanOnCompletion {
			cleanOne()
			return ctrl.Result{}, nil
		}
	}

	// clean pod unnecessary
	if len(childPods.Items) > pdj.Spec.PS.Replicas+pdj.Spec.Worker.Replicas {
		for i, pod := range childPods.Items {
			resType, idx := extractNameIndex(pod.Name)
			if resType == pdv1.ResourcePS {
				if idx >= pdj.Spec.PS.Replicas {
					r.deleteResource(ctx, &pdj, &childPods.Items[i])
					return ctrl.Result{RequeueAfter: time.Second}, nil
				}
			} else if resType == pdv1.ResourceWorker {
				if idx >= pdj.Spec.Worker.Replicas {
					r.deleteResource(ctx, &pdj, &childPods.Items[i])
					return ctrl.Result{RequeueAfter: time.Second}, nil
				}
			}
		}
	}

	// Ensure service for running pod
	for _, pod := range childPods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			if svcsMap[pod.Name] {
				continue
			}
			svc := constructService4Pod(pod)
			err := ctrl.SetControllerReference(&pdj, svc, r.Scheme)
			if err != nil {
				log.Error(err, "make reference failed")
				continue
			}
			if err := r.Get(ctx, client.ObjectKeyFromObject(svc), &corev1.Service{}); err == nil {
				continue
			}
			err = r.createResource(ctx, &pdj, svc)
			return ctrl.Result{}, err
		}
	}

	// Ensure PS resource ready
	if len(pdj.Status.PS.Refs) < pdj.Spec.PS.Replicas {
		for i := 0; i < pdj.Spec.PS.Replicas; i++ {
			name := genPaddleResName(pdj.Name, pdv1.ResourcePS, i)
			if podsMap[name] {
				continue
			}
			pod := constructPS4PaddleJob(&pdj, i)
			err := ctrl.SetControllerReference(&pdj, pod, r.Scheme)
			if err != nil {
				log.Error(err, "make reference failed")
				continue
			}
			if err := r.Get(ctx, client.ObjectKeyFromObject(pod), &corev1.Pod{}); err == nil {
				continue
			}
			err = r.createResource(ctx, &pdj, pod)
			return ctrl.Result{}, err
		}
	}

	// Ensure worker resource ready
	if pdj.Status.PS.Running == pdj.Spec.PS.Replicas && len(pdj.Status.Worker.Refs) < pdj.Spec.Worker.Replicas {
		for i := 0; i < pdj.Spec.Worker.Replicas; i++ {
			name := genPaddleResName(pdj.Name, pdv1.ResourceWorker, i)
			if podsMap[name] {
				continue
			}
			pod := constructWorker4PaddleJob(&pdj, i)
			err := ctrl.SetControllerReference(&pdj, pod, r.Scheme)
			if err != nil {
				log.Error(err, "make reference failed")
				continue
			}
			if err := r.Get(ctx, client.ObjectKeyFromObject(pod), &corev1.Pod{}); err == nil {
				continue
			}
			err = r.createResource(ctx, &pdj, pod)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PaddleJobReconciler) currentStatus(ctx context.Context, pdj *pdv1.PaddleJob, childPods corev1.PodList) pdv1.PaddleJobStatus {
	syncStatusByPod := func(ss *pdv1.ResourceStatus, pod *corev1.Pod) {
		switch pod.Status.Phase {
		case corev1.PodPending:
			ss.Pending++
		case corev1.PodRunning:
			if isPodRealRuning(pod) {
				ss.Running++
			} else {
				ss.Starting++
			}
		case corev1.PodFailed:
			ss.Failed++
		case corev1.PodSucceeded:
			ss.Succeeded++
		}
		pref, err := ref.GetReference(r.Scheme, pod)
		if err != nil {
			return
		}
		ss.Refs = append(ss.Refs, *pref)
	}

	psStatus := pdv1.ResourceStatus{}
	workerStatus := pdv1.ResourceStatus{}
	for i, pod := range childPods.Items {
		resType := pod.Annotations[pdv1.ResourceAnnotation]
		if resType == pdv1.ResourcePS {
			syncStatusByPod(&psStatus, &childPods.Items[i])
		} else if resType == pdv1.ResourceWorker {
			syncStatusByPod(&workerStatus, &childPods.Items[i])
		}
	}

	psStatus.Ready = fmt.Sprintf("%d/%d", psStatus.Running, pdj.Spec.PS.Replicas)
	workerStatus.Ready = fmt.Sprintf("%d/%d", workerStatus.Running, pdj.Spec.Worker.Replicas)

	return pdv1.PaddleJobStatus{
		Phase:  getPaddleJobPhase(pdj),
		Mode:   getPaddleJobMode(pdj),
		PS:     psStatus,
		Worker: workerStatus,
	}
}

func (r *PaddleJobReconciler) deleteResource(ctx context.Context, pdj *pdv1.PaddleJob, obj client.Object) error {
	if obj.GetDeletionTimestamp() != nil {
		return nil
	}
	tp := obj.GetObjectKind().GroupVersionKind().Kind
	if err := r.Delete(ctx, obj, client.PropagationPolicy(metav1.DeletePropagationBackground)); (err) != nil {
		r.Recorder.Event(pdj, corev1.EventTypeWarning, "Delete", fmt.Sprintf("delete failed %s %s", tp, obj.GetName()))
		return err
	} else {
		r.Recorder.Event(pdj, corev1.EventTypeNormal, "Deleted", fmt.Sprintf("deleted %s %s", tp, obj.GetName()))
		return nil
	}
}

func (r *PaddleJobReconciler) createResource(ctx context.Context, pdj *pdv1.PaddleJob, obj client.Object) error {
	tp := obj.GetObjectKind().GroupVersionKind().Kind
	if err := r.Create(ctx, obj); err != nil {
		r.Recorder.Event(pdj, corev1.EventTypeWarning, "Create", fmt.Sprintf("create failed %s %s", tp, obj.GetName()))
		return err
	} else {
		r.Recorder.Event(pdj, corev1.EventTypeNormal, "Created", fmt.Sprintf("created %s %s", tp, obj.GetName()))
		return nil
	}
}

func (r *PaddleJobReconciler) finalize(ctx context.Context, obj *pdv1.PaddleJob) error {
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
