package event_bank

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

type eventBank struct {
	Configmap  corev1.ConfigMap
	Deployment appv1.Deployment
	Service    corev1.Service
	Client     client.Client
	Error      error
}

func (e eventBank) ModifyConfigmap(namespace string, db v1alpha1.DB) service.EventBank {
	if e.Configmap.ObjectMeta.Labels == nil {
		e.Configmap.ObjectMeta.Labels = make(map[string]string)
	}
	e.Configmap.ObjectMeta.Labels["app"] = "klovercloudCD"
	e.Configmap.ObjectMeta.Namespace = namespace
	if db.Type == enums.MONGO || db.Type == "" {
		e.Configmap.Data["MONGO"] = string(enums.MONGO)
		e.Configmap.Data["MONGO_SERVER"] = db.ServerURL
		e.Configmap.Data["MONGO_PORT"] = db.ServerPort
	}

	KLOVERCLOUD_CI_CORE_URL := e.Configmap.Data["KLOVERCLOUD_CI_CORE_URL"]
	replacedUrl := strings.ReplaceAll(KLOVERCLOUD_CI_CORE_URL, ".klovercloud.", "."+namespace+".")
	e.Configmap.Data["KLOVERCLOUD_CI_CORE_URL"] = replacedUrl

	return e
}

func (e eventBank) ModifyDeployment(namespace string, eventBank v1alpha1.EventBank) service.EventBank {
	if e.Deployment.ObjectMeta.Labels == nil {
		e.Deployment.ObjectMeta.Labels = make(map[string]string)
	}
	e.Deployment.ObjectMeta.Labels["app"] = "klovercloudCD"
	e.Deployment.ObjectMeta.Namespace = namespace
	if eventBank.Resources.Requests.Cpu() != nil || eventBank.Resources.Limits.Cpu() != nil {
		for index, _ := range e.Deployment.Spec.Template.Spec.Containers {
			e.Deployment.Spec.Template.Spec.Containers[index].Resources = eventBank.Resources
		}
	}
	return e
}

func (e eventBank) ModifyService(namespace string) service.EventBank {
	if e.Service.ObjectMeta.Labels == nil {
		e.Service.ObjectMeta.Labels = make(map[string]string)
	}
	e.Service.ObjectMeta.Labels["app"] = "klovercloudCD"
	e.Service.ObjectMeta.Namespace = namespace
	return e
}

func (e eventBank) Apply(wait bool) error {
	if e.Error != nil {
		return e.Error
	}
	err := e.ApplyConfigMap()
	if err != nil {
		log.Println("[ERROR]: Failed to create configmap for event bank service.", "Deployment.Namespace", e.Deployment.Namespace, "Deployment.Name", e.Deployment.Name, err.Error())
		return err
	}
	existingPodListObject := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(e.Deployment.Namespace),
		client.MatchingLabels(e.Deployment.ObjectMeta.Labels),
	}
	if err = e.Client.List(context.Background(), existingPodListObject, listOpts...); err != nil {
		log.Println(err, "[ERROR]: Failed to list pods", "Deployment.Namespace", e.Deployment.Namespace, "Deployment.Name", e.Deployment.Name)
	}
	existingPodMap := make(map[string]bool)

	if len(existingPodListObject.Items) > 0 {
		for _, each := range existingPodListObject.Items {
			existingPodMap[each.Name] = true
		}
	}

	err = e.ApplyDeployment()
	if err != nil {
		log.Println("[ERROR]: Failed to apply deployment for event bank service.", "Deployment.Namespace: ", e.Deployment.Namespace, " Deployment.Name: ", e.Deployment.Name+". ", err.Error())
		return err
	}
	if wait {
		err = utility.WaitUntilPodsAreReady(e.Client, existingPodMap, listOpts, e.Deployment.Namespace, e.Deployment.Name, *e.Deployment.Spec.Replicas, 10)
		if err != nil {
			return err
		}
	}
	err = e.ApplyService()
	if err != nil {
		log.Println("[ERROR]: Failed to apply service for event bank service.", "Deployment.Namespace: ", e.Deployment.Namespace, " Deployment.Name: ", e.Deployment.Name+". ", err.Error())
		return err
	}

	return nil
}

func (e eventBank) ApplyConfigMap() error {
	return e.Client.Create(context.Background(), &e.Configmap)
}

func (e eventBank) ApplyDeployment() error {
	return e.Client.Create(context.Background(), &e.Deployment)
}

func (e eventBank) ApplyService() error {
	return e.Client.Create(context.Background(), &e.Service)
}

func getConfigMapFromFile() corev1.ConfigMap {
	data, err := ioutil.ReadFile("event-bank-configmap.yaml")
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
	data, err := ioutil.ReadFile("event-bank-service.yaml")
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
	data, err := ioutil.ReadFile("event-bank-deployment.yaml")
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

func New(client client.Client) service.EventBank {
	return eventBank{
		Configmap:  getConfigMapFromFile(),
		Deployment: getDeploymentFromFile(),
		Service:    getServiceFromFile(),
		Client:     client,
	}
}
