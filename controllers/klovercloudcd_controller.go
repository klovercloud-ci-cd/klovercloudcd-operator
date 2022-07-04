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
	"k8s.io/client-go/rest"
	"reflect"
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


	// ********************************************** All About Prerequisites ****************************************************
	existingMongoSecret:=&corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-mongo-secret", Namespace: config.Namespace}, existingMongoSecret)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		err = descriptor.ApplyPrerequisites(r.Client, config.Namespace, config.Spec.Database, config.Spec.Security, string(config.Spec.Version))
		if err != nil {
			log.Error(err, "Failed to apply Prerequisites.",err.Error())
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get klovercloud-mongo-secret")
		return ctrl.Result{}, err
	}

	redeploy:=false

	isSecretChanged:=existingMongoSecret.StringData["MONGO_USERNAME"]!=config.Spec.Database.UserName || existingMongoSecret.StringData["MONGO_PASSWORD"]!=config.Spec.Database.Password
	if isSecretChanged{
		redeploy=true
		existingMongoSecret.StringData["MONGO_USERNAME"]=config.Spec.Database.UserName
		existingMongoSecret.StringData["MONGO_PASSWORD"]=config.Spec.Database.Password
	}

	existingSecurityServerConfigmap:=&corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-security-envar-config", Namespace: config.Namespace}, existingSecurityServerConfigmap)

	if existingSecurityServerConfigmap.Data["MAIL_SERVER_HOST_EMAIL"]!=config.Spec.Security.MailServerHostEmail{
		redeploy=true
		existingSecurityServerConfigmap.Data["MAIL_SERVER_HOST_EMAIL"]=config.Spec.Security.MailServerHostEmail
	}
	if existingSecurityServerConfigmap.Data["MAIL_SERVER_HOST_EMAIL_SECRET"]!=config.Spec.Security.MailServerHostEmailSecret{
		redeploy=true
		existingSecurityServerConfigmap.Data["MAIL_SERVER_HOST_EMAIL_SECRET"]=config.Spec.Security.MailServerHostEmailSecret
	}
	if existingSecurityServerConfigmap.Data["SMTP_HOST"]!=config.Spec.Security.SMTPHost{
		redeploy=true
		existingSecurityServerConfigmap.Data["SMTP_HOST"]=config.Spec.Security.SMTPHost
	}
	if existingSecurityServerConfigmap.Data["SMTP_PORT"]!=config.Spec.Security.SMTPPort{
		redeploy=true
		existingSecurityServerConfigmap.Data["SMTP_PORT"]=config.Spec.Security.SMTPPort
	}
	if existingSecurityServerConfigmap.Data["USER_FIRST_NAME"]!=config.Spec.Security.User.FirstName{
		redeploy=true
		existingSecurityServerConfigmap.Data["USER_FIRST_NAME"]=config.Spec.Security.User.FirstName
	}
	if existingSecurityServerConfigmap.Data["USER_LAST_NAME"]!=config.Spec.Security.User.LastName{
		redeploy=true
		existingSecurityServerConfigmap.Data["USER_LAST_NAME"]=config.Spec.Security.User.LastName
	}
	if existingSecurityServerConfigmap.Data["USER_EMAIL"]!=config.Spec.Security.User.Email{
		redeploy=true
		existingSecurityServerConfigmap.Data["USER_EMAIL"]=config.Spec.Security.User.Email
	}
	if existingSecurityServerConfigmap.Data["USER_PHONE"]!=config.Spec.Security.User.Phone{
		redeploy=true
		existingSecurityServerConfigmap.Data["USER_PHONE"]=config.Spec.Security.User.Phone
	}
	if existingSecurityServerConfigmap.Data["USER_PASSWORD"]!=config.Spec.Security.User.Password{
		redeploy=true
		existingSecurityServerConfigmap.Data["USER_PASSWORD"]=config.Spec.Security.User.Password
	}
	if existingSecurityServerConfigmap.Data["COMPANY_NAME"]!=config.Spec.Security.User.CompanyName{
		redeploy=true
		existingSecurityServerConfigmap.Data["COMPANY_NAME"]=config.Spec.Security.User.CompanyName
	}

	if redeploy{
		err = r.Update(ctx, existingSecurityServerConfigmap)
		if err != nil {
			log.Error(err, "Failed to update Security Servers Configmap.", err.Error())
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}


	// ********************************************** Prerequisites Finished **************************************************************



	// ********************************************** All About Api Service ****************************************************
	// Apply api service
	// Check if the deployment already exists, if not create a new one
	existingApiService := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-api-service", Namespace: config.Namespace}, existingApiService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		err=descriptor.ApplyApiService(r.Client,config.Namespace,config.Spec.ApiService,string(config.Spec.Version))
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", config.Namespace, "Deployment.Name", "klovercloud-api-service")
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size  and other fields are same as the spec,

	redeploy=false

	if *existingApiService.Spec.Replicas != config.Spec.ApiService.Size {
		redeploy=true
		existingApiService.Spec.Replicas = &config.Spec.ApiService.Size
	}

	for i,each:= range existingApiService.Spec.Template.Spec.Containers{
		if each.Name=="app"{
			isRequestedResourcesChanged:=each.Resources.Requests.Cpu()!=config.Spec.ApiService.Resources.Requests.Cpu() || each.Resources.Requests.Memory()!=config.Spec.ApiService.Resources.Requests.Memory()
			isLimitedRequestedChanged:=each.Resources.Limits.Cpu()!=config.Spec.ApiService.Resources.Limits.Cpu() || each.Resources.Limits.Memory()!=config.Spec.ApiService.Resources.Limits.Memory()

			if isRequestedResourcesChanged || isLimitedRequestedChanged{
				redeploy=true
				existingApiService.Spec.Template.Spec.Containers[i].Resources=config.Spec.ApiService.Resources
				break
			}

		}
	}
	if redeploy{
		err = r.Update(ctx, existingApiService)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace:", existingApiService.Namespace, "Deployment.Name:", existingApiService.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}



	// Update the ApiService status with the pod names
	// List the pods for this api service's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(config.Namespace),
		client.MatchingLabels(map[string]string{"app":"klovercloud-api-service"}),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods.", "ApiService.Namespace:", config.Namespace, "ApiService.Name:","klovercloud-api-service")
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, config.Status.ApiServicePods) {
		config.Status.ApiServicePods = podNames
		err := r.Status().Update(ctx, config)
		if err != nil {
			log.Error(err, "Failed to update ApiService status")
			return ctrl.Result{}, err
		}
	}

    // **********************************************  Api Service Finished **************************************************************


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

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}


// SetupWithManager sets up the controller with the Manager.
func (r *KlovercloudCDReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&basev1alpha1.KlovercloudCD{}).
		Complete(r)
}
