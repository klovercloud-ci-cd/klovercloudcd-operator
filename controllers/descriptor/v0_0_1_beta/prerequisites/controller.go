package prerequisites

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/service"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/utility"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/enums"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	k8Sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"log"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type prerequisites struct {
	Secret           coreV1.Secret
	Client           client.Client
	RestConfig       *rest.Config
	Configmap        coreV1.ConfigMap
	TektonDescriptor []unstructured.Unstructured
	Error            error
}

func (p prerequisites) ModifySecurityConfigMap(namespace string, db v1alpha1.DB, security v1alpha1.Security) service.Prerequisites {
	found := &coreV1.ConfigMap{}
	_ = p.Client.Get(context.Background(), types.NamespacedName{Name: "klovercloud-security-envar-config", Namespace: namespace}, found)
	if found.Name != "" {
		p.Configmap.Data["PRIVATE_KEY"] = found.Data["PRIVATE_KEY"]
		p.Configmap.Data["PUBLIC_KEY"] = found.Data["PUBLIC_KEY"]
	} else {
		private, public, err := utility.New().Generate()
		if err != nil {
			p.Error = err
			log.Println("[ERROR]: Failed to modify secrets configmap." + err.Error())
		}
		p.Configmap.Data["PRIVATE_KEY"] = string(private)
		p.Configmap.Data["PUBLIC_KEY"] = string(public)
	}
	if p.Configmap.ObjectMeta.Labels == nil {
		p.Configmap.ObjectMeta.Labels = make(map[string]string)
	}
	p.Configmap.ObjectMeta.Labels["app"] = "klovercloudCD"
	p.Configmap.ObjectMeta.Namespace = namespace
	p.Configmap.Data["MAIL_SERVER_HOST_EMAIL"] = security.MailServerHostEmail
	p.Configmap.Data["MAIL_SERVER_HOST_EMAIL_SECRET"] = security.MailServerHostEmailSecret
	p.Configmap.Data["SMTP_HOST"] = security.SMTPHost
	p.Configmap.Data["SMTP_PORT"] = security.SMTPPort
	p.Configmap.Data["USER_FIRST_NAME"] = security.User.FirstName
	p.Configmap.Data["USER_LAST_NAME"] = security.User.LastName
	p.Configmap.Data["USER_EMAIL"] = security.User.Email
	p.Configmap.Data["USER_PHONE"] = security.User.Phone
	p.Configmap.Data["USER_PASSWORD"] = security.User.Password
	p.Configmap.Data["COMPANY_NAME"] = security.User.CompanyName
	if db.Type == enums.MONGO || db.Type == "" {
		p.Configmap.Data["MONGO"] = string(enums.MONGO)
		p.Configmap.Data["MONGO_SERVER"] = db.ServerURL
		p.Configmap.Data["MONGO_PORT"] = db.ServerPort
	}

	API_SERVER_URL := p.Configmap.Data["API_SERVER_URL"]
	p.Configmap.Data["API_SERVER_URL"] = strings.ReplaceAll(API_SERVER_URL, ".klovercloud.", "."+namespace+".")

	INTEGRATION_MANAGER_URL := p.Configmap.Data["INTEGRATION_MANAGER_URL"]
	p.Configmap.Data["INTEGRATION_MANAGER_URL"] = strings.ReplaceAll(INTEGRATION_MANAGER_URL, ".klovercloud.", "."+namespace+".")

	return p
}

func (p prerequisites) ApplySecurityConfigMap() error {
	return p.Client.Create(context.Background(), &p.Configmap)
}

func (p prerequisites) ModifyTektonDescriptor(namespace string) service.Prerequisites {
	for i := range p.TektonDescriptor {
		p.TektonDescriptor[i].SetNamespace(namespace)
	}
	return p
}

func (p prerequisites) ModifySecret(namespace string, db v1alpha1.DB) service.Prerequisites {
	if p.Secret.ObjectMeta.Labels == nil {
		p.Secret.ObjectMeta.Labels = make(map[string]string)
	}
	p.Secret.ObjectMeta.Labels["app"] = "klovercloudCD"
	p.Secret.ObjectMeta.Namespace = namespace
	p.Secret.StringData["MONGO_USERNAME"] = db.UserName
	p.Secret.StringData["MONGO_PASSWORD"] = db.Password

	return p
}

func (p prerequisites) ApplySecret() error {
	return p.Client.Create(context.Background(), &p.Secret)
}

