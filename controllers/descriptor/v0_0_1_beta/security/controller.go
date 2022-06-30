package security

import (
	"context"
	"errors"
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
	"time"
)


type security struct {
	Configmap corev1.ConfigMap
	Deployment appv1.Deployment
	Service corev1.Service
	Client client.Client
	Error error
}

func (s security) ModifyService(namespace string) service.Security {
	if s.Service.ObjectMeta.Labels==nil{
		s.Service.ObjectMeta.Labels=make(map[string]string)
	}
	s.Service.ObjectMeta.Labels["app"]="klovercloudCD"
	s.Service.ObjectMeta.Namespace=namespace
	return s
}

func (s security) ApplyService() error {
	return s.Client.Create(context.Background(), &s.Service)
}

func (s security) ModifyDeployment(namespace string,security v1alpha1.Security)  service.Security  {
	if s.Deployment.ObjectMeta.Labels==nil{
		s.Deployment.ObjectMeta.Labels=make(map[string]string)
	}
	s.Deployment.ObjectMeta.Labels["app"]="klovercloudCD"
	s.Deployment.ObjectMeta.Namespace=namespace
	for i,_:=range s.Deployment.Spec.Template.Spec.Containers{
		s.Deployment.Spec.Template.Spec.Containers[i].Resources=security.Resources
	}
	return s
}

func (s security) ModifyConfigmap(namespace string,db v1alpha1.DB, security v1alpha1.Security)  service.Security  {
	if s.Configmap.ObjectMeta.Labels==nil{
		s.Configmap.ObjectMeta.Labels=make(map[string]string)
	}
	s.Configmap.ObjectMeta.Labels["app"]="klovercloudCD"
	s.Configmap.ObjectMeta.Namespace=namespace
	s.Configmap.Data["MAIL_SERVER_HOST_EMAIL"]=security.MailServerHostEmail
	s.Configmap.Data["MAIL_SERVER_HOST_EMAIL_SECRET"]=security.MailServerHostEmailSecret
	s.Configmap.Data["SMTP_HOST"]=security.SMTPHost
	s.Configmap.Data["SMTP_PORT"]=security.SMTPPort
	s.Configmap.Data["USER_FIRST_NAME"]=security.User.FirstName
	s.Configmap.Data["USER_LAST_NAME"]=security.User.LastName
	s.Configmap.Data["USER_EMAIL"]=security.User.Email
	s.Configmap.Data["USER_PHONE"]=security.User.Phone
	s.Configmap.Data["USER_PASSWORD"]=security.User.Password
	s.Configmap.Data["COMPANY_NAME"]=security.User.CompanyName
	private, public, err := utility.New().Generate()
	if err!=nil{
		s.Error=err
		log.Println("[ERROR]: Failed to modify secrets configmap."+err.Error())
	}
	s.Configmap.Data["PRIVATE_KEY"]=string(private)
	s.Configmap.Data["PUBLIC_KEY"]=string(public)
	if db.Type==enums.MONGO {
		s.Configmap.Data["MONGO"] = string(enums.MONGO)
		s.Configmap.Data["MONGO_SERVER"] = db.ServerURL
		s.Configmap.Data["MONGO_PORT"] = db.ServerPort
	}
	return s
}

func  getConfigMapFromFile() corev1.ConfigMap {
	data, err := ioutil.ReadFile("security-server-configmap.yaml")
	if err!=nil{
		panic(err.Error())
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*corev1.ConfigMap)
}

