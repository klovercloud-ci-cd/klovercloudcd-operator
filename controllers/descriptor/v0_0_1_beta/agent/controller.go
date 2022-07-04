package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/service"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/utility"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type agent struct {
	ClusterRole        rbacv1.ClusterRole
	ClusterRoleBinding rbacv1.ClusterRoleBinding
	ServiceAccount     corev1.ServiceAccount
	Configmap          corev1.ConfigMap
	Deployment         appv1.Deployment
	Service            corev1.Service
	Client             client.Client
	RestConfig         *rest.Config
	Error              error
}

func (a agent) ModifyClusterRole() service.Agent {
	if a.ClusterRole.ObjectMeta.Labels == nil {
		a.ClusterRole.ObjectMeta.Labels = make(map[string]string)
	}
	a.ClusterRole.ObjectMeta.Labels["app"] = "klovercloudCD"
	return a
}

func (a agent) ModifyClusterRoleBinding(namespace string, agent v1alpha1.Agent) service.Agent {
	if a.ClusterRoleBinding.ObjectMeta.Labels == nil {
		a.ClusterRoleBinding.ObjectMeta.Labels = make(map[string]string)
	}
	a.ClusterRoleBinding.ObjectMeta.Labels["app"] = "klovercloudCD"
	a.ClusterRoleBinding.ObjectMeta.Namespace = namespace
	return a
}

func (a agent) ModifyServiceAccount(namespace string, agent v1alpha1.Agent) service.Agent {
	if a.ServiceAccount.ObjectMeta.Labels == nil {
		a.ServiceAccount.ObjectMeta.Labels = make(map[string]string)
	}
	a.ServiceAccount.ObjectMeta.Labels["app"] = "klovercloudCD"
	a.ServiceAccount.ObjectMeta.Namespace = namespace
	return a
}

func (a agent) ModifyConfigmap(namespace string, agent v1alpha1.Agent) service.Agent {
	if a.Configmap.ObjectMeta.Labels == nil {
		a.Configmap.ObjectMeta.Labels = make(map[string]string)
	}
	a.Configmap.ObjectMeta.Labels["app"] = "klovercloudCD"
	a.Configmap.ObjectMeta.Namespace = namespace

	apiServicePodListObject := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"app": "klovercloud-api-service"}),
	}
	if err := a.Client.List(context.Background(), apiServicePodListObject, listOpts...); err != nil {
		log.Println(err, "[ERROR]: Failed to generate token as failed to list pods.", "Deployment.Namespace: ", namespace, ", Deployment.Name: ", "klovercloud-api-service")
		a.Error = errors.New("[ERROR]: Failed to generate token as failed to list pods." + "Deployment.Namespace: " + namespace + ", Deployment.Name: " + "klovercloud-api-service")
		return a
	}

	if len(apiServicePodListObject.Items) > 0 {
		data, _, err := utility.ExecuteRemoteCommand(a.RestConfig, apiServicePodListObject.Items[0].Name, apiServicePodListObject.Items[0].Namespace, "app", []string{"kcpctl", "generate-jwt", "client=local_agent"})

		if err != nil {
			log.Println(err, "[ERROR]: Failed to generate token.", "Pod.Namespace: ", namespace, ", Pod.Name: ", apiServicePodListObject.Items[0].Name)
			a.Error = errors.New("[ERROR]: Failed to generate token." + "Pod.Namespace: " + namespace + ", Pod.Name: " + apiServicePodListObject.Items[0].Name)
			return a
		}
		if data != "" {
			outputs := strings.Split(data, " ")
			if len(outputs) == 2 {
				a.Configmap.Data["TOKEN"] = outputs[1]
			} else {
				a.Error = errors.New("[ERROR]: Failed to generate token. Expected format: token:  <token> . Found: " + data)
				return a
			}
		}
	} else {
		a.Error = errors.New("[ERROR]: Found no api service! Please verify, if api service is working properly!")
		return a
	}
	if agent.PullSize != "" {
		a.Configmap.Data["PULL_SIZE"] = agent.PullSize
	}
	if agent.LightHouseEnabled == "true" {
		a.Configmap.Data["LIGHTHOUSE_ENABLED"] = "true"
	}
	if agent.TerminalBaseUrl != "" {
		a.Configmap.Data["TERMINAL_BASE_URL"] = agent.TerminalBaseUrl
	}
	if agent.TerminalApiVersion != "" {
		a.Configmap.Data["TERMINAL_API_VERSION"] = agent.TerminalApiVersion
	}

	eventStoreUrl := a.Configmap.Data["EVENT_STORE_URL"]
	replacedUrl := strings.ReplaceAll(eventStoreUrl, ".klovercloud.", "."+namespace+".")
	a.Configmap.Data["EVENT_STORE_URL"] = replacedUrl
	apiServiceUrl := a.Configmap.Data["API_SERVICE_URL"]
	replacedUrl = strings.ReplaceAll(apiServiceUrl, ".klovercloud.", "."+namespace+".")
	a.Configmap.Data["API_SERVICE_URL"] = replacedUrl

	return a
}