func (p prerequisites) ApplyTektonDescriptor(config *v1alpha1.KlovercloudCD, scheme *runtime.Scheme) error {
	existingTektonController := &appsv1.Deployment{}
	err := p.Client.Get(context.Background(), types.NamespacedName{Name: "tekton-pipelines-controller", Namespace: p.Secret.Namespace}, existingTektonController)
	if err != nil && errors.IsNotFound(err) {
		for _, each := range p.TektonDescriptor {
			ctrl.SetControllerReference(config, &each, scheme)
			each.SetNamespace(p.Secret.Namespace)
			if each.GetKind() == "Namespace" || each.GetKind() == "Namespaces" {
				each.SetName(p.Secret.Namespace)
				ctrl.SetControllerReference(config, &each, scheme)
				p.Client.Update(context.Background(), &each)
			} else if (each.GetKind() == "ClusterRole" || each.GetKind() == "ClusterRoles") && each.GetName() == "tekton-pipelines-webhook-cluster-access" {
				var clusterRole rbacv1.ClusterRole
				err = runtime.DefaultUnstructuredConverter.FromUnstructured(each.UnstructuredContent(), &clusterRole)
				if err != nil {
					log.Println("[ERROR:]: Failed to serialize. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
					return err
				}
				clusterRole.Rules[6].ResourceNames = []string{config.Namespace}
				clusterRole.Rules[7].ResourceNames = []string{config.Namespace}
				err = p.Client.Create(context.Background(), &clusterRole)
				if err != nil {
					log.Println("[WARINING:]: While applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
					if strings.Contains(err.Error(), "already exists") {
						log.Println(err.Error(), ". Updating ...")
						p.Client.Update(context.Background(), &each)
					} else {
						log.Println("[ERROR:]: Failed applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
						return err
					}
				}

			} else if each.GetKind() == "ClusterRoleBinding" || each.GetKind() == "ClusterRoleBindings" {
				var clusterRoleBinding rbacv1.ClusterRoleBinding
				err = runtime.DefaultUnstructuredConverter.FromUnstructured(each.UnstructuredContent(), &clusterRoleBinding)
				if err != nil {
					log.Println("[ERROR:]: Failed to serialize. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
					return err
				}
				clusterRoleBinding.Subjects[0].Namespace = config.Namespace
				err = p.Client.Create(context.Background(), &clusterRoleBinding)
				if err != nil {
					log.Println("[WARINING:]: While applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
					if strings.Contains(err.Error(), "already exists") {
						log.Println(err.Error(), ". Updating ...")
						p.Client.Update(context.Background(), &each)
					} else {
						log.Println("[ERROR:]: Failed applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
						return err
					}
				}
			} else if (each.GetKind() == "RoleBinding" || each.GetKind() == "RoleBindings") && each.GetName() != "tekton-pipelines-info" {
				var roleBinding rbacv1.RoleBinding
				err = runtime.DefaultUnstructuredConverter.FromUnstructured(each.UnstructuredContent(), &roleBinding)
				if err != nil {
					log.Println("[ERROR:]: Failed to serialize. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
					return err
				}
				roleBinding.Subjects[0].Namespace = config.Namespace
				err = p.Client.Create(context.Background(), &roleBinding)
				if err != nil {
					log.Println("[WARINING:]: While applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
					if strings.Contains(err.Error(), "already exists") {
						log.Println(err.Error(), ". Updating ...")
						p.Client.Update(context.Background(), &each)
					} else {
						log.Println("[ERROR:]: Failed applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
						return err
					}
				}
			} else if each.GetKind() == "CustomResourceDefinition" || each.GetKind() == "CustomResourceDefinitions" {

				definition := apiextensionsv1.CustomResourceDefinition{}
				err = runtime.DefaultUnstructuredConverter.FromUnstructured(each.UnstructuredContent(), &definition)
				if err != nil {
					log.Println("[ERROR:]: Failed to serialize. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
					return err
				}
				if definition.Spec.Conversion != nil {
					definition.Spec.Conversion.Webhook.ClientConfig.Service.Namespace = config.Namespace
				}

				data, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(definition)
				each.SetUnstructuredContent(data)
				each.SetKind("CustomResourceDefinition")
				each.SetLabels(definition.Labels)
				apiextensionsv1.AddToScheme(scheme)
				err = p.Client.Create(context.Background(), definition.DeepCopy())
				if err != nil {
					log.Println("[WARINING:]: While applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
					if strings.Contains(err.Error(), "already exists") {
						log.Println(err.Error(), ". Updating ...")
						p.Client.Update(context.Background(), &each)
					} else {
						log.Println("[ERROR:]: Failed applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
						return err
					}
				}
			} else if each.GetKind() == "ValidatingWebhookConfiguration" || each.GetKind() == "ValidatingWebhookConfigurations" {
				webhook := make(map[string]interface{})
				clientConfig := make(map[string]interface{})
				service := make(map[string]interface{})
				objectSelector := make(map[string]interface{})
				matchLabel := make(map[string]string)

				var webhooks []interface{}
				webhooks = append(webhooks, webhook)
				service["name"] = "tekton-pipelines-webhook"
				service["namespace"] = config.Namespace
				clientConfig["service"] = service
				webhook["admissionReviewVersions"] = []string{"v1"}
				webhook["clientConfig"] = clientConfig
				webhook["failurePolicy"] = "Fail"
				webhook["sideEffects"] = "None"
				webhook["name"] = each.GetName()
				if each.GetName() == "config.webhook.pipeline.tekton.dev" {
					matchLabel["app.kubernetes.io/part-of"] = "tekton-pipelines"
					objectSelector["matchLabels"] = matchLabel
					webhook["objectSelector"] = objectSelector
				}
				each.Object["webhooks"] = webhooks

				err = p.Client.Create(context.Background(), &each)
				if err != nil {
					log.Println("[WARINING:]: While applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
					if strings.Contains(err.Error(), "already exists") {
						log.Println(err.Error(), ". Updating ...")
						p.Client.Update(context.Background(), &each)
					} else {
						log.Println("[ERROR:]: Failed applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
						return err
					}
				}
			} else if each.GetKind() == "MutatingWebhookConfiguration" || each.GetKind() == "MutatingWebhookConfigurations" {
				webhook := make(map[string]interface{})
				clientConfig := make(map[string]interface{})
				service := make(map[string]interface{})
				service["name"] = "tekton-pipelines-webhook"
				service["namespace"] = config.Namespace
				clientConfig["service"] = service
				webhook["admissionReviewVersions"] = []string{"v1"}
				webhook["clientConfig"] = clientConfig
				webhook["failurePolicy"] = "Fail"
				webhook["sideEffects"] = "None"
				webhook["name"] = each.GetName()
				var webhooks []interface{}
				webhooks = append(webhooks, webhook)
				each.Object["webhooks"] = webhooks
				err = p.Client.Create(context.Background(), &each)
				if err != nil {
					log.Println("[WARINING:]: While applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
					if strings.Contains(err.Error(), "already exists") {
						log.Println(err.Error(), ". Updating ...")
						p.Client.Update(context.Background(), &each)
					} else {
						log.Println("[ERROR:]: Failed applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
						return err
					}
				}
			} else {
				err := p.Client.Create(context.Background(), &each)
				if err != nil {
					log.Println("Object: ", each)
					if strings.Contains(err.Error(), "already exists") {
						log.Println(err.Error())
					} else {
						log.Println("[ERROR]: While creating regular tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
						return err
					}
					log.Println("[WARINING:]: While applying tekton descriptors. Kind: ", each.GetKind(), ", name: ", each.GetName(), err.Error())
				}
			}
		}
	}
	return nil
}

func (p prerequisites) Apply(config *v1alpha1.KlovercloudCD, scheme *runtime.Scheme) error {
	if p.Error != nil {
		return p.Error
	}

	err := p.ApplyTektonDescriptor(config, scheme)
	if err != nil {
		log.Println("[ERROR]: Failed to create tekton", err.Error())
		return err
	}

	ctrl.SetControllerReference(config, &p.Secret, scheme)
	err = p.ApplySecret()
	if err != nil {
		log.Println("[ERROR]: Failed to create secret ", "Secret.Namespace", p.Secret.Namespace, "Secret.Name", p.Secret.Name, err.Error())
		return err
	}

	ctrl.SetControllerReference(config, &p.Configmap, scheme)
	err = p.ApplySecurityConfigMap()
	if err != nil {
		log.Println("[ERROR]: Failed to create security service configMap ", "Secret.Namespace", p.Secret.Namespace, "Secret.Name", p.Secret.Name, err.Error())
		return err
	}
	return nil
}

func getSecretFromFile() coreV1.Secret {
	absPath, _ := filepath.Abs("descriptor/v0_0_1_beta/prerequisites/mongo-secret.yaml")
	data, err := ioutil.ReadFile(absPath)
	//log.Println("reading from " + absPath)
	if err != nil {
		panic(err.Error())
	}
	decode := k8Sscheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*coreV1.Secret)
}

func getTektonDescriptorFromFile() []unstructured.Unstructured {
	var files []unstructured.Unstructured
	data, _ := ioutil.ReadFile("descriptor/v0_0_1_beta/prerequisites/tekton-release.yaml")
	fileAsString := string(data)[:]
	sepFiles := strings.Split(fileAsString, "---")
	for _, each := range sepFiles {
		obj := &unstructured.Unstructured{
			Object: map[string]interface{}{},
		}
		if err := yaml.Unmarshal([]byte(each), &obj.Object); err != nil {
			if err := json.Unmarshal([]byte(each), &obj.Object); err != nil {
			}
		}
		files = append(files, *obj)
	}
	return files
}

func getConfigMapFromFile() coreV1.ConfigMap {
	data, err := ioutil.ReadFile("descriptor/v0_0_1_beta/prerequisites/security-server-configmap.yaml")
	if err != nil {
		panic(err.Error())
	}
	decode := k8Sscheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*coreV1.ConfigMap)
}

func New(client client.Client, restConfig *rest.Config) service.Prerequisites {
	return prerequisites{
		Secret:           getSecretFromFile(),
		Client:           client,
		RestConfig:       restConfig,
		TektonDescriptor: getTektonDescriptorFromFile(),
		Configmap:        getConfigMapFromFile(),
	}
}
