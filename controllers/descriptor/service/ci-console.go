package service

import "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"

type Console interface {
	ModifyConfigmap(namespace string) Console
	ModifyDeployment(namespace string, console v1alpha1.Console) Console
	ModifyService(namespace string) Console
	Apply(wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
}
