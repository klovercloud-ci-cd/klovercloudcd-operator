package core_engine

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
	rbacv1 "k8s.io/api/rbac/v1"
	_ "k8s.io/client-go/informers/rbac"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type coreEngine struct {
	Deployment         appv1.Deployment
	Service            corev1.Service
	ConfigMap          corev1.ConfigMap
	ClusterRole        rbacv1.ClusterRole
	ClusterRoleBinding rbacv1.ClusterRoleBinding
	ServiceAccount     corev1.ServiceAccount
	Client             client.Client
	Error              error
}

func (c coreEngine) ModifyConfigmap(namespace string, db v1alpha1.DB) service.CoreEngine {
	if c.ConfigMap.ObjectMeta.Labels == nil {
		c.ConfigMap.ObjectMeta.Labels = make(map[string]string)
	}
	c.ConfigMap.ObjectMeta.Labels["app"] = "klovercloudCD"
	c.ConfigMap.ObjectMeta.Namespace = namespace
	if db.Type == enums.MONGO || db.Type == "" {
		c.ConfigMap.Data["MONGO"] = string(enums.MONGO)
		c.ConfigMap.Data["MONGO_SERVER"] = db.ServerURL
		c.ConfigMap.Data["MONGO_PORT"] = db.ServerPort
	}

	EVENT_STORE_URL := c.ConfigMap.Data["EVENT_STORE_URL"]
	replacedUrl := strings.ReplaceAll(EVENT_STORE_URL, ".klovercloud.", "."+namespace+".")
	c.ConfigMap.Data["EVENT_STORE_URL"] = replacedUrl

	return c
}

func (c coreEngine) ModifyDeployment(namespace string, coreEngine v1alpha1.CoreEngine) service.CoreEngine {
	if c.Deployment.ObjectMeta.Labels == nil {
		c.Deployment.ObjectMeta.Labels = make(map[string]string)
	}
	c.Deployment.ObjectMeta.Labels["app"] = "klovercloudCD"
	c.Deployment.ObjectMeta.Namespace = namespace
	if coreEngine.Resources.Requests.Cpu() != nil || coreEngine.Resources.Limits.Cpu() != nil {
		for index, _ := range c.Deployment.Spec.Template.Spec.Containers {
			c.Deployment.Spec.Template.Spec.Containers[index].Resources = coreEngine.Resources
		}
	}
	c.Deployment.Spec.Replicas=&coreEngine.Size
	return c
}

func (c coreEngine) ModifyService(namespace string) service.CoreEngine {
	if c.Service.ObjectMeta.Labels == nil {
		c.Service.ObjectMeta.Labels = make(map[string]string)
	}
	c.Service.ObjectMeta.Labels["app"] = "klovercloudCD"
	c.Service.ObjectMeta.Namespace = namespace
	return c
}

func (c coreEngine) ModifyClusterRole(namespace string) service.CoreEngine {
	if c.ClusterRole.ObjectMeta.Labels == nil {
		c.ClusterRole.ObjectMeta.Labels = make(map[string]string)
	}
	c.ClusterRole.ObjectMeta.Labels["app"] = "klovercloudCD"
	c.ClusterRole.ObjectMeta.Namespace = namespace

	return c
}

func (c coreEngine) ModifyClusterRoleBinding(namespace string) service.CoreEngine {
	if c.ClusterRoleBinding.ObjectMeta.Labels == nil {
		c.ClusterRoleBinding.ObjectMeta.Labels = make(map[string]string)
	}
	c.ClusterRoleBinding.ObjectMeta.Labels["app"] = "klovercloudCD"
	c.ClusterRoleBinding.ObjectMeta.Namespace = namespace

	return c
}

func (c coreEngine) ModifyServiceAccount(namespace string) service.CoreEngine {
	if c.ServiceAccount.ObjectMeta.Labels == nil {
		c.ServiceAccount.ObjectMeta.Labels = make(map[string]string)
	}
	c.ServiceAccount.ObjectMeta.Labels["app"] = "klovercloudCD"
	c.ServiceAccount.ObjectMeta.Namespace = namespace

	return c
}

