package api_service

import (
	"context"
	"fmt"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/service"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/utility"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	basev1alpha1 "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
)

type apiService struct {
	Configmap  corev1.ConfigMap
	Deployment appv1.Deployment
	Service    corev1.Service
	Client     client.Client
	Error      error
}

func (a apiService) ModifyConfigmap(namespace string) service.ApiService {
	found := &corev1.ConfigMap{}
	_ = a.Client.Get(context.Background(), types.NamespacedName{Name: "klovercloud-api-service-envar-config", Namespace: namespace}, found)
	if found.Name != "" {
		a.Configmap.Data["PRIVATE_KEY_INTERNAL_CALL"] = found.Data["PRIVATE_KEY_INTERNAL_CALL"]
		a.Configmap.Data["PUBLIC_KEY_INTERNAL_CALL"] = found.Data["PUBLIC_KEY_INTERNAL_CALL"]
	} else {
		private, public, err := utility.New().Generate()
		if err != nil {
			a.Error = err
			log.Println("[ERROR]: Failed to modify secrets configmap." + err.Error())
		}
		a.Configmap.Data["PRIVATE_KEY_INTERNAL_CALL"] = string(private)
		a.Configmap.Data["PUBLIC_KEY_INTERNAL_CALL"] = string(public)
	}
	if a.Configmap.ObjectMeta.Labels == nil {
		a.Configmap.ObjectMeta.Labels = make(map[string]string)
	}
	a.Configmap.ObjectMeta.Labels["app"] = "klovercloudCD"
	a.Configmap.ObjectMeta.Namespace = namespace
	found = &corev1.ConfigMap{}
	_ = a.Client.Get(context.Background(), types.NamespacedName{Name: "klovercloud-security-envar-config", Namespace: namespace}, found)
	a.Configmap.Data["PUBLIC_KEY"] = found.Data["PRIVATE_KEY"]

	integrationManagerUrl := a.Configmap.Data["KLOVERCLOUD_CI_INTEGRATION_MANAGER_URL"]
	replacedUrl := strings.ReplaceAll(integrationManagerUrl, ".klovercloud.", "."+namespace+".")
	a.Configmap.Data["KLOVERCLOUD_CI_INTEGRATION_MANAGER_URL"] = replacedUrl

	eventStoreUrl := a.Configmap.Data["KLOVERCLOUD_CI_EVENT_STORE"]
	replacedUrl = strings.ReplaceAll(eventStoreUrl, ".klovercloud.", "."+namespace+".")
	a.Configmap.Data["KLOVERCLOUD_CI_EVENT_STORE"] = replacedUrl

	wsEventStoreUrl := a.Configmap.Data["KLOVERCLOUD_CI_EVENT_STORE_WS"]
	replacedUrl = strings.ReplaceAll(wsEventStoreUrl, ".klovercloud.", "."+namespace+".")
	a.Configmap.Data["KLOVERCLOUD_CI_EVENT_STORE_WS"] = replacedUrl

	lightHouseCommandUrl := a.Configmap.Data["LIGHTHOUSE_COMMAND_SERVER_URL"]
	replacedUrl = strings.ReplaceAll(lightHouseCommandUrl, ".klovercloud.", "."+namespace+".")
	a.Configmap.Data["LIGHTHOUSE_COMMAND_SERVER_URL"] = replacedUrl

	lightHouseQuery := a.Configmap.Data["LIGHTHOUSE_QUERY_SERVER_URL"]
	a.Configmap.Data["LIGHTHOUSE_QUERY_SERVER_URL"] = strings.ReplaceAll(lightHouseQuery, ".klovercloud.", "."+namespace+".")

	return a
}

