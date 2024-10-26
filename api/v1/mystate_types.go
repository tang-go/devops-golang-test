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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MyStateSpec defines the desired state of MyState.
type MyStateSpec struct {
	// Replicas is the number of pod
	Replicas int `json:"replicas,omitempty"`

	// Selector is the labels to match pods
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Template is the template of pod
	Template PodTemplateSpec `json:"template,omitempty"`

	// VolumeClaimTemplates is the templates of pvc in pods
	VolumeClaimTemplates []PVCTemplate `json:"volumeClaimTemplates,omitempty"`

	// ServiceName is the service who serves pods of this statefulSet
	ServiceName string `json:"serviceName,omitempty"`

	// todo
	// UpdateStrategy StatefulSetUpdateStrategy

	// todo
	// MinReadySeconds int

	Ordinals MyOrdinals `json:"ordinals"`
}

type PodTemplateSpec struct {
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:XPreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec v1.PodSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type PVCTemplate struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:XPreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec defines the desired characteristics of a volume requested by a pod author.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
	// +optional
	Spec v1.PersistentVolumeClaimSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// status represents the current information/status of a persistent volume claim.
	// Read-only.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
	// +optional
	Status v1.PersistentVolumeClaimStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type MyOrdinals struct {
	Start int `json:"start"`
}

// MyStateStatus defines the observed state of MyState.
type MyStateStatus struct {
	// CurrentGeneration is the revision of spec which is not updated yet
	CurrentGeneration string `json:"currentRevision,omitempty"`

	// Replicas is all pods created by this myState
	Replicas int `json:"replicas"`

	// Replicas is all ready pods created by this myState
	ReadyReplicas int `json:"readyReplicas"`

	// CurrentReplicas is pods with currentVersion created by this myState
	CurrentReplicas int `json:"currentReplicas"`

	// UpdatedReplicas is pods with updateVersion created by this myState
	UpdatedReplicas int `json:"updatedReplicas"`

	// todo
	// AvailableReplicas is all available pods created by this myState
	// AvailableReplicas int `json:"availableReplicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MyState is the Schema for the mystates API.
type MyState struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyStateSpec   `json:"spec,omitempty"`
	Status MyStateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MyStateList contains a list of MyState.
type MyStateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyState `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyState{}, &MyStateList{})
}
