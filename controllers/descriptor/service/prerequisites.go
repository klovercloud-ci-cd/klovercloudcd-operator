package service

import "github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"

type Prerequisites interface {
	ModifySecret(namespace string, db v1alpha1.DB) Prerequisites
	ModifyTektonDescriptor(namespace string) Prerequisites
	ApplySecret(wait bool) error
	ApplyTektonDescriptor(wait bool) error
	Apply(wait bool) error
}
