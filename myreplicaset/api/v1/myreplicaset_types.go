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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MyReplicaSetSpec defines the desired state of MyReplicaSet
type MyReplicaSetSpec struct {
	Replicas              int32 `json:"replicas,omitempty"`
	*metav1.LabelSelector `json:"selector"`
	Template              corev1.PodTemplateSpec `json:"template"`
}

// MyReplicaSetStatus defines the observed state of MyReplicaSet
type MyReplicaSetStatus struct {
	Replicas int32 `json:"replicas"`
	// AvailableReplicas    int32 `json:"availableReplicas"`
	// FullyLabeledReplicas int32 `json:"fullyLabeledReplicas"`
	// ObservedGeneration   int32 `json:"observedGeneration"`
	// ReadyReplicas        int32 `json:"readyReplicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MyReplicaSet is the Schema for the myreplicasets API
type MyReplicaSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyReplicaSetSpec   `json:"spec"`
	Status MyReplicaSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MyReplicaSetList contains a list of MyReplicaSet
type MyReplicaSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyReplicaSet //`json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyReplicaSet{}, &MyReplicaSetList{})
}
