package service

import "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"

type EventBank interface {
	ModifyConfigmap(namespace string, db v1alpha1.DB) EventBank
	ModifyDeployment(namespace string, eventBank v1alpha1.EventBank) EventBank
	ModifyService(namespace string) EventBank
	Apply(wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
}
