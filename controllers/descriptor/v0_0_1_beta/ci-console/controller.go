package ci_console

import (
	"context"
	"fmt"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/service"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/utility"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type console struct {
	Configmap  corev1.ConfigMap
	Deployment appv1.Deployment
	Service    corev1.Service
	Client     client.Client
	Error      error
}

func (c console) ModifyConfigmap(namespace string) service.Console {
	if c.Configmap.ObjectMeta.Labels == nil {
		c.Configmap.ObjectMeta.Labels = make(map[string]string)
	}
	c.Configmap.ObjectMeta.Labels["app"] = "klovercloudCD"
	c.Configmap.ObjectMeta.Namespace = namespace

	v1AuthEndpoint := c.Configmap.Data["v1AuthEndpoint"]
	replacedUrl := strings.ReplaceAll(v1AuthEndpoint, ".klovercloud.", "."+namespace+".")
	c.Configmap.Data["v1AuthEndpoint"] = replacedUrl

	v1ApiEndPoint := c.Configmap.Data["v1ApiEndPoint"]
	replacedUrl = strings.ReplaceAll(v1ApiEndPoint, ".klovercloud.", "."+namespace+".")
	c.Configmap.Data["v1ApiEndPoint"] = replacedUrl

	v1ApiEndPointWS := c.Configmap.Data["v1ApiEndPointWS"]
	replacedUrl = strings.ReplaceAll(v1ApiEndPointWS, ".klovercloud.", "."+namespace+".")
	c.Configmap.Data["v1ApiEndPointWS"] = replacedUrl

	return c
}

func (c console) ModifyDeployment(namespace string, console v1alpha1.Console) service.Console {
	if c.Deployment.ObjectMeta.Labels == nil {
		c.Deployment.ObjectMeta.Labels = make(map[string]string)
	}
	c.Deployment.ObjectMeta.Labels["app"] = "klovercloudCD"
	c.Deployment.ObjectMeta.Namespace = namespace
	if console.Resources.Requests.Cpu() != nil || console.Resources.Limits.Cpu() != nil {
		for i, _ := range c.Deployment.Spec.Template.Spec.Containers {
			c.Deployment.Spec.Template.Spec.Containers[i].Resources = console.Resources
		}
	}
	c.Deployment.Spec.Replicas = &console.Size
	return c
}

func (c console) ModifyService(namespace string) service.Console {
	if c.Service.ObjectMeta.Labels == nil {
		c.Service.ObjectMeta.Labels = make(map[string]string)
	}
	c.Service.ObjectMeta.Labels["app"] = "klovercloudCD"
	c.Service.ObjectMeta.Namespace = namespace
	return c
}

func (c console) Apply(config *v1alpha1.KlovercloudCD, scheme *runtime.Scheme, wait bool) error {
	if c.Error != nil {
		return c.Error
	}
	ctrl.SetControllerReference(config, &c.Configmap, scheme)

	err := c.ApplyConfigMap()
	if err != nil {
		log.Println("[ERROR]: Failed to create configmap for console service.", "Deployment.Namespace", c.Deployment.Namespace, "Deployment.Name", c.Deployment.Name, err.Error())
		return err
	}
	existingPodListObject := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(c.Deployment.Namespace),
		client.MatchingLabels(c.Deployment.ObjectMeta.Labels),
	}
	if err = c.Client.List(context.Background(), existingPodListObject, listOpts...); err != nil {
		log.Println(err, "[ERROR]: Failed to list pods", "Deployment.Namespace", c.Deployment.Namespace, "Deployment.Name", c.Deployment.Name)
	}
	existingPodMap := make(map[string]bool)

	if len(existingPodListObject.Items) > 0 {
		for _, each := range existingPodListObject.Items {
			existingPodMap[each.Name] = true
		}
	}

	ctrl.SetControllerReference(config, &c.Deployment, scheme)
	err = c.ApplyDeployment()
	if err != nil {
		log.Println("[ERROR]: Failed to apply deployment for console service.", "Deployment.Namespace: ", c.Deployment.Namespace, " Deployment.Name: ", c.Deployment.Name+". ", err.Error())
		return err
	}
	if wait {
		err = utility.WaitUntilPodsAreReady(c.Client, existingPodMap, listOpts, c.Deployment.Namespace, c.Deployment.Name, *c.Deployment.Spec.Replicas, 10)
		if err != nil {
			return err
		}
	}

	ctrl.SetControllerReference(config, &c.Service, scheme)
	err = c.ApplyService()
	if err != nil {
		log.Println("[ERROR]: Failed to apply service for console service.", "Deployment.Namespace: ", c.Deployment.Namespace, " Deployment.Name: ", c.Deployment.Name+". ", err.Error())
		return err
	}

	return nil
}

func (c console) ApplyConfigMap() error {
	return c.Client.Create(context.Background(), &c.Configmap)
}

func (c console) ApplyDeployment() error {
	return c.Client.Create(context.Background(), &c.Deployment)
}

func (c console) ApplyService() error {
	return c.Client.Create(context.Background(), &c.Service)
}

func getConfigMapFromFile() corev1.ConfigMap {
	data, err := ioutil.ReadFile("descriptor/v0_0_1_beta/ci-console/ci-console-configMap.yml")
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
	data, err := ioutil.ReadFile("descriptor/v0_0_1_beta/ci-console/ci-console-service.yml")
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
	data, err := ioutil.ReadFile("descriptor/v0_0_1_beta/ci-console/ci-console-deployment.yml")
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
func New(client client.Client) service.Console {
	return console{
		Configmap:  getConfigMapFromFile(),
		Deployment: getDeploymentFromFile(),
		Service:    getServiceFromFile(),
		Client:     client,
	}
}
