package service

import "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"

type Security interface {
	ModifyConfigmap(namespace string,db v1alpha1.DB,security v1alpha1.Security) Security
	ModifyDeployment(namespace string,security v1alpha1.Security) Security
	Apply( wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
}
