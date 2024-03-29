/*

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BindingSpec defines the desired state of Binding
// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// Important: Run "make" to regenerate code after modifying this file
type BindingSpec struct {
	// BindingSecret is the name of the Binding Secret that is created with the credentials
	BindingSecret string `json:"bindingSecret,omitempty"`
	// BindingType is the type of the Binding Secret to create, 'keys' or 'config-yaml:prefix' or 'config-properties:prefix'
	BindingType string `json:"bindingType,omitempty"`
	// SecretRef is a reference to a Secret containing the credentials
	SecretRef string `json:"secretRef,omitempty"`
	// URI is the service URI that can be used to connect to the service
	URI string `json:"uri,omitempty"`
	// URIKey is the key for the URI in the secret specified by SecretRef
	URIKey string `json:"uriKey,omitempty"`
	// PasswordKey is the key for the password in the secret specified by SecretRef
	PasswordKey string `json:"passwordKey,omitempty"`
	// Username is the username to use for connecting to the service
	Username string `json:"username,omitempty"`
	// UsernameKey is the key for the username in the secret specified by SecretRef
	UsernameKey string `json:"usernameKey,omitempty"`
	// Host is the hostname or IP address for the service
	Host string `json:"host,omitempty"`
	// HostKey is the key for the host in the secret specified by SecretRef
	HostKey string `json:"hostKey,omitempty"`
	// Port is the port used by the service
	Port string `json:"port,omitempty"`
	// PortKey is the key for the port in the secret specified by SecretRef
	PortKey string `json:"portKey,omitempty"`
}

// BindingStatus defines the observed state of Binding
type BindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Binding is the Schema for the bindings API
// +k8s:openapi-gen=true
type Binding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BindingSpec   `json:"spec,omitempty"`
	Status BindingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BindingList contains a list of Binding
type BindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Binding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Binding{}, &BindingList{})
}
