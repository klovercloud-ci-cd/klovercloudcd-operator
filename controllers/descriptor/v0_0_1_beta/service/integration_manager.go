package service

import (
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type IntegrationManager interface {
	ModifyConfigmap(namespace string,db v1alpha1.DB,integrationManager v1alpha1.IntegrationManager) IntegrationManager
	ModifyDeployment(namespace string,integrationManager v1alpha1.IntegrationManager) IntegrationManager
	ModifyService(namespace string) IntegrationManager
	Apply(scheme *runtime.Scheme, wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
}
