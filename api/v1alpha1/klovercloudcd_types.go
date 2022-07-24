/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/klovercloud-ci-cd/klovercloudcd-operator/enums"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KlovercloudCDSpec defines the desired state of KlovercloudCD
type KlovercloudCDSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// KlovercloudCD Version, default Latest. See available in enums.VERSIONS
	Version enums.VERSIONS `json:"version,omitempty"`

	// DB config to be used
	Database DB `json:"db"`

	// Security service config
	Security Security `json:"security"`

	// LightHouse config of lighthouse server
	LightHouse LightHouse `json:"light_house,omitempty"`

	// ApiService config of api server
	ApiService ApiService `json:"api_service,omitempty"`

	// Agent config of agent server
	Agent Agent `json:"agent,omitempty"`

	// IntegrationManager config of IntegrationManager server
	IntegrationManager IntegrationManager `json:"integration_manager,omitempty"`

	// EventBank config of EventBank server
	EventBank EventBank `json:"event_bank"`

	// CoreEngine config of CoreEngine server
	CoreEngine CoreEngine `json:"core_engine"`

	// Terminal config of Terminal server
	Terminal Terminal `json:"terminal"`
}

// KlovercloudCDStatus defines the observed state of KlovercloudCD
type KlovercloudCDStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// SecurityPods are the names of the Security pods
	SecurityPods []string `json:"security_pods"`

	// LightHouseQueryPods are the names of the LightHouseQuery pods
	LightHouseQueryPods []string `json:"light_house_query_pods"`

	// LightHouseCommandPods are the names of the LightHouseCommand pods
	LightHouseCommandPods []string `json:"light_house_command_pods"`

	// ApiServicePods are the names of the ApiService pods
	ApiServicePods []string `json:"api_service_pods"`

	// AgentPods are the names of the Agent pods
	AgentPods []string `json:"agent_pods"`

	// IntegrationManagerPods are the names of the IntegrationManager pods
	IntegrationManagerPods []string `json:"integration_manager_pods"`

	// EventBankPods are the names of the EventBank pods
	EventBankPods []string `json:"event_bank_pods"`

	// CoreEnginePods are the names of the CoreEngine pods
	CoreEnginePods []string `json:"core_engine_pods"`

	// TerminalPods are the names of the Terminal pods
	TerminalPods []string `json:"terminal_pods"`
}

// Terminal defines the config of Terminal service
type Terminal struct {
	// Enabled can be true or false.
	Enabled string `json:"enabled"`

	// Size is the number of instance
	Size int32 `json:"size"`

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// CoreEngine defines the config of CoreEngine service
type CoreEngine struct {

	// Size is the number of instance
	Size int32 `json:"size"`

	// NumberOfConcurrentProcess is the number of concurrent jobs for (build,jenkins,intermediary)
	NumberOfConcurrentProcess int `json:"number_of_concurrent_process"`

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// EventBank defines the config of EventBank service
type EventBank struct {
	// Size is the number of instance
	Size int32 `json:"size"`

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// IntegrationManager defines the config of IntegrationManager service
type IntegrationManager struct {

	// Size is the number of instance
	Size int32 `json:"size"`

	// PerDayTotalProcess defines allowed per day total process
	PerDayTotalProcess string `json:"per_day_total_process"` //default 30

	// ConcurrentProcess defines concurrent total process
	ConcurrentProcess string `json:"concurrent_process"` //default 10

	// GithubWebhookConsumingUrl defines GitHub webhook consuming url
	GithubWebhookConsumingUrl string `json:"github_webhook_consuming_url"`

	// BitbucketWebhookConsumingUrl defines Bitbucket webhook consuming url
	BitbucketWebhookConsumingUrl string `json:"bitbucket_webhook_consuming_url"`

	// PipelinePurging defines if all objects will be purged after process finished
	PipelinePurging string `json:"pipeline_purging"` // default ENABLE

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// Agent defines the config of Agent service
type Agent struct {

	// Size is the number of instance
	Size int32 `json:"size"`

	// PullSize defines how many jobs it will pull every period. It should depend on consumed resources
	PullSize string `json:"pull_size,omitempty"`

	// Token defines token to communicate with api service. Generate this by doing exec inside api service, then run kcpctl generate-jwt client={your agent name}
	Token string `json:"token,omitempty"`

	// LightHouseEnabled defines if Light House is enabled or not. By default, it is false.
	LightHouseEnabled string `json:"light_house_enabled,omitempty"`

	// TerminalBaseUrl defines base url of terminal. LightHouseEnabled should be true for this feature.
	TerminalBaseUrl string `json:"terminal_base_url,omitempty"`

	// TerminalApiVersion defines the api version of terminal. By default, it is api/v1
	TerminalApiVersion string `json:"terminal_api_version,omitempty"`

	// EventStoreUrl defines the event bank url. For external agent, it should be api service base url with api version
	EventStoreUrl string `json:"event_store_url,omitempty"`

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// ApiService defines the config of api service
type ApiService struct {

	// Size is the number of instance
	Size int32 `json:"size"`

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// LightHouse defines the config of lighthouse service
type LightHouse struct {
	// Enabled can be true or false.
	Enabled string `json:"enabled,omitempty"`

	// LightHouseCommand config of LightHouseCommand server
	Command LightHouseCommand `json:"command"`

	// LightHouseQuery config of LightHouseQuery server
	Query LightHouseQuery `json:"query"`
}

// LightHouse defines the config of lighthouse service
type LightHouseCommand struct {

	// Size is the number of instance
	Size int32 `json:"size"`

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// LightHouseQuery defines the config of LightHouseQuery service
type LightHouseQuery struct {

	// Size is the number of instance
	Size int32 `json:"size"`

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// Security defines security service configuration
type Security struct {
	// User config of security server
	User User `json:"user"`

	// MailServerHostEmail mail servers host email
	MailServerHostEmail string `json:"mail_server_host_email,omitempty"`

	// MailServerHostEmailSecret mail servers host emails secret
	MailServerHostEmailSecret string `json:"mail_server_host_email_secret,omitempty"`

	// SMTPHost mail server smtp host
	SMTPHost string `json:"smtp_host,omitempty"`

	// SMTPPort mail server smtp port
	SMTPPort string `json:"smtp_port,omitempty"`

	// Size is the number of instance
	Size int32 `json:"size"`

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// User defines the defualt user of security service
type User struct {
	// FirstName firstname of default user
	FirstName string `json:"first_name"`

	// LastName lastname of default user
	LastName string `json:"last_name"`

	// Email email of default user
	Email string `json:"email"`

	// Password password of default user
	Password string `json:"password"`

	// Phone phone number of default user
	Phone string `json:"phone"`

	// CompanyName company name of default user
	CompanyName string `json:"company_name"`
}

// DB defines the database configurations
type DB struct {
	// Type of database , dafault MONGO. See supported Databases in enums.DATABASE_OPTION
	Type enums.DATABASE_OPTION `json:"type,omitempty"`

	// UserName of database server
	UserName string `json:"user_name,omitempty"`

	// Password of database server
	Password string `json:"password,omitempty"`

	// ServerURL represents database server url
	ServerURL string `json:"server_url,omitempty"`

	// ServerPort represents database server port
	ServerPort string `json:"server_port,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KlovercloudCD is the Schema for the klovercloudcds API
type KlovercloudCD struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KlovercloudCDSpec   `json:"spec,omitempty"`
	Status KlovercloudCDStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KlovercloudCDList contains a list of KlovercloudCD
type KlovercloudCDList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KlovercloudCD `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KlovercloudCD{}, &KlovercloudCDList{})
}