func  getServiceFromFile() corev1.Service {
	data, err := ioutil.ReadFile("security-server-service.yaml")
	if err!=nil{
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
	data, err := ioutil.ReadFile("security-server-deployment.yaml")
	if err!=nil{
		panic(err.Error())
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return *obj.(*appv1.Deployment)
}

func (s security) Apply(wait bool) error {
	if s.Error!=nil{
		return s.Error
	}
	err:=s.ApplyConfigMap()
	if err!=nil{
		log.Println("[ERROR]: Failed to create configmap for security service.", "Deployment.Namespace", s.Deployment.Namespace, "Deployment.Name", s.Deployment.Name,err.Error())
		return err
	}

	existingPodListObject := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(s.Deployment.Namespace),
		client.MatchingLabels(s.Deployment.ObjectMeta.Labels),
	}

	if err = s.Client.List(context.Background(), existingPodListObject, listOpts...); err != nil {
		log.Println(err, "[ERROR]: Failed to list pods", "Deployment.Namespace", s.Deployment.Namespace, "Deployment.Name", s.Deployment.Name)
	}

	existingPodMap:=make(map[string]bool)

	if len(existingPodListObject.Items)>0{
		for _,each:=range existingPodListObject.Items{
			existingPodMap[each.Name]=true
		}
	}
	err=s.ApplyDeployment()
	if err!=nil{
		log.Println("[ERROR]: Failed to apply deployment for security service.","Deployment.Namespace: ", s.Deployment.Namespace, " Deployment.Name: ", s.Deployment.Name+". ",err.Error())
		return err
	}
	if wait{
		err=s.WaitUntilPodsAreReady(existingPodMap,listOpts,s.Deployment.Namespace,s.Deployment.Name,*s.Deployment.Spec.Replicas,10)
		if err!=nil{
			return err
		}
	}
	err=s.ApplyService()
	if err!=nil{
		log.Println("[ERROR]: Failed to apply service for security service.","Deployment.Namespace: ", s.Deployment.Namespace, " Deployment.Name: ", s.Deployment.Name+". ",err.Error())
		return err
	}

	return nil
}

func(s security) WaitUntilPodsAreReady(existingPods map[string]bool,listOption []client.ListOption, namespace string,deployment string,replica int32,retryCount int) error{

	if retryCount<=0{
		return errors.New("[ERROR]: Failed to watch pod lifecycle event."+ "Deployment.Namespace:"+namespace+", Deployment.Name: "+deployment)
	}


	podListObject := &corev1.PodList{}
	if err := s.Client.List(context.Background(), podListObject, listOption...); err != nil {
		log.Println(err, "Failed to list pods", "Deployment.Namespace", s.Deployment.Namespace, "Deployment.Name", s.Deployment.Name)
	}

	newlyCreatedPods:=make(map[string]corev1.Pod)

	for _,each:=range podListObject.Items{
		if _,ok:=existingPods[each.ObjectMeta.Name];ok{
			continue
		}
		newlyCreatedPods[each.ObjectMeta.Name]=each
	}

	if int32(len(newlyCreatedPods))<replica{
		time.Sleep(time.Second*5)
		retryCount=retryCount-1
		s.WaitUntilPodsAreReady(existingPods,listOption,namespace,deployment,replica,retryCount)
	}

	if int32(len(newlyCreatedPods))==replica{
		for _,value:=range newlyCreatedPods{
				for _,containerStatus:= range value.Status.ContainerStatuses{
					if containerStatus.State.Waiting!=nil{
						if containerStatus.State.Waiting.Reason=="ImagePullBackOff" || containerStatus.State.Waiting.Reason=="CrashLoopBackOff"{
							return errors.New("[ERROR]: Failed to watch pod lifecycle event."+ "Deployment.Namespace:"+namespace+", Deployment.Name: "+deployment)
						}
						retryCount=retryCount-1
						time.Sleep(time.Second*5)
						s.WaitUntilPodsAreReady(existingPods,listOption,namespace,deployment,replica,retryCount)
					}
					if containerStatus.State.Terminated!=nil{
						return errors.New("[ERROR]: Failed to watch pod lifecycle event."+ "Deployment.Namespace:"+namespace+", Deployment.Name: "+deployment)
					}
					if containerStatus.State.Running!=nil{
						continue
					}
				}
		}
	}
	return nil
}


func (s security) ApplyConfigMap() error {
	return s.Client.Create(context.Background(), &s.Configmap)
}

func (s security) ApplyDeployment() error {
	return s.Client.Create(context.Background(), &s.Deployment)
}

func New(client client.Client) service.Security {
	return security{
		Configmap:  getConfigMapFromFile(),
		Deployment: getDeploymentFromFile(),
		Service: getServiceFromFile(),
		Client: client,
	}
}
