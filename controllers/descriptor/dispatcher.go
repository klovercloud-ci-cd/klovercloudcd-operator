package descriptor

import (
	"errors"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	v0_0_1_betaApiService "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/api-service"
	v0_0_1_betaPrerequisites "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/prerequisites"
	v0_0_1_betaSecurity "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/security"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/enums"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ApplySecurity(client client.Client, namespace string, db v1alpha1.DB, v1alpha1Security v1alpha1.Security, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaSecurity.New(client).ModifyConfigmap(namespace, db, v1alpha1Security).ModifyDeployment(namespace, v1alpha1Security).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply security service")
}
func ApplyPrerequisites(client client.Client, namespace string, db v1alpha1.DB, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaPrerequisites.New(client).ModifySecret(namespace, db).ModifyTektonDescriptor(namespace).Apply(false)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply prerequisites")
}
func ApplyApiService(client client.Client, namespace string, db v1alpha1.DB, v1alpha1ApiService v1alpha1.ApiService, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaApiService.New(client).ModifyConfigmap(namespace).ModifyDeployment(namespace, v1alpha1ApiService).ModifyService(namespace).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply api service")
}
