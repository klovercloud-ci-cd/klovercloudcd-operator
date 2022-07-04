package integration_manager

import (
	"context"
	"fmt"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/service"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/utility"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/enums"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type integrationManager struct {
	Configmap  corev1.ConfigMap
	Deployment appv1.Deployment
	Service    corev1.Service
	Client     client.Client
	Error      error
}

func (i integrationManager) ModifyConfigmap(namespace string, db v1alpha1.DB, integrationManager v1alpha1.IntegrationManager) service.IntegrationManager {

	if i.Configmap.ObjectMeta.Labels == nil {
		i.Configmap.ObjectMeta.Labels = make(map[string]string)
	}
	i.Configmap.ObjectMeta.Labels["app"] = "klovercloudCD"
	i.Configmap.ObjectMeta.Namespace = namespace

	if db.Type == enums.MONGO || db.Type == "" {
		i.Configmap.Data["MONGO"] = string(enums.MONGO)
		i.Configmap.Data["MONGO_SERVER"] = db.ServerURL
		i.Configmap.Data["MONGO_PORT"] = db.ServerPort
	}

	if integrationManager.PerDayTotalProcess != "" {
		i.Configmap.Data["DEFAULT_PER_DAY_TOTAL_PROCESS"] = integrationManager.PerDayTotalProcess
	}

	if integrationManager.ConcurrentProcess != "" {
		i.Configmap.Data["DEFAULT_NUMBER_OF_CONCURRENT_PROCESS"] = integrationManager.ConcurrentProcess
	}
	i.Configmap.Data["GITHUB_WEBHOOK_CONSUMING_URL"] = integrationManager.GithubWebhookConsumingUrl

	i.Configmap.Data["BITBUCKET_WEBHOOK_CONSUMING_URL"] = integrationManager.BitbucketWebhookConsumingUrl

	if integrationManager.PipelinePurging == "DISABLE" {
		i.Configmap.Data["PIPELINE_PURGING"] = "DISABLE"
	}

	KLOVERCLOUD_CI_CORE_URL := i.Configmap.Data["KLOVERCLOUD_CI_CORE_URL"]
	replacedUrl := strings.ReplaceAll(KLOVERCLOUD_CI_CORE_URL, ".klovercloud.", "."+namespace+".")
	i.Configmap.Data["KLOVERCLOUD_CI_CORE_URL"] = replacedUrl

	EVENT_STORE_URL := i.Configmap.Data["EVENT_STORE_URL"]
	replacedUrl = strings.ReplaceAll(EVENT_STORE_URL, ".klovercloud.", "."+namespace+".")
	i.Configmap.Data["EVENT_STORE_URL"] = replacedUrl

	return i
}

func (i integrationManager) ModifyDeployment(namespace string, integrationManager v1alpha1.IntegrationManager) service.IntegrationManager {
	if i.Deployment.ObjectMeta.Labels == nil {
		i.Deployment.ObjectMeta.Labels = make(map[string]string)
	}
	i.Deployment.ObjectMeta.Labels["app"] = "klovercloudCD"
	i.Deployment.ObjectMeta.Namespace = namespace
	if integrationManager.Resources.Requests.Cpu() != nil || integrationManager.Resources.Limits.Cpu() != nil {
		for index, _ := range i.Deployment.Spec.Template.Spec.Containers {
			i.Deployment.Spec.Template.Spec.Containers[index].Resources = integrationManager.Resources
		}
	}
	return i
}

func (i integrationManager) ModifyService(namespace string) service.IntegrationManager {
	if i.Service.ObjectMeta.Labels == nil {
		i.Service.ObjectMeta.Labels = make(map[string]string)
	}
	i.Service.ObjectMeta.Labels["app"] = "klovercloudCD"
	i.Service.ObjectMeta.Namespace = namespace
	return i
}

func (i integrationManager) Apply(wait bool) error {
	if i.Error != nil {
		return i.Error
	}
	err := i.ApplyConfigMap()
	if err != nil {
		log.Println("[ERROR]: Failed to create configmap for security service.", "Deployment.Namespace", i.Deployment.Namespace, "Deployment.Name", i.Deployment.Name, err.Error())
		return err
	}

	existingPodListObject := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(i.Deployment.Namespace),
		client.MatchingLabels(i.Deployment.ObjectMeta.Labels),
	}

	if err = i.Client.List(context.Background(), existingPodListObject, listOpts...); err != nil {
		log.Println(err, "[ERROR]: Failed to list pods", "Deployment.Namespace", i.Deployment.Namespace, " Deployment.Name", i.Deployment.Name)
	}

	existingPodMap := make(map[string]bool)

	if len(existingPodListObject.Items) > 0 {
		for _, each := range existingPodListObject.Items {
			existingPodMap[each.Name] = true
		}
	}
	err = i.ApplyDeployment()
	if err != nil {
		log.Println("[ERROR]: Failed to apply deployment for security service.", "Deployment.Namespace: ", i.Deployment.Namespace, " Deployment.Name: ", i.Deployment.Name+". ", err.Error())
		return err
	}

	if wait {
		err = utility.WaitUntilPodsAreReady(i.Client, existingPodMap, listOpts, i.Deployment.Namespace, i.Deployment.Name, *i.Deployment.Spec.Replicas, 10)
		if err != nil {
			return err
		}
	}
	err = i.ApplyService()
	if err != nil {
		log.Println("[ERROR]: Failed to apply service for security service.", "Deployment.Namespace: ", i.Deployment.Namespace, " Deployment.Name: ", i.Deployment.Name+". ", err.Error())
		return err
	}
	return nil
}

func (i integrationManager) ApplyConfigMap() error {
	return i.Client.Create(context.Background(), &i.Configmap)
}

func (i integrationManager) ApplyDeployment() error {
	return i.Client.Create(context.Background(), &i.Deployment)
}

func (i integrationManager) ApplyService() error {
	return i.Client.Create(context.Background(), &i.Service)
}

func getConfigMapFromFile() corev1.ConfigMap {
	data, err := ioutil.ReadFile("integration-manager-configmap.yaml")
	if err != nil {
		panic(err.Error())
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*corev1.ConfigMap)
}

func getServiceFromFile() corev1.Service {
	data, err := ioutil.ReadFile("integration-manager-service.yaml")
	if err != nil {
		panic(err.Error())
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*corev1.Service)
}

func getDeploymentFromFile() appv1.Deployment {
	data, err := ioutil.ReadFile("integration-manager-deployment.yaml")
	if err != nil {
		panic(err.Error())
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*appv1.Deployment)
}

func New(client client.Client) service.IntegrationManager {
	return integrationManager{
		Configmap:  getConfigMapFromFile(),
		Deployment: getDeploymentFromFile(),
		Service:    getServiceFromFile(),
		Client:     client,
	}
}
