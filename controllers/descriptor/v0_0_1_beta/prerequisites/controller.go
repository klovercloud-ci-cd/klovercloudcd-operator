package prerequisites

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/service"
	"io/ioutil"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type prerequisites struct {
	Secret           coreV1.Secret
	Client           client.Client
	TektonDescriptor []unstructured.Unstructured
	Error            error
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

func (p prerequisites) ApplySecret(wait bool) error {
	return p.Client.Create(context.Background(), &p.Secret)
}

func (p prerequisites) ApplyTektonDescriptor(wait bool) error {
	for _, each := range p.TektonDescriptor {
		return p.Client.Create(context.Background(), &each)
	}
	return nil
}

func (p prerequisites) Apply(wait bool) error {
	if p.Error != nil {
		return p.Error
	}
	err := p.ApplyTektonDescriptor(false)
	if err != nil {
		log.Println("[ERROR]: Failed to create tekton", err.Error())
		return err
	}
	err = p.ApplySecret(false)
	if err != nil {
		log.Println("[ERROR]: Failed to create secret ", "Secret.Namespace", p.Secret.Namespace, "Deployment.Name", p.Secret.Name, err.Error())
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

func New(client client.Client) service.Prerequisites {
	return prerequisites{
		Secret:           getSecretFromFile(),
		Client:           client,
		TektonDescriptor: getTektonDescriptorFromFile(),
	}
}
