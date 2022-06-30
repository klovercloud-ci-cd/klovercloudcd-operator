package service

import "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"

type Prerequisites interface {
	ModifySecret(namespace string, db v1alpha1.DB) Prerequisites
	ModifyTektonDescriptor(namespace string) Prerequisites
	ModifySecurityConfigMap(namespace string, db v1alpha1.DB, security v1alpha1.Security) Prerequisites
	ApplySecret() error
	ApplyTektonDescriptor() error
	ApplySecurityConfigMap() error
	Apply() error
}
