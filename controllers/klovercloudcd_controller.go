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
	"strconv"

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
	securityServiceRedeploy:=false

	isSecretChanged:=existingMongoSecret.StringData["MONGO_USERNAME"]!=config.Spec.Database.UserName || existingMongoSecret.StringData["MONGO_PASSWORD"]!=config.Spec.Database.Password
	if isSecretChanged{
		redeploy=true
		existingMongoSecret.StringData["MONGO_USERNAME"]=config.Spec.Database.UserName
		existingMongoSecret.StringData["MONGO_PASSWORD"]=config.Spec.Database.Password
		err = r.Update(ctx, existingMongoSecret)
		if err != nil {
			log.Error(err, "Failed to update Mongo Secret.", err.Error())
			return ctrl.Result{}, err
		}
	}

	existingSecurityServerConfigmap:=&corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-security-envar-config", Namespace: config.Namespace}, existingSecurityServerConfigmap)
	if err != nil {
		log.Error(err, "Failed to get klovercloud-mongo-secret.",err.Error())
		return ctrl.Result{}, err
	}
	if existingSecurityServerConfigmap.Data["MAIL_SERVER_HOST_EMAIL"]!=config.Spec.Security.MailServerHostEmail{
		securityServiceRedeploy=true
		existingSecurityServerConfigmap.Data["MAIL_SERVER_HOST_EMAIL"]=config.Spec.Security.MailServerHostEmail
	}
	if existingSecurityServerConfigmap.Data["MAIL_SERVER_HOST_EMAIL_SECRET"]!=config.Spec.Security.MailServerHostEmailSecret{
		securityServiceRedeploy=true
		existingSecurityServerConfigmap.Data["MAIL_SERVER_HOST_EMAIL_SECRET"]=config.Spec.Security.MailServerHostEmailSecret
	}
	if existingSecurityServerConfigmap.Data["SMTP_HOST"]!=config.Spec.Security.SMTPHost{
		securityServiceRedeploy=true
		existingSecurityServerConfigmap.Data["SMTP_HOST"]=config.Spec.Security.SMTPHost
	}
	if existingSecurityServerConfigmap.Data["SMTP_PORT"]!=config.Spec.Security.SMTPPort{
		securityServiceRedeploy=true
		existingSecurityServerConfigmap.Data["SMTP_PORT"]=config.Spec.Security.SMTPPort
	}
	if existingSecurityServerConfigmap.Data["USER_FIRST_NAME"]!=config.Spec.Security.User.FirstName{
		securityServiceRedeploy=true
		existingSecurityServerConfigmap.Data["USER_FIRST_NAME"]=config.Spec.Security.User.FirstName
	}
	if existingSecurityServerConfigmap.Data["USER_LAST_NAME"]!=config.Spec.Security.User.LastName{
		securityServiceRedeploy=true
		existingSecurityServerConfigmap.Data["USER_LAST_NAME"]=config.Spec.Security.User.LastName
	}
	if existingSecurityServerConfigmap.Data["USER_EMAIL"]!=config.Spec.Security.User.Email{
		securityServiceRedeploy=true
		existingSecurityServerConfigmap.Data["USER_EMAIL"]=config.Spec.Security.User.Email
	}
	if existingSecurityServerConfigmap.Data["USER_PHONE"]!=config.Spec.Security.User.Phone{
		securityServiceRedeploy=true
		existingSecurityServerConfigmap.Data["USER_PHONE"]=config.Spec.Security.User.Phone
	}
	if existingSecurityServerConfigmap.Data["USER_PASSWORD"]!=config.Spec.Security.User.Password{
		securityServiceRedeploy=true
		existingSecurityServerConfigmap.Data["USER_PASSWORD"]=config.Spec.Security.User.Password
	}
	if existingSecurityServerConfigmap.Data["COMPANY_NAME"]!=config.Spec.Security.User.CompanyName{
		securityServiceRedeploy=true
		existingSecurityServerConfigmap.Data["COMPANY_NAME"]=config.Spec.Security.User.CompanyName
	}

	if redeploy{
		err = r.Update(ctx, existingMongoSecret)
		if err != nil {
			log.Error(err, "Failed to update mongo secret.", err.Error())
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	if securityServiceRedeploy{
		err = r.Update(ctx, existingSecurityServerConfigmap)
		if err != nil {
			log.Error(err, "Failed to update Security Servers Configmap.", err.Error())
			return ctrl.Result{}, err
		}
		existingSecurity := &appsv1.Deployment{}
		err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-security", Namespace: config.Namespace}, existingSecurity)
		if err != nil && errors.IsNotFound(err) {
			log.Info("No security deploy found. Namespace:",existingSecurity.Namespace," Name:",existingSecurity.Name)
		}
		r.Update(ctx, existingSecurity)
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


	// ********************************************** All About Integration Manager ****************************************************

	// Apply integration manager
	// Check if the deployment already exists, if not create a new one
	existingIntegrationManager := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-integration-manager", Namespace: config.Namespace}, existingIntegrationManager)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		err = descriptor.ApplyIntegrationManager(r.Client, config.Namespace, config.Spec.Database, config.Spec.IntegrationManager, string(config.Spec.Version))
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", config.Namespace, "Deployment.Name", "klovercloud-integration-manager")
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	existingIntegrationManagerConfigmap:=&corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-integration-manager-envar-config", Namespace: config.Namespace}, existingIntegrationManagerConfigmap)
	if err != nil {
		log.Error(err, "Failed to get klovercloud-integration-manager-envar-config.",err.Error())
		return ctrl.Result{}, err
	}
	redeployConfigmap:=false
	redeploy=false

	if existingIntegrationManagerConfigmap.Data["MONGO_SERVER"]!=config.Spec.Database.ServerURL{
		redeployConfigmap=true
		existingIntegrationManagerConfigmap.Data["MONGO_SERVER"]=config.Spec.Database.ServerURL
	}

	if existingIntegrationManagerConfigmap.Data["MONGO_PORT"]!=config.Spec.Database.ServerPort{
		redeployConfigmap=true
		existingIntegrationManagerConfigmap.Data["MONGO_PORT"]=config.Spec.Database.ServerPort
	}

	if existingIntegrationManagerConfigmap.Data["DEFAULT_PER_DAY_TOTAL_PROCESS"]!=config.Spec.IntegrationManager.PerDayTotalProcess{
		redeployConfigmap=true
		existingIntegrationManagerConfigmap.Data["DEFAULT_PER_DAY_TOTAL_PROCESS"]=config.Spec.IntegrationManager.PerDayTotalProcess
	}
	if existingIntegrationManagerConfigmap.Data["DEFAULT_NUMBER_OF_CONCURRENT_PROCESS"]!=config.Spec.IntegrationManager.ConcurrentProcess{
		redeployConfigmap=true
		existingIntegrationManagerConfigmap.Data["DEFAULT_NUMBER_OF_CONCURRENT_PROCESS"]=config.Spec.IntegrationManager.ConcurrentProcess
	}
	if existingIntegrationManagerConfigmap.Data["GITHUB_WEBHOOK_CONSUMING_URL"]!=config.Spec.IntegrationManager.GithubWebhookConsumingUrl{
		redeployConfigmap=true
		existingIntegrationManagerConfigmap.Data["GITHUB_WEBHOOK_CONSUMING_URL"]=config.Spec.IntegrationManager.GithubWebhookConsumingUrl
	}
	if existingIntegrationManagerConfigmap.Data["GITHUB_WEBHOOK_CONSUMING_URL"]!=config.Spec.IntegrationManager.GithubWebhookConsumingUrl{
		redeployConfigmap=true
		existingIntegrationManagerConfigmap.Data["GITHUB_WEBHOOK_CONSUMING_URL"]=config.Spec.IntegrationManager.GithubWebhookConsumingUrl
	}
	if existingIntegrationManagerConfigmap.Data["BITBUCKET_WEBHOOK_CONSUMING_URL"]!=config.Spec.IntegrationManager.BitbucketWebhookConsumingUrl{
		redeployConfigmap=true
		existingIntegrationManagerConfigmap.Data["BITBUCKET_WEBHOOK_CONSUMING_URL"]=config.Spec.IntegrationManager.BitbucketWebhookConsumingUrl
	}
	if existingIntegrationManagerConfigmap.Data["BITBUCKET_WEBHOOK_CONSUMING_URL"]!=config.Spec.IntegrationManager.BitbucketWebhookConsumingUrl{
		redeployConfigmap=true
		existingIntegrationManagerConfigmap.Data["BITBUCKET_WEBHOOK_CONSUMING_URL"]=config.Spec.IntegrationManager.BitbucketWebhookConsumingUrl
	}
	if *existingIntegrationManager.Spec.Replicas != config.Spec.IntegrationManager.Size {
		redeploy=true
		existingIntegrationManager.Spec.Replicas = &config.Spec.IntegrationManager.Size
	}


	for i,each:= range existingIntegrationManager.Spec.Template.Spec.Containers{
		if each.Name=="app"{
			isRequestedResourcesChanged:=each.Resources.Requests.Cpu()!=config.Spec.IntegrationManager.Resources.Requests.Cpu() || each.Resources.Requests.Memory()!=config.Spec.IntegrationManager.Resources.Requests.Memory()
			isLimitedRequestedChanged:=each.Resources.Limits.Cpu()!=config.Spec.IntegrationManager.Resources.Limits.Cpu() || each.Resources.Limits.Memory()!=config.Spec.IntegrationManager.Resources.Limits.Memory()

			if isRequestedResourcesChanged || isLimitedRequestedChanged{
				redeploy=true
				existingIntegrationManager.Spec.Template.Spec.Containers[i].Resources=config.Spec.ApiService.Resources
				break
			}

		}
	}

	if redeployConfigmap{
		err = r.Update(ctx, existingIntegrationManagerConfigmap)
		if err != nil {
			log.Error(err, "Failed to update Configmap.", "Namespace:", existingIntegrationManagerConfigmap.Namespace, "Name:", existingIntegrationManagerConfigmap.Name)
			return ctrl.Result{}, err
		}
	}
	if redeploy{
		err = r.Update(ctx, existingIntegrationManager)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace:", existingIntegrationManager.Namespace, "Deployment.Name:", existingIntegrationManager.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the IntegrationManager status with the pod names
	// List the pods for this api service's deployment
	podList = &corev1.PodList{}
	listOpts = []client.ListOption{
		client.InNamespace(config.Namespace),
		client.MatchingLabels(map[string]string{"app":"klovercloud-integration-manager"}),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods.", "IntegrationManager.Namespace:", config.Namespace, "IntegrationManager.Name:","klovercloud-integration-manager")
		return ctrl.Result{}, err
	}
	podNames = getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, config.Status.IntegrationManagerPods) {
		config.Status.IntegrationManagerPods = podNames
		err := r.Status().Update(ctx, config)
		if err != nil {
			log.Error(err, "Failed to update IntegrationManager status")
			return ctrl.Result{}, err
		}
	}

	// ********************************************** Integration Manager Finished **************************************************************


	// ********************************************** All About Event Bank ****************************************************

	// Apply event bank
	// Check if the deployment already exists, if not create a new one
	existingEventBank := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-event-bank", Namespace: config.Namespace}, existingEventBank)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		err = descriptor.ApplyEventBank(r.Client, config.Namespace, config.Spec.Database, config.Spec.EventBank, string(config.Spec.Version))
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", config.Namespace, "Deployment.Name", "klovercloud-ci-event-bank")
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}


	existingEventBankConfigmap:=&corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-event-bank-envar-config", Namespace: config.Namespace}, existingEventBankConfigmap)
	if err != nil {
		log.Error(err, "Failed to get klovercloud-ci-event-bank-envar-config.",err.Error())
		return ctrl.Result{}, err
	}

	redeploy=false
	redeployConfigmap=false
	if existingEventBankConfigmap.Data["MONGO_SERVER"]!=config.Spec.Database.ServerURL{
		redeployConfigmap=true
		existingEventBankConfigmap.Data["MONGO_SERVER"]=config.Spec.Database.ServerURL
	}

	if existingEventBankConfigmap.Data["MONGO_PORT"]!=config.Spec.Database.ServerPort{
		redeployConfigmap=true
		existingEventBankConfigmap.Data["MONGO_PORT"]=config.Spec.Database.ServerPort
	}

	if *existingEventBank.Spec.Replicas != config.Spec.EventBank.Size {
		redeploy=true
		existingEventBank.Spec.Replicas = &config.Spec.EventBank.Size
	}


	for i,each:= range existingEventBank.Spec.Template.Spec.Containers{
		if each.Name=="app"{
			isRequestedResourcesChanged:=each.Resources.Requests.Cpu()!=config.Spec.EventBank.Resources.Requests.Cpu() || each.Resources.Requests.Memory()!=config.Spec.EventBank.Resources.Requests.Memory()
			isLimitedRequestedChanged:=each.Resources.Limits.Cpu()!=config.Spec.EventBank.Resources.Limits.Cpu() || each.Resources.Limits.Memory()!=config.Spec.EventBank.Resources.Limits.Memory()

			if isRequestedResourcesChanged || isLimitedRequestedChanged{
				redeploy=true
				existingEventBank.Spec.Template.Spec.Containers[i].Resources=config.Spec.EventBank.Resources
				break
			}

		}
	}
	if redeployConfigmap{
		err = r.Update(ctx, existingEventBankConfigmap)
		if err != nil {
			log.Error(err, "Failed to update Configmap.", "Namespace:", existingEventBankConfigmap.Namespace, "Name:", existingEventBankConfigmap.Name)
			return ctrl.Result{}, err
		}
	}

	if redeploy{
		err = r.Update(ctx, existingEventBank)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace:", existingEventBank.Namespace, "Deployment.Name:", existingEventBank.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}


	// Update the EventBank status with the pod names
	// List the pods for this api service's deployment
	podList = &corev1.PodList{}
	listOpts = []client.ListOption{
		client.InNamespace(config.Namespace),
		client.MatchingLabels(map[string]string{"app":"klovercloud-ci-event-bank"}),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods.", "EventBank.Namespace:", config.Namespace, "EventBank.Name:","klovercloud-ci-event-bank")
		return ctrl.Result{}, err
	}
	podNames = getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, config.Status.EventBankPods) {
		config.Status.EventBankPods = podNames
		err := r.Status().Update(ctx, config)
		if err != nil {
			log.Error(err, "Failed to update EventBank status")
			return ctrl.Result{}, err
		}
	}

	// ********************************************** Event Bank Finished **************************************************************


	// ********************************************** All About Core Engine ****************************************************

	// Apply core engine
	existingCoreEngine := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-core", Namespace: config.Namespace}, existingCoreEngine)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		err=descriptor.ApplyCoreEngine(r.Client,config.Namespace,config.Spec.Database,config.Spec.CoreEngine,string(config.Spec.Version))
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", config.Namespace, "Deployment.Name", "klovercloud-ci-core")
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	existingCoreEngineConfigmap:=&corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-core-envar-config", Namespace: config.Namespace}, existingCoreEngineConfigmap)
	if err != nil {
		log.Error(err, "Failed to get klovercloud-ci-core-envar-config.",err.Error())
		return ctrl.Result{}, err
	}
	redeploy=false
	if existingCoreEngineConfigmap.Data["ALLOWED_CONCURRENT_BUILD"]!=strconv.Itoa(config.Spec.CoreEngine.NumberOfConCurrentProcess){
		redeploy=true
		existingCoreEngineConfigmap.Data["ALLOWED_CONCURRENT_BUILD"]=strconv.Itoa(config.Spec.CoreEngine.NumberOfConCurrentProcess)
	}


	if *existingCoreEngine.Spec.Replicas != config.Spec.CoreEngine.Size {
		redeploy=true
		existingCoreEngine.Spec.Replicas = &config.Spec.CoreEngine.Size
	}


	for i,each:= range existingCoreEngine.Spec.Template.Spec.Containers{
		if each.Name=="app"{
			isRequestedResourcesChanged:=each.Resources.Requests.Cpu()!=config.Spec.CoreEngine.Resources.Requests.Cpu() || each.Resources.Requests.Memory()!=config.Spec.CoreEngine.Resources.Requests.Memory()
			isLimitedRequestedChanged:=each.Resources.Limits.Cpu()!=config.Spec.CoreEngine.Resources.Limits.Cpu() || each.Resources.Limits.Memory()!=config.Spec.CoreEngine.Resources.Limits.Memory()

			if isRequestedResourcesChanged || isLimitedRequestedChanged{
				redeploy=true
				existingCoreEngine.Spec.Template.Spec.Containers[i].Resources=config.Spec.CoreEngine.Resources
				break
			}

		}
	}

	if redeploy{
		err = r.Update(ctx, existingCoreEngineConfigmap)
		if err != nil {
			log.Error(err, "Failed to update Configmap.", "Namespace:", existingCoreEngineConfigmap.Namespace, "Name:", existingCoreEngineConfigmap.Name)
			return ctrl.Result{}, err
		}
		err = r.Update(ctx, existingCoreEngine)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace:", existingCoreEngine.Namespace, "Deployment.Name:", existingCoreEngine.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the CoreEngine status with the pod names
	// List the pods for this api service's deployment
	podList = &corev1.PodList{}
	listOpts = []client.ListOption{
		client.InNamespace(config.Namespace),
		client.MatchingLabels(map[string]string{"app":"klovercloud-ci-core"}),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods.", "CoreEngine.Namespace:", config.Namespace, "CoreEngine.Name:","klovercloud-ci-core")
		return ctrl.Result{}, err
	}
	podNames = getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, config.Status.CoreEnginePods) {
		config.Status.CoreEnginePods = podNames
		err := r.Status().Update(ctx, config)
		if err != nil {
			log.Error(err, "Failed to update CoreEngine status")
			return ctrl.Result{}, err
		}
	}

	// ********************************************** Core Engine Finished **************************************************************


	// ********************************************** All About Security ***************************************************************

	// Apply security
	existingSecurity := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-security", Namespace: config.Namespace}, existingSecurity)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		err = descriptor.ApplySecurity(r.Client, config.Namespace, config.Spec.Database, config.Spec.Security, string(config.Spec.Version))
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", config.Namespace, "Deployment.Name", "klovercloud-security")
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}
	redeploy=false
	if *existingSecurity.Spec.Replicas != config.Spec.Security.Size {
		redeploy=true
		existingSecurity.Spec.Replicas = &config.Spec.Security.Size
	}

	if redeploy{
		err = r.Update(ctx, existingSecurity)
		if err != nil {
			log.Error(err, "Failed to update Security.", "Namespace:", existingSecurity.Namespace, "Name:", existingSecurity.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the Security status with the pod names
	// List the pods for this api service's deployment
	podList = &corev1.PodList{}
	listOpts = []client.ListOption{
		client.InNamespace(config.Namespace),
		client.MatchingLabels(map[string]string{"app":"klovercloud-security"}),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods.", "Security.Namespace:", config.Namespace, "Security.Name:","klovercloud-security")
		return ctrl.Result{}, err
	}
	podNames = getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, config.Status.SecurityPods) {
		config.Status.SecurityPods = podNames
		err := r.Status().Update(ctx, config)
		if err != nil {
			log.Error(err, "Failed to update Security status")
			return ctrl.Result{}, err
		}
	}

	// ********************************************** Security Finished **************************************************************

	// Apply lighthouse
	if config.Spec.Agent.LightHouseEnabled=="true"{


		// ********************************************** All About Lighthouse Command ***************************************************************
		existingLightHouseCommand := &appsv1.Deployment{}
		err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-light-house-command", Namespace: config.Namespace}, existingLightHouseCommand)
		if err != nil && errors.IsNotFound(err) {
			// Define a new deployment
			err=descriptor.ApplyLightHouseCommand(r.Client,config.Namespace,config.Spec.Database,config.Spec.LightHouse.Command,string(config.Spec.Version))
			if err != nil {
				log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", config.Namespace, "Deployment.Name", "klovercloud-ci-light-house-command")
				return ctrl.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}

		existingLightHouseCommandConfigmap:=&corev1.ConfigMap{}
		err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-light-house-command-config", Namespace: config.Namespace}, existingLightHouseCommandConfigmap)
		if err != nil {
			log.Error(err, "Failed to get klovercloud-ci-light-house-command-config.",err.Error())
			return ctrl.Result{}, err
		}

		redeploy=false
		redeployConfigmap=false
		if existingLightHouseCommandConfigmap.Data["MONGO_SERVER"]!=config.Spec.Database.ServerURL{
			redeployConfigmap=true
			existingLightHouseCommandConfigmap.Data["MONGO_SERVER"]=config.Spec.Database.ServerURL
		}

		if existingLightHouseCommandConfigmap.Data["MONGO_PORT"]!=config.Spec.Database.ServerPort{
			redeployConfigmap=true
			existingLightHouseCommandConfigmap.Data["MONGO_PORT"]=config.Spec.Database.ServerPort
		}

		if *existingLightHouseCommand.Spec.Replicas != config.Spec.LightHouse.Command.Size {
			redeploy=true
			existingLightHouseCommand.Spec.Replicas = &config.Spec.LightHouse.Command.Size
		}


		for i,each:= range existingLightHouseCommand.Spec.Template.Spec.Containers{
			if each.Name=="app"{
				isRequestedResourcesChanged:=each.Resources.Requests.Cpu()!=config.Spec.LightHouse.Command.Resources.Requests.Cpu() || each.Resources.Requests.Memory()!=config.Spec.LightHouse.Command.Resources.Requests.Memory()
				isLimitedRequestedChanged:=each.Resources.Limits.Cpu()!=config.Spec.LightHouse.Command.Resources.Limits.Cpu() || each.Resources.Limits.Memory()!=config.Spec.LightHouse.Command.Resources.Limits.Memory()

				if isRequestedResourcesChanged || isLimitedRequestedChanged{
					redeploy=true
					existingLightHouseCommand.Spec.Template.Spec.Containers[i].Resources=config.Spec.LightHouse.Command.Resources
					break
				}

			}
		}
		if redeployConfigmap{
			err = r.Update(ctx, existingLightHouseCommand)
			if err != nil {
				log.Error(err, "Failed to update Configmap.", "Namespace:", existingLightHouseCommand.Namespace, "Name:", existingLightHouseCommand.Name)
				return ctrl.Result{}, err
			}
		}

		if redeploy{
			err = r.Update(ctx, existingLightHouseCommand)
			if err != nil {
				log.Error(err, "Failed to update Deployment.", "Deployment.Namespace:", existingLightHouseCommand.Namespace, "Deployment.Name:", existingLightHouseCommand.Name)
				return ctrl.Result{}, err
			}
			// Spec updated - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}


		// Update the LightHouseCommand status with the pod names
		// List the pods for this api service's deployment
		podList = &corev1.PodList{}
		listOpts = []client.ListOption{
			client.InNamespace(config.Namespace),
			client.MatchingLabels(map[string]string{"app":"klovercloud-ci-light-house-command"}),
		}
		if err = r.List(ctx, podList, listOpts...); err != nil {
			log.Error(err, "Failed to list pods.", "LightHouseCommand.Namespace:", config.Namespace, "LightHouseCommand.Name:","klovercloud-ci-light-house-command")
			return ctrl.Result{}, err
		}
		podNames = getPodNames(podList.Items)

		// Update status.Nodes if needed
		if !reflect.DeepEqual(podNames, config.Status.LightHouseCommandPods) {
			config.Status.EventBankPods = podNames
			err := r.Status().Update(ctx, config)
			if err != nil {
				log.Error(err, "Failed to update LightHouseCommandPod status")
				return ctrl.Result{}, err
			}
		}


		// ********************************************** Lighthouse Command Finished **************************************************************


		// ********************************************** All About Lighthouse Query ***************************************************************
		existingLightHouseQuery := &appsv1.Deployment{}
		err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-light-house-query", Namespace: config.Namespace}, existingLightHouseQuery)
		if err != nil && errors.IsNotFound(err) {
			// Define a new deployment
			err=descriptor.ApplyLightHouseQuery(r.Client,config.Namespace,config.Spec.Database,config.Spec.LightHouse.Query,string(config.Spec.Version))
			if err != nil {
				log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", config.Namespace, "Deployment.Name", "klovercloud-ci-light-house-query")
				return ctrl.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}

		existingLightHouseQueryConfigmap:=&corev1.ConfigMap{}
		err = r.Get(ctx, types.NamespacedName{Name: "klovercloud-ci-light-house-query-config", Namespace: config.Namespace}, existingLightHouseQueryConfigmap)
		if err != nil {
			log.Error(err, "Failed to get klovercloud-ci-light-house-query-config.",err.Error())
			return ctrl.Result{}, err
		}

		redeploy=false
		redeployConfigmap=false
		if existingLightHouseQueryConfigmap.Data["MONGO_SERVER"]!=config.Spec.Database.ServerURL{
			redeployConfigmap=true
			existingLightHouseQueryConfigmap.Data["MONGO_SERVER"]=config.Spec.Database.ServerURL
		}

		if existingLightHouseQueryConfigmap.Data["MONGO_PORT"]!=config.Spec.Database.ServerPort{
			redeployConfigmap=true
			existingLightHouseQueryConfigmap.Data["MONGO_PORT"]=config.Spec.Database.ServerPort
		}

		if *existingLightHouseQuery.Spec.Replicas != config.Spec.LightHouse.Query.Size {
			redeploy=true
			existingLightHouseQuery.Spec.Replicas = &config.Spec.LightHouse.Query.Size
		}


		for i,each:= range existingLightHouseQuery.Spec.Template.Spec.Containers{
			if each.Name=="app"{
				isRequestedResourcesChanged:=each.Resources.Requests.Cpu()!=config.Spec.LightHouse.Query.Resources.Requests.Cpu() || each.Resources.Requests.Memory()!=config.Spec.LightHouse.Query.Resources.Requests.Memory()
				isLimitedRequestedChanged:=each.Resources.Limits.Cpu()!=config.Spec.LightHouse.Query.Resources.Limits.Cpu() || each.Resources.Limits.Memory()!=config.Spec.LightHouse.Query.Resources.Limits.Memory()

				if isRequestedResourcesChanged || isLimitedRequestedChanged{
					redeploy=true
					existingLightHouseQuery.Spec.Template.Spec.Containers[i].Resources=config.Spec.LightHouse.Query.Resources
					break
				}

			}
		}
		if redeployConfigmap{
			err = r.Update(ctx, existingLightHouseQuery)
			if err != nil {
				log.Error(err, "Failed to update Configmap.", "Namespace:", existingLightHouseQuery.Namespace, "Name:", existingLightHouseQuery.Name)
				return ctrl.Result{}, err
			}
		}

		if redeploy{
			err = r.Update(ctx, existingLightHouseQuery)
			if err != nil {
				log.Error(err, "Failed to update Deployment.", "Deployment.Namespace:", existingLightHouseQuery.Namespace, "Deployment.Name:", existingLightHouseQuery.Name)
				return ctrl.Result{}, err
			}
			// Spec updated - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}


		// Update the LightHouseQuery status with the pod names
		// List the pods for this api service's deployment
		podList = &corev1.PodList{}
		listOpts = []client.ListOption{
			client.InNamespace(config.Namespace),
			client.MatchingLabels(map[string]string{"app":"klovercloud-ci-light-house-query"}),
		}
		if err = r.List(ctx, podList, listOpts...); err != nil {
			log.Error(err, "Failed to list pods.", "LightHouseQuery.Namespace:", config.Namespace, "LightHouseQuery.Name:","klovercloud-ci-light-house-query")
			return ctrl.Result{}, err
		}
		podNames = getPodNames(podList.Items)

		// Update status.Nodes if needed
		if !reflect.DeepEqual(podNames, config.Status.LightHouseQueryPods) {
			config.Status.LightHouseQueryPods = podNames
			err := r.Status().Update(ctx, config)
			if err != nil {
				log.Error(err, "Failed to update LightHouseQueryPod status")
				return ctrl.Result{}, err
			}
		}

		// ********************************************** Lighthouse Query Finished **************************************************************
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
