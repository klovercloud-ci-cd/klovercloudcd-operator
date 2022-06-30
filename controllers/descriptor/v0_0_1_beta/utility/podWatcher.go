package utility

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func WaitUntilPodsAreReady(client client.Client, existingPods map[string]bool, listOption []client.ListOption, namespace string, deployment string, replica int32, retryCount int) error {

	if retryCount <= 0 {
		return errors.New("[ERROR]: Failed to watch pod lifecycle event." + "Deployment.Namespace:" + namespace + ", Deployment.Name: " + deployment)
	}

	podListObject := &corev1.PodList{}
	if err := client.List(context.Background(), podListObject, listOption...); err != nil {
		log.Println(err, "Failed to list pods", "Deployment.Namespace", namespace, "Deployment.Name", deployment)
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
		WaitUntilPodsAreReady(client, existingPods, listOption, namespace, deployment, replica, retryCount)
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
					WaitUntilPodsAreReady(client, existingPods, listOption, namespace, deployment, replica, retryCount)
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