func (c coreEngine) Apply(wait bool) error {
	if c.Error != nil {
		return c.Error
	}
	err := c.ApplyClusterRole()
	if err != nil {
		log.Println("[ERROR]: Failed to create cluster role for core engine service.", "Deployment.Namespace", c.Deployment.Namespace, "Deployment.Name", c.Deployment.Name, err.Error())
		return err
	}
	err = c.ApplyClusterRoleBinding()
	if err != nil {
		log.Println("[ERROR]: Failed to create cluster role binding for core engine service.", "Deployment.Namespace", c.Deployment.Namespace, "Deployment.Name", c.Deployment.Name, err.Error())
		return err
	}
	err = c.ApplyServiceAccount()
	if err != nil {
		log.Println("[ERROR]: Failed to create service account for core engine service.", "Deployment.Namespace", c.Deployment.Namespace, "Deployment.Name", c.Deployment.Name, err.Error())
		return err
	}
	err = c.ApplyConfigMap()
	if err != nil {
		log.Println("[ERROR]: Failed to create configmap for event bank service.", "Deployment.Namespace", c.Deployment.Namespace, "Deployment.Name", c.Deployment.Name, err.Error())
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

	err = c.ApplyDeployment()
	if err != nil {
		log.Println("[ERROR]: Failed to apply deployment for event bank service.", "Deployment.Namespace: ", c.Deployment.Namespace, " Deployment.Name: ", c.Deployment.Name+". ", err.Error())
		return err
	}
	if wait {
		err = utility.WaitUntilPodsAreReady(c.Client, existingPodMap, listOpts, c.Deployment.Namespace, c.Deployment.Name, *c.Deployment.Spec.Replicas, 10)
		if err != nil {
			return err
		}
	}
	err = c.ApplyService()
	if err != nil {
		log.Println("[ERROR]: Failed to apply service for event bank service.", "Deployment.Namespace: ", c.Deployment.Namespace, " Deployment.Name: ", c.Deployment.Name+". ", err.Error())
		return err
	}

	return nil
}

func (c coreEngine) ApplyConfigMap() error {
	return c.Client.Create(context.Background(), &c.ConfigMap)
}

func (c coreEngine) ApplyDeployment() error {
	return c.Client.Create(context.Background(), &c.Deployment)
}

func (c coreEngine) ApplyService() error {
	return c.Client.Create(context.Background(), &c.Service)
}

func (c coreEngine) ApplyClusterRole() error {
	return c.Client.Create(context.Background(), &c.ClusterRole)
}

func (c coreEngine) ApplyClusterRoleBinding() error {
	return c.Client.Create(context.Background(), &c.ClusterRoleBinding)
}

func (c coreEngine) ApplyServiceAccount() error {
	return c.Client.Create(context.Background(), &c.ServiceAccount)
}

func getDeploymentFromFile() appv1.Deployment {
	data, err := ioutil.ReadFile("security-server-deployment.yaml")
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
func getConfigMapFromFile() corev1.ConfigMap {
	data, err := ioutil.ReadFile("security-server-deployment.yaml")
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
	data, err := ioutil.ReadFile("security-server-deployment.yaml")
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
func getClusterRoleFromFile() rbacv1.ClusterRole {
	data, err := ioutil.ReadFile("security-server-deployment.yaml")
	if err != nil {
		panic(err.Error())
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*rbacv1.ClusterRole)
}
func getClusterRoleBindingFromFile() rbacv1.ClusterRoleBinding {
	data, err := ioutil.ReadFile("security-server-deployment.yaml")
	if err != nil {
		panic(err.Error())
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*rbacv1.ClusterRoleBinding)
}
func getServiceAccountFromFile() corev1.ServiceAccount {
	data, err := ioutil.ReadFile("security-server-deployment.yaml")
	if err != nil {
		panic(err.Error())
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*corev1.ServiceAccount)
}
func New(client client.Client) service.CoreEngine {
	return coreEngine{
		Deployment:         getDeploymentFromFile(),
		Service:            getServiceFromFile(),
		ConfigMap:          getConfigMapFromFile(),
		ClusterRole:        getClusterRoleFromFile(),
		ClusterRoleBinding: getClusterRoleBindingFromFile(),
		ServiceAccount:     getServiceAccountFromFile(),
		Client:             client,
	}
}
