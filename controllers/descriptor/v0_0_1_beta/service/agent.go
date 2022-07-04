package service

import "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"

type Agent interface {
	ModifyClusterRole() Agent
	ModifyClusterRoleBinding(namespace string,agent v1alpha1.Agent) Agent
	ModifyServiceAccount(namespace string,agent v1alpha1.Agent) Agent
	ModifyConfigmap(namespace string,agent v1alpha1.Agent) Agent
	ModifyDeployment(namespace string, agent v1alpha1.Agent) Agent
	ModifyService(namespace string) Agent
	Apply(wait bool) error
	ApplyClusterRole() error
	ApplyClusterRoleBinding() error
	ApplyServiceAccount() error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
}
