package service

import (
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Console interface {
	ModifyConfigmap(namespace string) Console
	ModifyDeployment(namespace string, console v1alpha1.Console) Console
	ModifyService(namespace string) Console
	Apply(config *v1alpha1.KlovercloudCD, scheme *runtime.Scheme, wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
}
