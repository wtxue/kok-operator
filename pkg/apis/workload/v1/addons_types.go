package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AddonsSpec is a description of Addons.
type AddonsSpec struct {
	Foo string `json:"foo,omitempty"`
}

// AddonsStatus represents information about the status of an Addons.
type AddonsStatus struct {
	Phase string `json:"phase,omitempty"`
	Foo   string `json:"foo,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true

// Addons is the Schema for the Addon API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PHASE",type="string",JSONPath=".status.phase",description="The Addons phase."
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. "
type Addons struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonsSpec   `json:"spec,omitempty"`
	Status AddonsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AddonsList contains a list of Addons
type AddonsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Addons `json:"items"`
}
