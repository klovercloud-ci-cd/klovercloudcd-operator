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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type prerequisites struct {
	Secret           coreV1.Secret
	Client           client.Client
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

func (p prerequisites) ApplyTektonDescriptor() error {
	existingTektonController := &appsv1.Deployment{}
	err := p.Client.Get(context.Background(), types.NamespacedName{Name: "tekton-pipelines-controller", Namespace: p.Secret.Namespace}, existingTektonController)
	if err != nil && errors.IsNotFound(err) {
		for _, each := range p.TektonDescriptor {
			return p.Client.Create(context.Background(), &each)
		}
	}
	return nil
}

func (p prerequisites) Apply() error {
	if p.Error != nil {
		return p.Error
	}
	err := p.ApplyTektonDescriptor()
	if err != nil {
		log.Println("[ERROR]: Failed to create tekton", err.Error())
		return err
	}
	err = p.ApplySecret()
	if err != nil {
		log.Println("[ERROR]: Failed to create secret ", "Secret.Namespace", p.Secret.Namespace, "Deployment.Name", p.Secret.Name, err.Error())
		return err
	}
	err = p.ApplySecurityConfigMap()
	if err != nil {
		log.Println("[ERROR]: Failed to create security service configMap ", "Secret.Namespace", p.Secret.Namespace, "Deployment.Name", p.Secret.Name, err.Error())
		return err
	}
	return nil
}

func getSecretFromFile() coreV1.Secret {
	data, err := ioutil.ReadFile("mongo-secret.yaml")
	if err != nil {
		panic(err.Error())
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*coreV1.Secret)
}

func getTektonDescriptorFromFile() []unstructured.Unstructured {
	var files []unstructured.Unstructured
	data, _ := ioutil.ReadFile("tekton-release.yaml")
	fileAsString := string(data)[:]
	sepFiles := strings.Split(fileAsString, "---")
	for _, each := range sepFiles {
		obj := &unstructured.Unstructured{
			Object: map[string]interface{}{},
		}
		if err := yaml.Unmarshal([]byte(each), &obj.Object); err != nil {
			log.Println(err.Error())
			if err := json.Unmarshal([]byte(each), &obj.Object); err != nil {
				log.Println(err.Error())
			}
		}
		files = append(files, *obj)
	}
	return files
}

func getConfigMapFromFile() coreV1.ConfigMap {
	data, err := ioutil.ReadFile("security-server-configmap.yaml")
	if err != nil {
		panic(err.Error())
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*coreV1.ConfigMap)
}

func New(client client.Client) service.Prerequisites {
	return prerequisites{
		Secret:           getSecretFromFile(),
		Client:           client,
		TektonDescriptor: getTektonDescriptorFromFile(),
		Configmap:        getConfigMapFromFile(),
	}
}
