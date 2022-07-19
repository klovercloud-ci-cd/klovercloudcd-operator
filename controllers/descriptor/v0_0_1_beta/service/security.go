package service

import (
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Security interface {
	ModifyDeployment(namespace string, security v1alpha1.Security) Security
	ModifyService(namespace string) Security
	Apply(config *v1alpha1.KlovercloudCD, scheme *runtime.Scheme, wait bool) error
	ApplyDeployment() error
	ApplyService() error
}
