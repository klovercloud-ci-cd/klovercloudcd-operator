package service

import (
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type LightHouseCommand interface {
	ModifyConfigmap(namespace string, db v1alpha1.DB) LightHouseCommand
	ModifyDeployment(namespace string, lightHouseCommand v1alpha1.LightHouseCommand) LightHouseCommand
	ModifyService(namespace string) LightHouseCommand
	Apply(scheme *runtime.Scheme,wait bool) error
	ApplyConfigMap() error
	ApplyDeployment() error
	ApplyService() error
	Delete()error
}
