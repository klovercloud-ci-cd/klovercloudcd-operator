package service

import (
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type EventBank interface {
	ModifyConfigmap(namespace string, db v1alpha1.DB) EventBank
	ModifyDeployment(namespace string, eventBank v1alpha1.EventBank) EventBank
	ModifyService(namespace string) EventBank
	Apply(config *v1alpha1.KlovercloudCD, scheme *runtime.Scheme, wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
}
