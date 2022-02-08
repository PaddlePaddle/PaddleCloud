/*
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
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	elasticservingv1 "github.com/paddlepaddle/paddlecloud/pkg/apis/serving/v1"
	"github.com/paddlepaddle/paddlecloud/pkg/serving/reconcilers/knative"
)

// PaddleServiceReconciler reconciles a PaddleService object
type PaddleServiceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=serving.paddlepaddle.org,resources=paddleservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serving.paddlepaddle.org,resources=paddleservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *PaddleServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("paddlesvc", req.NamespacedName)

	// your logic here
	log.Info("reconciling paddlesvc")

	// Load the PaddleService by name
	var paddlesvc elasticservingv1.PaddleService
	if err := r.Get(ctx, req.NamespacedName, &paddlesvc); err != nil {
		log.Error(err, "unable to fetch PaddleService")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Successfully fetching paddlesvc")

	serviceReconciler := knative.NewServiceReconciler(r.Client, r.Scheme, &paddlesvc)

	if err := serviceReconciler.Reconcile(&paddlesvc); err != nil {
		r.Log.Error(err, "Failed to finish knative reconcile")
		r.Recorder.Eventf(&paddlesvc, core.EventTypeWarning, "InternalError", err.Error())
		return reconcile.Result{}, err
	}

	// Update status
	if err := r.Status().Update(ctx, &paddlesvc); err != nil {
		r.Recorder.Eventf(&paddlesvc, core.EventTypeWarning, "InternalError", err.Error())
		return ctrl.Result{}, err
	}

	log.Info("resource status synced")

	return ctrl.Result{}, nil
}

func (r *PaddleServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticservingv1.PaddleService{}).
		Owns(&apps.Deployment{}).
		Owns(&knservingv1.Service{}).
		Complete(r)
}
