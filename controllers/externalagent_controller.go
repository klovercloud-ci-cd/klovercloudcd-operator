/*
Copyright 2022.

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
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	basev1alpha1 "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
)

// ExternalAgentReconciler reconciles a ExternalAgent object
type ExternalAgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=base.cd.klovercloud.com,resources=externalagents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=base.cd.klovercloud.com,resources=externalagents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=base.cd.klovercloud.com,resources=externalagents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ExternalAgent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ExternalAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	config := &basev1alpha1.ExternalAgent{}
	err := r.Get(ctx, req.NamespacedName, config)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, " Error reading the object - requeue the request")
		return reconcile.Result{}, err
	}

	// TODO(user): your logic here

	err = descriptor.ApplyExternalAgent(r.Client, config.Namespace, config.Spec.Agent, string(config.Spec.Version))
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExternalAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&basev1alpha1.ExternalAgent{}).
		Complete(r)
}
