package service

import "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"

type ExternalAgent interface {
	ModifyClusterRole() ExternalAgent
	ModifyClusterRoleBinding(namespace string,agent v1alpha1.Agent) ExternalAgent
	ModifyServiceAccount(namespace string,agent v1alpha1.Agent) ExternalAgent
	ModifyConfigmap(namespace string,agent v1alpha1.Agent) ExternalAgent
	ModifyDeployment(namespace string, agent v1alpha1.Agent) ExternalAgent
	ModifyService(namespace string) ExternalAgent
	Apply(wait bool) error
	ApplyClusterRole() error
	ApplyClusterRoleBinding() error
	ApplyServiceAccount() error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
}
