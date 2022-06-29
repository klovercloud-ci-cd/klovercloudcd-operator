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

	// DB config to be used.
	Database DB `json:"db"`


}


// DB defines the database configuration option
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
