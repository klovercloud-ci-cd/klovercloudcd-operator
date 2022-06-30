package service

import "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"

type ApiService interface {
	ModifyConfigmap(namespace string) ApiService
	ModifyDeployment(namespace string, apiService v1alpha1.ApiService) ApiService
	ModifyService(namespace string) ApiService
	Apply(wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
}
