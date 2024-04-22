/*
Copyright 2024.

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

// MultusNotifierSpec defines the desired state of MultusNotifier
type MultusNotifierSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MultusNotifier. Edit multusnotifier_types.go to remove/update
	WatchLabel      string   `json:"watchLabel,omitempty"`
	WatchNamespaces []string `json:"watchNamespaces,omitempty"`
}

// MultusNotifierStatus defines the observed state of MultusNotifier
type MultusNotifierStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PodIPList []MultusPodIPList `json:"podIPList,omitempty"`
}

type MultusPodIPList struct {
	PodName      string `json:"podName,omitempty"`
	PodNamespace string `json:"podNamespace,omitempty"`
	MultusIP     string `json:"multusIP,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MultusNotifier is the Schema for the multusnotifiers API
type MultusNotifier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MultusNotifierSpec   `json:"spec,omitempty"`
	Status MultusNotifierStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MultusNotifierList contains a list of MultusNotifier
type MultusNotifierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MultusNotifier `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MultusNotifier{}, &MultusNotifierList{})
}
