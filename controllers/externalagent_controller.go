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
	basev1alpha1 "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
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

	existingAgent := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-agent", Namespace: config.Namespace}, existingAgent)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		err = descriptor.ApplyExternalAgent(r.Client, config.Namespace, config.Spec.Agent, string(config.Spec.Version))
		if err != nil {
			log.Error(err, "Failed to create external agent Deployment", "Deployment.Namespace", config.Namespace, "Deployment.Name", "klovercloud-ci-agent")
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get external agent Deployment")
		return ctrl.Result{}, err
	}
	existingAgentConfigmap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-agent-envar-config", Namespace: config.Namespace}, existingAgentConfigmap)
	if err != nil {
		log.Error(err, "Failed to get klovercloud-ci-agent-envar-config.", err.Error())
		return ctrl.Result{}, err
	}
	redeploy := false
	redeployConfigmap := false
	if existingAgentConfigmap.Data["PULL_SIZE"] != config.Spec.Agent.PullSize {
		redeployConfigmap = true
		existingAgentConfigmap.Data["PULL_SIZE"] = config.Spec.Agent.PullSize
	}
	if existingAgentConfigmap.Data["TERMINAL_BASE_URL"] != config.Spec.Agent.TerminalBaseUrl {
		redeployConfigmap = true
		existingAgentConfigmap.Data["TERMINAL_BASE_URL"] = config.Spec.Agent.TerminalBaseUrl
	}
	if existingAgentConfigmap.Data["TERMINAL_API_VERSION"] != config.Spec.Agent.TerminalApiVersion {
		redeployConfigmap = true
		existingAgentConfigmap.Data["TERMINAL_API_VERSION"] = config.Spec.Agent.TerminalApiVersion
	}
	if existingAgentConfigmap.Data["TOKEN"] != config.Spec.Agent.Token {
		redeployConfigmap = true
		existingAgentConfigmap.Data["TOKEN"] = config.Spec.Agent.Token
	}
	for i, each := range existingAgent.Spec.Template.Spec.Containers {
		if each.Name == "app" {
			isRequestedResourcesChanged := each.Resources.Requests.Cpu() != config.Spec.Agent.Resources.Requests.Cpu() || each.Resources.Requests.Memory() != config.Spec.Agent.Resources.Requests.Memory()
			isLimitedRequestedChanged := each.Resources.Limits.Cpu() != config.Spec.Agent.Resources.Limits.Cpu() || each.Resources.Limits.Memory() != config.Spec.Agent.Resources.Limits.Memory()

			if isRequestedResourcesChanged || isLimitedRequestedChanged {
				redeploy = true
				existingAgent.Spec.Template.Spec.Containers[i].Resources = config.Spec.Agent.Resources
				break
			}

		}
	}
	if redeployConfigmap {
		err = r.Update(ctx, existingAgentConfigmap)
		if err != nil {
			log.Error(err, "Failed to update external agent configmap.", "Namespace:", existingAgent.Namespace, "Name:", existingAgent.Name)
			return ctrl.Result{}, err
		}
		redeploy=true
	}
	if redeploy {
		existingAgent.Spec.Template.ObjectMeta.Annotations = map[string]string{"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.RFC3339)}
		err = r.Update(ctx, existingAgent)
		if err != nil {
			log.Error(err, "Failed to update external agent Deployment.", "Deployment.Namespace:", existingAgent.Namespace, "Deployment.Name:", existingAgent.Name)
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
		client.MatchingLabels(map[string]string{"app": "klovercloud-ci-agent"}),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods.", "Agent.Namespace:", config.Namespace, "Agent.Name:", "klovercloud-ci-agent")
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, config.Status.AgentPods) {
		config.Status.AgentPods = podNames
		err := r.Status().Update(ctx, config)
		if err != nil {
			log.Error(err, "Failed to update AgentPods status")
			return ctrl.Result{}, err
		}
	}

	// ********************************************** Agent Finished **************************************************************
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExternalAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&basev1alpha1.ExternalAgent{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Pod{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 5}).
		Complete(r)
}
