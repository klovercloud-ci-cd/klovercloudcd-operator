package lighthouse_command

import (
	"context"
	"fmt"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/service"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/utility"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/enums"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	basev1alpha1 "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
)

type lighthouseCommand struct {
	Configmap  corev1.ConfigMap
	Deployment appv1.Deployment
	Service    corev1.Service
	Client     client.Client
	Error      error
}

func (l lighthouseCommand) Delete() error {
	l.Client.Delete(context.Background(), &l.Service)
	l.Client.Delete(context.Background(), &l.Deployment)
	l.Client.Delete(context.Background(), &l.Configmap)
	return nil
}

func (l lighthouseCommand) ModifyConfigmap(namespace string, db v1alpha1.DB) service.LightHouseCommand {
	if l.Configmap.ObjectMeta.Labels == nil {
		l.Configmap.ObjectMeta.Labels = make(map[string]string)
	}
	l.Configmap.ObjectMeta.Labels["app"] = "klovercloudCD"
	l.Configmap.ObjectMeta.Namespace = namespace
	if db.Type == enums.MONGO || db.Type == "" {
		l.Configmap.Data["MONGO"] = string(enums.MONGO)
		l.Configmap.Data["MONGO_SERVER"] = db.ServerURL
		l.Configmap.Data["MONGO_PORT"] = db.ServerPort
	}
	return l
}

func (l lighthouseCommand) ModifyDeployment(namespace string, lightHouseCommand v1alpha1.LightHouseCommand) service.LightHouseCommand {
	if l.Deployment.ObjectMeta.Labels == nil {
		l.Deployment.ObjectMeta.Labels = make(map[string]string)
	}
	l.Deployment.ObjectMeta.Labels["app"] = "klovercloudCD"
	l.Deployment.ObjectMeta.Namespace = namespace
	if lightHouseCommand.Resources.Requests.Cpu() != nil || lightHouseCommand.Resources.Limits.Cpu() != nil {
		for i, _ := range l.Deployment.Spec.Template.Spec.Containers {
			l.Deployment.Spec.Template.Spec.Containers[i].Resources = lightHouseCommand.Resources
		}
	}
	l.Deployment.Spec.Replicas = &lightHouseCommand.Size
	return l
}

func (l lighthouseCommand) ModifyService(namespace string) service.LightHouseCommand {
	if l.Service.ObjectMeta.Labels == nil {
		l.Service.ObjectMeta.Labels = make(map[string]string)
	}
	l.Service.ObjectMeta.Labels["app"] = "klovercloudCD"
	l.Service.ObjectMeta.Namespace = namespace
	return l
}

func (l lighthouseCommand) Apply(scheme *runtime.Scheme,wait bool) error {
	if l.Error != nil {
		return l.Error
	}

	config := &basev1alpha1.KlovercloudCD{}

	ctrl.SetControllerReference(config, &l.Configmap, scheme)
	err := l.ApplyConfigMap()
	if err != nil {
		log.Println("[ERROR]: Failed to create configmap for lighthouse command service.", "Deployment.Namespace", l.Deployment.Namespace, "Deployment.Name", l.Deployment.Name, err.Error())
		return err
	}
	existingPodListObject := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(l.Deployment.Namespace),
		client.MatchingLabels(l.Deployment.ObjectMeta.Labels),
	}
	if err = l.Client.List(context.Background(), existingPodListObject, listOpts...); err != nil {
		log.Println(err, "[ERROR]: Failed to list pods", "Deployment.Namespace", l.Deployment.Namespace, "Deployment.Name", l.Deployment.Name)
	}
	existingPodMap := make(map[string]bool)

	if len(existingPodListObject.Items) > 0 {
		for _, each := range existingPodListObject.Items {
			existingPodMap[each.Name] = true
		}
	}

	ctrl.SetControllerReference(config, &l.Deployment, scheme)
	err = l.ApplyDeployment()
	if err != nil {
		log.Println("[ERROR]: Failed to apply deployment for lighthouse command service.", "Deployment.Namespace: ", l.Deployment.Namespace, " Deployment.Name: ", l.Deployment.Name+". ", err.Error())
		return err
	}
	if wait {
		err = utility.WaitUntilPodsAreReady(l.Client, existingPodMap, listOpts, l.Deployment.Namespace, l.Deployment.Name, *l.Deployment.Spec.Replicas, 10)
		if err != nil {
			return err
		}
	}

	ctrl.SetControllerReference(config, &l.Service, scheme)
	err = l.ApplyService()
	if err != nil {
		log.Println("[ERROR]: Failed to apply service for lighthouse command service.", "Deployment.Namespace: ", l.Deployment.Namespace, " Deployment.Name: ", l.Deployment.Name+". ", err.Error())
		return err
	}

	return nil
}

func (l lighthouseCommand) ApplyConfigMap() error {
	return l.Client.Create(context.Background(), &l.Configmap)
}

func (l lighthouseCommand) ApplyDeployment() error {
	return l.Client.Create(context.Background(), &l.Deployment)
}

func (l lighthouseCommand) ApplyService() error {
	return l.Client.Create(context.Background(), &l.Service)
}

func getConfigMapFromFile() corev1.ConfigMap {
	data, err := ioutil.ReadFile("light-house-command-configmap.yml")
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
	data, err := ioutil.ReadFile("light-house-command-service.yml")
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
	data, err := ioutil.ReadFile("light-house-command-deployment.yml")
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

func New(client client.Client) service.LightHouseCommand {
	return lighthouseCommand{
		Configmap:  getConfigMapFromFile(),
		Deployment: getDeploymentFromFile(),
		Service:    getServiceFromFile(),
		Client:     client,
	}
}
