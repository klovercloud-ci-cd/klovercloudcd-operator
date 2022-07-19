package service

import (
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type CoreEngine interface {
	ModifyConfigmap(namespace string, coreEngine v1alpha1.CoreEngine, db v1alpha1.DB) CoreEngine
	ModifyDeployment(namespace string, coreEngine v1alpha1.CoreEngine) CoreEngine
	ModifyService(namespace string) CoreEngine
	ModifyClusterRole(namespace string) CoreEngine
	ModifyClusterRoleBinding(namespace string) CoreEngine
	ModifyServiceAccount(namespace string) CoreEngine
	Apply(config *v1alpha1.KlovercloudCD, scheme *runtime.Scheme, wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
	ApplyClusterRole() error
	ApplyClusterRoleBinding() error
	ApplyServiceAccount() error
}
