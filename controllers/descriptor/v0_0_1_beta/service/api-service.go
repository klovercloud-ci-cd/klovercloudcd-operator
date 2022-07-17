package service

import (
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ApiService interface {
	ModifyConfigmap(namespace string) ApiService
	ModifyDeployment(namespace string, apiService v1alpha1.ApiService) ApiService
	ModifyService(namespace string) ApiService
	Apply(scheme *runtime.Scheme,wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
}
