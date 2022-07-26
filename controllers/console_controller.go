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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	basev1alpha1 "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
)

// ConsoleReconciler reconciles a Console object
type ConsoleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=base.cd.klovercloud.com,resources=consoles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=base.cd.klovercloud.com,resources=consoles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=base.cd.klovercloud.com,resources=consoles/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;delete;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Console object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ConsoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	config := &basev1alpha1.Console{}
	err := r.Get(ctx, req.NamespacedName, config)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, " Error reading the object - requeue the request")
		return ctrl.Result{}, err
	}
	// TODO(user): your logic here

	log.Info("Applying all api service ...")
	// Apply UI Console service
	// Check if the deployment already exists, if not create a new one
	existingUIConsoleService := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-console", Namespace: config.Namespace}, existingUIConsoleService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		err = descriptor.ApplyConsole(r.Client, config, r.Scheme, config.Namespace, string(config.Spec.Version))
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", config.Namespace, "Deployment.Name", "klovercloud-ci-console")
			return ctrl.Result{}, err
		}
		//return ctrl.Result{}, err
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}
	existingUIConsoleConfigmap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-console-envar-config", Namespace: config.Namespace}, existingUIConsoleConfigmap)
	if err != nil {
		log.Error(err, "Failed to get klovercloud-ci-console-envar-config.", err.Error())
		return ctrl.Result{}, err
	}
	redeploy := false
	redeployConfigmap := false

	if existingUIConsoleConfigmap.Data["V1_AUTH_ENDPOINT"] != config.Spec.Console.AuthEndpoint {
		redeployConfigmap = true
		existingUIConsoleConfigmap.Data["V1_AUTH_ENDPOINT"] = config.Spec.Console.AuthEndpoint
	}
	if existingUIConsoleConfigmap.Data["V1_API_ENDPOINT"] != config.Spec.Console.ApiEndpoint {
		redeployConfigmap = true
		existingUIConsoleConfigmap.Data["V1_API_ENDPOINT"] = config.Spec.Console.ApiEndpoint
	}
	if existingUIConsoleConfigmap.Data["V1_API_ENDPOINT_WS"] != config.Spec.Console.ApiEndpointWS {
		redeployConfigmap = true
		existingUIConsoleConfigmap.Data["V1_API_ENDPOINT_WS"] = config.Spec.Console.ApiEndpointWS
	}

	for i, each := range existingUIConsoleService.Spec.Template.Spec.Containers {
		if each.Name == "app" {
			isRequestedResourcesChanged := each.Resources.Requests.Cpu().ToDec().String() != config.Spec.Console.Resources.Requests.Cpu().ToDec().String() || each.Resources.Requests.Memory().ToDec().String() != config.Spec.Console.Resources.Requests.Memory().ToDec().String()
			isLimitedRequestedChanged := each.Resources.Limits.Cpu().ToDec().String() != config.Spec.Console.Resources.Limits.Cpu().ToDec().String() || each.Resources.Limits.Memory().ToDec().String() != config.Spec.Console.Resources.Limits.Memory().ToDec().String()

			if isRequestedResourcesChanged || isLimitedRequestedChanged {
				redeploy = true
				existingUIConsoleService.Spec.Template.Spec.Containers[i].Resources = config.Spec.Console.Resources
				break
			}

		}
	}
	if redeployConfigmap {
		err = r.Update(ctx, existingUIConsoleConfigmap)
		if err != nil {
			log.Error(err, "Failed to update ui console configmap.", "Namespace:", existingUIConsoleConfigmap.Namespace, "Name:", existingUIConsoleConfigmap.Name)
			return ctrl.Result{}, err
		}
		redeploy = true
	}

	if redeploy {
		existingUIConsoleService.Spec.Template.ObjectMeta.Annotations = map[string]string{"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.RFC3339)}
		err = r.Update(ctx, existingUIConsoleService)
		if err != nil {
			log.Error(err, "Failed to update ui console Deployment.", "Deployment.Namespace:", existingUIConsoleService.Namespace, "Deployment.Name:", existingUIConsoleService.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the agent status with the pod names
	// List the pods for this api service's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(config.Namespace),
		client.MatchingLabels(map[string]string{"app": "klovercloud-ci-console"}),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods.", "Console.Namespace:", config.Namespace, "Console.Name:", "klovercloud-ci-console")
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, config.Status.ConsolePods) {
		config.Status.ConsolePods = podNames
		err := r.Status().Update(ctx, config)
		if err != nil {
			log.Error(err, "Failed to update ConsolePods status")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConsoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&basev1alpha1.Console{}).
		Complete(r)
}
