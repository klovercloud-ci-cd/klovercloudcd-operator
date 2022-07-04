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
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	basev1alpha1 "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
)

// KlovercloudCDReconciler reconciles a KlovercloudCD object
type KlovercloudCDReconciler struct {
	client.Client
	*rest.Config
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=base.cd.klovercloud.com,resources=klovercloudcds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=base.cd.klovercloud.com,resources=klovercloudcds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=base.cd.klovercloud.com,resources=klovercloudcds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KlovercloudCD object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *KlovercloudCDReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	config := &basev1alpha1.KlovercloudCD{}
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
	// Apply prerequisites
	err = descriptor.ApplyPrerequisites(r.Client, config.Namespace, config.Spec.Database, config.Spec.Security, string(config.Spec.Version))
	if err != nil {
		return ctrl.Result{}, err
	}

	// Apply api service
	err=descriptor.ApplyApiService(r.Client,config.Namespace,config.Spec.ApiService,string(config.Spec.Version))
	if err != nil {
		return ctrl.Result{}, err
	}
	// Apply integration manager
	err = descriptor.ApplyIntegrationManager(r.Client, config.Namespace, config.Spec.Database, config.Spec.IntegrationManager, string(config.Spec.Version))
	if err != nil {
		return ctrl.Result{}, err
	}

	// Apply event bank
	err = descriptor.ApplyEventBank(r.Client, config.Namespace, config.Spec.Database, config.Spec.EventBank, string(config.Spec.Version))
	if err != nil {
		return ctrl.Result{}, err
	}

	// Apply core engine

	err=descriptor.ApplyCoreEngine(r.Client,config.Namespace,config.Spec.Database,config.Spec.CoreEngine,string(config.Spec.Version))
	if err != nil {
		return ctrl.Result{}, err
	}

	// Apply security
	err = descriptor.ApplySecurity(r.Client, config.Namespace, config.Spec.Database, config.Spec.Security, string(config.Spec.Version))
	if err != nil {
		return ctrl.Result{}, err
	}

	// Apply lighthouse
	if config.Spec.Agent.LightHouseEnabled=="true"{
		// Apply lighthouse command
		err=descriptor.ApplyLightHouseCommand(r.Client,config.Namespace,config.Spec.Database,config.Spec.LightHouse.Command,string(config.Spec.Version))
		if err != nil {
			return ctrl.Result{}, err
		}
		// Apply lighthouse query
		err=descriptor.ApplyLightHouseQuery(r.Client,config.Namespace,config.Spec.Database,config.Spec.LightHouse.Query,string(config.Spec.Version))
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Apply agent
	err = descriptor.ApplyAgent(r.Client,r.Config, config.Namespace,config.Spec.Agent, string(config.Spec.Version))
	if err != nil {
		return ctrl.Result{}, err
	}



	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KlovercloudCDReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&basev1alpha1.KlovercloudCD{}).
		Complete(r)
}
