package security

import (
	"context"
	"errors"
	"fmt"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	//ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"time"
)

type Controller interface {
	ModifyConfigmap(db v1alpha1.DB,security v1alpha1.Security)
	Apply( wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
}

type Security struct {
	Configmap corev1.ConfigMap
	Deployment appv1.Deployment
	Client client.Client
}

func (s Security) ModifyConfigmap(db v1alpha1.DB, security v1alpha1.Security) {
	panic("implement me")
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

func (s Security) Apply(wait bool) error {
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
		s.WaitUntilPodsAreReady(existingPodMap,listOpts,s.Deployment.Namespace,s.Deployment.Name,*s.Deployment.Spec.Replicas,10)
	}

	return nil
}

func(s Security) WaitUntilPodsAreReady(existingPods map[string]bool,listOption []client.ListOption, namespace string,deployment string,replica int32,retryCount int) error{

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


func (s Security) ApplyConfigMap() error {
	return s.Client.Create(context.Background(), &s.Configmap)
}

func (s Security) ApplyDeployment() error {
	return s.Client.Create(context.Background(), &s.Deployment)
}

func New(	client client.Client) Controller {
	return Security{
		Configmap:  getConfigMapFromFile(),
		Deployment: getDeploymentFromFile(),
		Client: client,
	}
}
