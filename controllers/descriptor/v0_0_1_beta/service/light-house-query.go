package service

import (
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type LightHouseQuery interface {
	ModifyConfigmap(namespace string, db v1alpha1.DB) LightHouseQuery
	ModifyDeployment(namespace string, lightHouseQuery v1alpha1.LightHouseQuery) LightHouseQuery
	ModifyService(namespace string) LightHouseQuery
	Apply(scheme *runtime.Scheme,wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
	Delete()error
}