func (a agent) ModifyDeployment(namespace string, agent v1alpha1.Agent) service.Agent {
	if a.Deployment.ObjectMeta.Labels == nil {
		a.Deployment.ObjectMeta.Labels = make(map[string]string)
	}
	a.Deployment.ObjectMeta.Labels["app"] = "klovercloudCD"
	a.Deployment.ObjectMeta.Namespace = namespace
	if agent.Resources.Requests.Cpu() != nil || agent.Resources.Limits.Cpu() != nil {
		for index := range a.Deployment.Spec.Template.Spec.Containers {
			a.Deployment.Spec.Template.Spec.Containers[index].Resources = agent.Resources
		}
	}
	return a
}

func (a agent) ModifyService(namespace string) service.Agent {
	if a.Service.ObjectMeta.Labels == nil {
		a.Service.ObjectMeta.Labels = make(map[string]string)
	}
	a.Service.ObjectMeta.Labels["app"] = "klovercloudCD"
	a.Service.ObjectMeta.Namespace = namespace
	return a
}

func (a agent) Apply(wait bool) error {
	if a.Error != nil {
		return a.Error
	}
	err := a.ApplyClusterRole()
	if err != nil {
		return err
	}

	err = a.ApplyServiceAccount()
	if err != nil {
		return err
	}

	err = a.ApplyClusterRoleBinding()
	if err != nil {
		return err
	}

	err = a.ApplyConfigMap()
	if err != nil {
		return err
	}

	err = a.ApplyDeployment()
	if err != nil {
		return err
	}

	err = a.ApplyService()
	if err != nil {
		return err
	}
	return nil
}

func (a agent) ApplyClusterRole() error {
	return a.Client.Create(context.Background(), &a.ClusterRole)
}

func (a agent) ApplyClusterRoleBinding() error {
	return a.Client.Create(context.Background(), &a.ClusterRoleBinding)
}

func (a agent) ApplyServiceAccount() error {
	return a.Client.Create(context.Background(), &a.ServiceAccount)
}

func (a agent) ApplyConfigMap() error {
	return a.Client.Create(context.Background(), &a.Configmap)
}

func (a agent) ApplyDeployment() error {
	return a.Client.Create(context.Background(), &a.Deployment)
}

func (a agent) ApplyService() error {
	return a.Client.Create(context.Background(), &a.Service)
}

func getConfigMapFromFile() corev1.ConfigMap {
	data, err := ioutil.ReadFile("agent-configmap.yaml")
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

func getClusterRoleFromFile() rbacv1.ClusterRole {
	data, err := ioutil.ReadFile("agent-cluster-role.yaml")
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
	data, err := ioutil.ReadFile("agent-cluster-rolebinding.yaml")
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
	data, err := ioutil.ReadFile("agent-service-account.yaml")
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

func getServiceFromFile() corev1.Service {
	data, err := ioutil.ReadFile("agent-service.yaml")
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
	data, err := ioutil.ReadFile("agent-deployment.yaml")
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

func New(client client.Client, restConfig *rest.Config) service.Agent {
	return agent{
		ClusterRole:        getClusterRoleFromFile(),
		ClusterRoleBinding: getClusterRoleBindingFromFile(),
		ServiceAccount:     getServiceAccountFromFile(),
		Configmap:          getConfigMapFromFile(),
		Deployment:         getDeploymentFromFile(),
		Service:            getServiceFromFile(),
		Client:             client,
		RestConfig:         restConfig,
		Error:              nil,
	}
}
