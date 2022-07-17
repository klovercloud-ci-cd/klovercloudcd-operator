package security

import (
	"context"
	"errors"
	"fmt"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/service"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/utility"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	basev1alpha1 "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
)

type security struct {
	Deployment appv1.Deployment
	Service    corev1.Service
	Client     client.Client
	Error      error
}

func (s security) ModifyService(namespace string) service.Security {
	if s.Service.ObjectMeta.Labels == nil {
		s.Service.ObjectMeta.Labels = make(map[string]string)
	}
	s.Service.ObjectMeta.Labels["app"] = "klovercloudCD"
	s.Service.ObjectMeta.Namespace = namespace
	return s
}

func (s security) ModifyDeployment(namespace string, security v1alpha1.Security) service.Security {
	if s.Deployment.ObjectMeta.Labels == nil {
		s.Deployment.ObjectMeta.Labels = make(map[string]string)
	}
	s.Deployment.ObjectMeta.Labels["app"] = "klovercloudCD"
	s.Deployment.ObjectMeta.Namespace = namespace
	if security.Resources.Requests.Cpu() != nil || security.Resources.Limits.Cpu() != nil {
		for index, _ := range s.Deployment.Spec.Template.Spec.Containers {
			s.Deployment.Spec.Template.Spec.Containers[index].Resources = security.Resources
		}
	}
	s.Deployment.Spec.Replicas = &security.Size
	return s
}

func getServiceFromFile() corev1.Service {
	data, err := ioutil.ReadFile("security-server-service.yaml")
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

func (s security) Apply(wait bool) error {
	if s.Error != nil {
		return s.Error
	}

	existingPodListObject := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(s.Deployment.Namespace),
		client.MatchingLabels(s.Deployment.ObjectMeta.Labels),
	}

	if err := s.Client.List(context.Background(), existingPodListObject, listOpts...); err != nil {
		log.Println(err, "[ERROR]: Failed to list pods", "Deployment.Namespace", s.Deployment.Namespace, "Deployment.Name", s.Deployment.Name)
	}

	existingPodMap := make(map[string]bool)

	if len(existingPodListObject.Items) > 0 {
		for _, each := range existingPodListObject.Items {
			existingPodMap[each.Name] = true
		}
	}

	config := &basev1alpha1.KlovercloudCD{}

	ctrl.SetControllerReference(config, &s.Deployment, controllers.KlovercloudCDReconciler{}.Scheme)
	err := s.ApplyDeployment()
	if err != nil {
		log.Println("[ERROR]: Failed to apply deployment for security service.", "Deployment.Namespace: ", s.Deployment.Namespace, " Deployment.Name: ", s.Deployment.Name+". ", err.Error())
		return err
	}
	if wait {
		err = utility.WaitUntilPodsAreReady(s.Client, existingPodMap, listOpts, s.Deployment.Namespace, s.Deployment.Name, *s.Deployment.Spec.Replicas, 10)
		if err != nil {
			return err
		}
	}

	ctrl.SetControllerReference(config, &s.Service, controllers.KlovercloudCDReconciler{}.Scheme)
	err = s.ApplyService()
	if err != nil {
		log.Println("[ERROR]: Failed to apply service for security service.", "Deployment.Namespace: ", s.Deployment.Namespace, " Deployment.Name: ", s.Deployment.Name+". ", err.Error())
		return err
	}

	return nil
}

func (s security) WaitUntilPodsAreReady(existingPods map[string]bool, listOption []client.ListOption, namespace string, deployment string, replica int32, retryCount int) error {

	if retryCount <= 0 {
		return errors.New("[ERROR]: Failed to watch pod lifecycle event." + "Deployment.Namespace:" + namespace + ", Deployment.Name: " + deployment)
	}

	podListObject := &corev1.PodList{}
	if err := s.Client.List(context.Background(), podListObject, listOption...); err != nil {
		log.Println(err, "Failed to list pods", "Deployment.Namespace", s.Deployment.Namespace, "Deployment.Name", s.Deployment.Name)
	}

	newlyCreatedPods := make(map[string]corev1.Pod)

	for _, each := range podListObject.Items {
		if _, ok := existingPods[each.ObjectMeta.Name]; ok {
			continue
		}
		newlyCreatedPods[each.ObjectMeta.Name] = each
	}

	if int32(len(newlyCreatedPods)) < replica {
		time.Sleep(time.Second * 5)
		retryCount = retryCount - 1
		s.WaitUntilPodsAreReady(existingPods, listOption, namespace, deployment, replica, retryCount)
	}

	if int32(len(newlyCreatedPods)) == replica {
		for _, value := range newlyCreatedPods {
			for _, containerStatus := range value.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil {
					if containerStatus.State.Waiting.Reason == "ImagePullBackOff" || containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
						return errors.New("[ERROR]: Failed to watch pod lifecycle event." + "Deployment.Namespace:" + namespace + ", Deployment.Name: " + deployment)
					}
					retryCount = retryCount - 1
					time.Sleep(time.Second * 5)
					s.WaitUntilPodsAreReady(existingPods, listOption, namespace, deployment, replica, retryCount)
				}
				if containerStatus.State.Terminated != nil {
					return errors.New("[ERROR]: Failed to watch pod lifecycle event." + "Deployment.Namespace:" + namespace + ", Deployment.Name: " + deployment)
				}
				if containerStatus.State.Running != nil {
					continue
				}
			}
		}
	}
	return nil
}

func (s security) ApplyService() error {
	return s.Client.Create(context.Background(), &s.Service)
}

func (s security) ApplyDeployment() error {
	return s.Client.Create(context.Background(), &s.Deployment)
}

func New(client client.Client) service.Security {
	return security{
		Deployment: getDeploymentFromFile(),
		Service:    getServiceFromFile(),
		Client:     client,
	}
}
