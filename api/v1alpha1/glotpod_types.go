/*
Copyright 2021.

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

// GlotpodSpec defines the desired state of Glotpod
type GlotpodSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Glotpod. Edit glotpod_types.go to remove/update
	ID       int64  `json:"id,omitempty"`
	Language string `json:"language,omitempty"`
	Code     string `json:"code,omitempty"`
	Input    string `json:"input,omitempty"`
}

// GlotpodStatus defines the observed state of Glotpod
type GlotpodStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ID int64 `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Glotpod is the Schema for the glotpods API
type Glotpod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlotpodSpec   `json:"spec,omitempty"`
	Status GlotpodStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GlotpodList contains a list of Glotpod
type GlotpodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Glotpod `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Glotpod{}, &GlotpodList{})
}
