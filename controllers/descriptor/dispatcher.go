package descriptor

import (
	"errors"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/api/v1alpha1"
	v0_0_1_betaAgent "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/agent"
	v0_0_1_betaApiService "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/api-service"
	v0_0_1_betaConsole "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/ci-console"
	v0_0_1_betaCoreEngine "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/core-engine"
	v0_0_1_betaEventBank "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/event-bank"
	v0_0_1_betaExternalAgent "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/external-agent"
	v0_0_1_betaIntegrationManager "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/integration-manager"
	v0_0_1_betaLighthouseQuery "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/light-house-query"
	v0_0_1_betaLightHouseCommand "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/lighthouse-command"
	v0_0_1_betaPrerequisites "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/prerequisites"
	v0_0_1_betaSecurity "github.com/klovercloud-ci-cd/klovercloudcd-operator/controllers/descriptor/v0_0_1_beta/security"
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/enums"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ApplySecurity(client client.Client, namespace string, db v1alpha1.DB, v1alpha1Security v1alpha1.Security, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaSecurity.New(client).ModifyDeployment(namespace, v1alpha1Security).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply security service")
}
func ApplyPrerequisites(client client.Client, namespace string, db v1alpha1.DB, v1alpha1Security v1alpha1.Security, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaPrerequisites.New(client).ModifySecret(namespace, db).ModifyTektonDescriptor(namespace).ModifySecurityConfigMap(namespace, db, v1alpha1Security).Apply()
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply prerequisites")
}
func ApplyApiService(client client.Client, namespace string, v1alpha1ApiService v1alpha1.ApiService, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaApiService.New(client).ModifyConfigmap(namespace).ModifyDeployment(namespace, v1alpha1ApiService).ModifyService(namespace).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply api service")
}

func ApplyIntegrationManager(client client.Client, namespace string, db v1alpha1.DB, integrationManager v1alpha1.IntegrationManager, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaIntegrationManager.New(client).ModifyConfigmap(namespace, db, integrationManager).ModifyDeployment(namespace, integrationManager).ModifyService(namespace).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply integration manager")
}

func ApplyEventBank(client client.Client, namespace string, db v1alpha1.DB, eventBank v1alpha1.EventBank, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaEventBank.New(client).ModifyConfigmap(namespace, db).ModifyDeployment(namespace, eventBank).ModifyService(namespace).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply event bank")
}

func ApplyAgent(client client.Client, restConfig *rest.Config, namespace string, agent v1alpha1.Agent, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaAgent.New(client, restConfig).ModifyClusterRole().ModifyServiceAccount(namespace, agent).ModifyClusterRoleBinding(namespace, agent).ModifyConfigmap(namespace, agent).ModifyDeployment(namespace, agent).ModifyService(namespace).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply agent")
}

func ApplyCoreEngine(client client.Client, namespace string, db v1alpha1.DB, coreEngine v1alpha1.CoreEngine, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaCoreEngine.New(client).ModifyConfigmap(namespace,coreEngine, db).ModifyDeployment(namespace, coreEngine).ModifyService(namespace).ModifyClusterRole(namespace).ModifyClusterRoleBinding(namespace).ModifyServiceAccount(namespace).Apply(false)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply core engine")
}

func ApplyLightHouseQuery(client client.Client, namespace string, db v1alpha1.DB, lighthouseQuery v1alpha1.LightHouseQuery, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaLighthouseQuery.New(client).ModifyConfigmap(namespace, db).ModifyDeployment(namespace, lighthouseQuery).ModifyService(namespace).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply light house query")
}

func DeleteLightHouseQuery(client client.Client, version string) error{
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaLighthouseQuery.New(client).Delete()
	}
	return errors.New("[ERROR]: Version is not valid! Failed to delete light house query")
}

func ApplyExternalAgent(client client.Client, restConfig *rest.Config, namespace string, agent v1alpha1.Agent, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaExternalAgent.New(client, restConfig).ModifyClusterRole().ModifyServiceAccount(namespace, agent).ModifyClusterRoleBinding(namespace, agent).ModifyConfigmap(namespace, agent).ModifyDeployment(namespace, agent).ModifyService(namespace).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply agent")
}

func ApplyLightHouseCommand(client client.Client, namespace string, db v1alpha1.DB, lightHouseCommand v1alpha1.LightHouseCommand, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaLightHouseCommand.New(client).ModifyConfigmap(namespace, db).ModifyDeployment(namespace, lightHouseCommand).ModifyService(namespace).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply lighthouse command")
}

func DeleteLightHouseCommand(client client.Client, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaLightHouseCommand.New(client).Delete()
	}
	return errors.New("[ERROR]: Version is not valid! Failed to delete lighthouse command")
}


func ApplyConsole(client client.Client, namespace string, console v1alpha1.Console, version string) error {
	if version == string(enums.V0_0_1_BETA) {
		return v0_0_1_betaConsole.New(client).ModifyConfigmap(namespace).ModifyDeployment(namespace, console).ModifyService(namespace).Apply(true)
	}
	return errors.New("[ERROR]: Version is not valid! Failed to apply ci console")
}
