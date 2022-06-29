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

	// Console config of Console server
	Console Console `json:"console"`

	// Terminal config of Terminal server
	Terminal Terminal `json:"terminal"`

}

// Terminal defines the config of Terminal service
type Terminal struct {
	// Enabled can be true or false.
	Enabled string  `json:"enabled"`

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// CoreEngine defines the config of CoreEngine service
type Console struct {
	// Enabled can be true or false.
	Enabled string  `json:"enabled"`

	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}


// CoreEngine defines the config of CoreEngine service
type CoreEngine struct {
	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}


// EventBank defines the config of EventBank service
type EventBank struct {
	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}


// IntegrationManager defines the config of IntegrationManager service
type IntegrationManager struct {
	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}


// Agent defines the config of Agent service
type Agent struct {
	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}


// ApiService defines the config of api service
type ApiService struct {
	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// LightHouse defines the config of lighthouse service
type LightHouse struct {
	// Enabled can be true or false.
	Enabled string  `json:"enabled"`

	// LightHouseCommand config of LightHouseCommand server
	Command LightHouseCommand `json:"command"`

	// LightHouseQuery config of LightHouseQuery server
	Query LightHouseQuery `json:"query"`
}

// LightHouse defines the config of lighthouse service
type LightHouseCommand struct {
	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// LightHouseQuery defines the config of LightHouseQuery service
type LightHouseQuery struct {
	// Resources defines cpu, memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}




// Security defines security service configuration
type Security struct {
	// User config of security server
	User User `json:"user"`

	// MailServerHostEmail mail servers host email
	MailServerHostEmail string `json:"mail_server_host_email"`

	// MailServerHostEmailSecret mail servers host emails secret
	MailServerHostEmailSecret string `json:"mail_server_host_email_secret"`

	// SMTPHost mail server smtp host
	SMTPHost string `json:"smtp_host"`

	// SMTPPort mail server smtp port
	SMTPPort string `json:"smtp_port"`

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

// KlovercloudCDStatus defines the observed state of KlovercloudCD
type KlovercloudCDStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
