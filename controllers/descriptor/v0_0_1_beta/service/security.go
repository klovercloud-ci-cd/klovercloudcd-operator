package service

import "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"

type Security interface {
	ModifyDeployment(namespace string, security v1alpha1.Security) Security
	ModifyService(namespace string) Security
	Apply(wait bool) error
	ApplyDeployment() error
	ApplyService() error
}
