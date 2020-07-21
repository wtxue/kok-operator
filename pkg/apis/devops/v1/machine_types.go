/*
Copyright 2020 wtxue.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachinePhase defines the phase of machine constructor
type MachinePhase string

const (
	// MachineRunning is the normal running phase
	MachineRunning MachinePhase = "Running"
	// MachineInitializing is the initialize phase
	MachineInitializing MachinePhase = "Initializing"
	// MachineFailed is the failed phase
	MachineFailed MachinePhase = "Failed"
	// MachineTerminating is the terminating phase
	MachineTerminating MachinePhase = "Terminating"
)

// MachineAddressType represents the type of machine address.
type MachineAddressType string

// These are valid address type of machine.
const (
	MachineHostName    MachineAddressType = "Hostname"
	MachineExternalIP  MachineAddressType = "ExternalIP"
	MachineInternalIP  MachineAddressType = "InternalIP"
	MachineExternalDNS MachineAddressType = "ExternalDNS"
	MachineInternalDNS MachineAddressType = "InternalDNS"
)

// MachineAddress contains information for the machine's address.
type MachineAddress struct {
	// Machine address type, one of Public, ExternalIP or InternalIP.
	Type MachineAddressType `json:"type"`
	// The machine address.
	Address string `json:"address"`
}

// MachineSystemInfo is a set of ids/uuids to uniquely identify the node.
type MachineSystemInfo struct {
	// MachineID reported by the node. For unique machine identification
	// in the cluster this field is preferred. Learn more from man(5)
	// machine-id: http://man7.org/linux/man-pages/man5/machine-id.5.html
	MachineID string `json:"machineID,omitempty"`
	// SystemUUID reported by the node. For unique machine identification
	// MachineID is preferred. This field is specific to Red Hat hosts
	// https://access.redhat.com/documentation/en-US/Red_Hat_Subscription_Management/1/html/RHSM/getting-system-uuid.html
	SystemUUID string `json:"systemUUID,omitempty"`
	// Boot ID reported by the node.
	BootID string `json:"bootID,omitempty"`
	// Kernel Version reported by the node.
	KernelVersion string `json:"kernelVersion,omitempty"`
	// OS Image reported by the node.
	OSImage string `json:"osImage,omitempty"`
	// ContainerRuntime Version reported by the node.
	ContainerRuntimeVersion string `json:"containerRuntimeVersion,omitempty"`
	// Kubelet Version reported by the node.
	KubeletVersion string `json:"kubeletVersion,omitempty"`
	// KubeProxy Version reported by the node.
	KubeProxyVersion string `json:"kubeProxyVersion,omitempty"`
	// The Operating System reported by the node
	OperatingSystem string `json:"operatingSystem,omitempty"`
	// The Architecture reported by the node
	Architecture string `json:"architecture,omitempty"`
}

// MachineCondition contains details for the current condition of this Machine.
type MachineCondition struct {
	// Type is the type of the condition.
	Type string `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type MachineFeature struct {
	// +optional
	SkipConditions []string `json:"skipConditions,omitempty"`
	// +optional
	Files []File `json:"files,omitempty"`
	// +optional
	Hooks map[string]string `json:"hooks,omitempty"`
}

// MachineSpec is a description of machine.
type MachineSpec struct {
	// Finalizers is an opaque list of values that must be empty to permanently remove object from storage.
	// +optional
	Finalizers  []FinalizerName `json:"finalizers,omitempty"`
	TenantID    string          `json:"tenantID,omitempty"`
	ClusterName string          `json:"clusterName"`
	Type        string          `json:"type"`
	Machine     *ClusterMachine `json:"machine,omitempty"`
	Feature     *MachineFeature `json:"feature,omitempty"`
	Pause       bool            `json:"pause,omitempty"`
}

// MachineStatus represents information about the status of an machine.
type MachineStatus struct {
	// +optional
	Locked *bool `json:"locked,omitempty"`
	// +optional
	Phase MachinePhase `json:"phase,omitempty"`
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []MachineCondition `json:"conditions,omitempty"`
	// A human readable message indicating details about why the machine is in this condition.
	// +optional
	Message string `json:"message,omitempty"`
	// A brief CamelCase message indicating details about why the machine is in this state.
	// +optional
	Reason string `json:"reason,omitempty"`
	// List of addresses reachable to the machine.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Addresses []MachineAddress `json:"addresses,omitempty"`
	// Set of ids/uuids to uniquely identify the node.
	// +optional
	MachineInfo MachineSystemInfo `json:"machineInfo,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true

// Machine is the Schema for the Machine API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=mc
// +kubebuilder:printcolumn:name="PHASE",type="string",JSONPath=".status.phase",description="The cluter phase."
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. "
type Machine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineSpec   `json:"spec,omitempty"`
	Status MachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MachineList contains a list of Machine
type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Machine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Machine{}, &MachineList{})
}