func (a apiService) ModifyDeployment(namespace string, apiService v1alpha1.ApiService) service.ApiService {
	if a.Deployment.ObjectMeta.Labels == nil {
		a.Deployment.ObjectMeta.Labels = make(map[string]string)
	}
	a.Deployment.ObjectMeta.Labels["app"] = "klovercloudCD"
	a.Deployment.ObjectMeta.Namespace = namespace
	if apiService.Resources.Requests.Cpu() != nil || apiService.Resources.Limits.Cpu() != nil {
		for index, _ := range a.Deployment.Spec.Template.Spec.Containers {
			a.Deployment.Spec.Template.Spec.Containers[index].Resources = apiService.Resources
		}
	}
	a.Deployment.Spec.Replicas = &apiService.Size
	return a
}

func (a apiService) ModifyService(namespace string) service.ApiService {
	if a.Service.ObjectMeta.Labels == nil {
		a.Service.ObjectMeta.Labels = make(map[string]string)
	}
	a.Service.ObjectMeta.Labels["app"] = "klovercloudCD"
	a.Service.ObjectMeta.Namespace = namespace
	return a
}

func (a apiService) Apply(wait bool) error {
	if a.Error != nil {
		return a.Error
	}

	config := &basev1alpha1.KlovercloudCD{}

	ctrl.SetControllerReference(config, &a.Configmap, controllers.KlovercloudCDReconciler{}.Scheme)

	err := a.ApplyConfigMap()
	if err != nil {
		log.Println("[ERROR]: Failed to create configmap for api service.", "Deployment.Namespace", a.Deployment.Namespace, "Deployment.Name", a.Deployment.Name, err.Error())
		return err
	}

	existingPodListObject := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(a.Deployment.Namespace),
		client.MatchingLabels(a.Deployment.ObjectMeta.Labels),
	}
	if err = a.Client.List(context.Background(), existingPodListObject, listOpts...); err != nil {
		log.Println(err, "[ERROR]: Failed to list pods", "Deployment.Namespace", a.Deployment.Namespace, "Deployment.Name", a.Deployment.Name)
	}
	existingPodMap := make(map[string]bool)

	if len(existingPodListObject.Items) > 0 {
		for _, each := range existingPodListObject.Items {
			existingPodMap[each.Name] = true
		}
	}

	ctrl.SetControllerReference(config, &a.Deployment, controllers.KlovercloudCDReconciler{}.Scheme)
	err = a.ApplyDeployment()
	if err != nil {
		log.Println("[ERROR]: Failed to apply deployment for api service.", "Deployment.Namespace: ", a.Deployment.Namespace, " Deployment.Name: ", a.Deployment.Name+". ", err.Error())
		return err
	}
	if wait {
		err = utility.WaitUntilPodsAreReady(a.Client, existingPodMap, listOpts, a.Deployment.Namespace, a.Deployment.Name, *a.Deployment.Spec.Replicas, 10)
		if err != nil {
			return err
		}
	}

	ctrl.SetControllerReference(config, &a.Service, controllers.KlovercloudCDReconciler{}.Scheme)
	err = a.ApplyService()
	if err != nil {
		log.Println("[ERROR]: Failed to apply service for api service.", "Deployment.Namespace: ", a.Deployment.Namespace, " Deployment.Name: ", a.Deployment.Name+". ", err.Error())
		return err
	}

	return nil
}

func (a apiService) ApplyConfigMap() error {
	return a.Client.Create(context.Background(), &a.Configmap)
}

func (a apiService) ApplyDeployment() error {
	return a.Client.Create(context.Background(), &a.Deployment)
}

func (a apiService) ApplyService() error {
	return a.Client.Create(context.Background(), &a.Service)
}

func getConfigMapFromFile() corev1.ConfigMap {
	data, err := ioutil.ReadFile("api-service-configmap.yaml")
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
	data, err := ioutil.ReadFile("api-service-service.yaml")
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
	data, err := ioutil.ReadFile("api-service-deployment.yaml")
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

func New(client client.Client) service.ApiService {
	return apiService{
		Configmap:  getConfigMapFromFile(),
		Deployment: getDeploymentFromFile(),
		Service:    getServiceFromFile(),
		Client:     client,
	}
}
