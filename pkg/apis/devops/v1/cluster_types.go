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

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// FinalizerName is the name identifying a finalizer during cluster lifecycle.
type FinalizerName string

const (
	// ClusterFinalize is an internal finalizer values to Cluster.
	ClusterFinalize FinalizerName = "cluster"

	// MachineFinalize is an internal finalizer values to Machine.
	MachineFinalize FinalizerName = "machine"
)

// NetworkType defines the network type of cluster.
type NetworkType string

// GPUType defines the gpu type of cluster.
type GPUType string

const (
	// GPUPhysical indicates the gpu type of cluster is physical.
	GPUPhysical GPUType = "Physical"
	// GPUVirtual indicates the gpu type of cluster is virtual.
	GPUVirtual GPUType = "Virtual"
)

// OperatingSystem defines the operating of system.
type OSType string

const (
	CentosType OSType = "centos"
	DebianType OSType = "debian"
	UbuntuType OSType = "ubuntu"
)

// RuntimeType defines the runtime of Container.
type CRIType string

const (
	DockerCRI     CRIType = "docker"
	ContainerdCRI CRIType = "containerd"
)

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[string]resource.Quantity

// ResourceRequirements describes the compute resource requirements.
type ResourceRequirements struct {
	Limits   ResourceList `json:"limits,omitempty"`
	Requests ResourceList `json:"requests,omitempty"`
}

// ClusterResource records the current available and maximum resource quota
// information for the cluster.
type ClusterResource struct {
	// Capacity represents the total resources of a cluster.
	// +optional
	Capacity ResourceList `json:"capacity,omitempty"`
	// Allocatable represents the resources of a cluster that are available for scheduling.
	// Defaults to Capacity.
	// +optional
	Allocatable ResourceList `json:"allocatable,omitempty"`
	// +optional
	Allocated ResourceList `json:"allocated,omitempty"`
}

// ClusterComponent records the number of copies of each component of the
// cluster master.
type ClusterComponent struct {
	Type     string                   `json:"type"`
	Replicas ClusterComponentReplicas `json:"replicas"`
}

// ClusterComponentReplicas records the number of copies of each state of each
// component of the cluster master.
type ClusterComponentReplicas struct {
	Desired   int32 `json:"desired"`
	Current   int32 `json:"current"`
	Available int32 `json:"available"`
	Updated   int32 `json:"updated"`
}

// ClusterPhase defines the phase of cluster constructor.
type ClusterPhase string

const (
	// ClusterRunning is the normal running phase.
	ClusterRunning ClusterPhase = "Running"
	// ClusterInitializing is the initialize phase.
	ClusterInitializing ClusterPhase = "Initializing"
	// ClusterFailed is the failed phase.
	ClusterFailed ClusterPhase = "Failed"
	// ClusterTerminating means the cluster is undergoing graceful termination.
	ClusterTerminating ClusterPhase = "Terminating"
	// ClusterNotSupport is the not support phase.
	ClusterNotSupport ClusterPhase = "NotSupport"
)

// ClusterCondition contains details for the current condition of this cluster.
type ClusterCondition struct {
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

type HookType string

const (
	HookPreInstall  HookType = "preInstall"
	HookPostInstall HookType = "postInstall"
	HookCniInstall  HookType = "cniInstall"
)

// AddressType indicates the type of cluster apiserver access address.
type AddressType string

// These are valid address type of cluster.
const (
	// AddressPublic indicates the address of the apiserver accessed from the external network.(such as public lb)
	AddressPublic AddressType = "Public"
	// AddressAdvertise indicates the address of the apiserver accessed from the worker node.(such as internal lb)
	AddressAdvertise AddressType = "Advertise"
	// AddressReal indicates the real address of one apiserver
	AddressReal AddressType = "Real"
	// AddressInternal indicates the address of the apiserver accessed from TKE control plane.
	AddressInternal AddressType = "Internal"
	// AddressSupport used for vpc lb which bind to JNS gateway as known AddressInternal
	AddressSupport AddressType = "Support"
)

type UpgradeMode string

const (
	// Upgrade nodes automatically.
	UpgradeModeAuto = UpgradeMode("Auto")
	// Manual upgrade nodes which means user need label node with `platform.tkestack.io/need-upgrade`.
	UpgradeModeManual = UpgradeMode("Manual")
)

// UpgradeStrategy used to control the upgrade process.
type UpgradeStrategy struct {
	// The maximum number of pods that can be unready during the upgrade.
	// 0% means all pods need to be ready after evition.
	// 100% means ignore any pods unready which may be used in one worker node, use this carefully!
	// default value is 0%.
	// +optional
	MaxUnready *intstr.IntOrString `json:"maxUnready,omitempty" protobuf:"bytes,1,opt,name=maxUnready"`
}

// ClusterAddress contains information for the cluster's address.
type ClusterAddress struct {
	// Cluster address type, one of Public, ExternalIP or InternalIP.
	Type AddressType `json:"type"`
	// The cluster address.
	Host string `json:"host"`
	Port int32  `json:"port"`
}

// LocalEtcd describes that kubeadm should run an etcd cluster locally
type LocalEtcd struct {
	// DataDir is the directory etcd will place its data.
	// Defaults to "/var/lib/etcd".
	DataDir string `json:"dataDir"`

	// ExtraArgs are extra arguments provided to the etcd binary
	// when run inside a static pod.
	ExtraArgs map[string]string `json:"extraArgs,omitempty"`

	// ServerCertSANs sets extra Subject Alternative Names for the etcd server signing cert.
	ServerCertSANs []string `json:"serverCertSANs,omitempty"`
	// PeerCertSANs sets extra Subject Alternative Names for the etcd peer signing cert.
	PeerCertSANs []string `json:"peerCertSANs,omitempty"`
}

type HA struct {
	DKEHA        *DKEHA        `json:"dke,omitempty"`
	ThirdPartyHA *ThirdPartyHA `json:"thirdParty,omitempty"`
}

type DKEHA struct {
	VIP string `json:"vip"`
}

type ThirdPartyHA struct {
	VIP   string `json:"vip"`
	VPort int32  `json:"vport"`
}

type File struct {
	Src string `json:"src"` // Only support regular file
	Dst string `json:"dst"`
}

// ClusterFeature records the features that are enabled by the cluster.
type ClusterFeature struct {
	// +optional
	IPVS *bool `json:"ipvs,omitempty"`
	// +optional
	PublicLB *bool `json:"publicLB,omitempty"`
	// +optional
	InternalLB *bool `json:"internalLB,omitempty" `
	// +optional
	GPUType *GPUType `json:"gpuType,omitempty" protobuf:"bytes,4,opt,name=gpuType"`
	// +optional
	EnableMasterSchedule bool `json:"enableMasterSchedule,omitempty"`
	// +optional
	HA *HA `json:"ha,omitempty"`
	// +optional
	SkipConditions []string `json:"skipConditions,omitempty"`
	// +optional
	Files []File `json:"files,omitempty"`
	// +optional
	Hooks map[HookType]string `json:"hooks,omitempty"`
}

// ClusterProperty records the attribute information of the cluster.
type ClusterProperty struct {
	// +optional
	MaxClusterServiceNum *int32 `json:"maxClusterServiceNum,omitempty"`
	// +optional
	MaxNodePodNum *int32 `json:"maxNodePodNum,omitempty"`
	// +optional
	OversoldRatio map[string]string `json:"oversoldRatio,omitempty"`
}

// ExternalEtcd describes an external etcd cluster.
// Kubeadm has no knowledge of where certificate files live and they must be supplied.
type ExternalEtcd struct {
	// Endpoints of etcd members. Required for ExternalEtcd.
	Endpoints []string `json:"endpoints"`

	// CAFile is an SSL Certificate Authority file used to secure etcd communication.
	// Required if using a TLS connection.
	CAFile string `json:"caFile"`

	// CertFile is an SSL certification file used to secure etcd communication.
	// Required if using a TLS connection.
	CertFile string `json:"certFile"`

	// KeyFile is an SSL key file used to secure etcd communication.
	// Required if using a TLS connection.
	KeyFile string `json:"keyFile"`
}

// Etcd contains elements describing Etcd configuration.
type Etcd struct {

	// Local provides configuration knobs for configuring the local etcd instance
	// Local and External are mutually exclusive
	Local *LocalEtcd `json:"local,omitempty"`

	// External describes how to connect to an external etcd cluster
	// Local and External are mutually exclusive
	External *ExternalEtcd `json:"external,omitempty"`
}

type Upgrade struct {
	// Upgrade mode, default value is Auto.
	// +optional
	Mode UpgradeMode `json:"mode,omitempty" protobuf:"bytes,1,opt,name=mode"`
	// Upgrade strategy config.
	// +optional
	Strategy UpgradeStrategy `json:"strategy,omitempty" protobuf:"bytes,2,opt,name=strategy"`
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// Finalizers is an opaque list of values that must be empty to permanently remove object from storage.
	// +optional
	Finalizers []FinalizerName `json:"finalizers,omitempty"`
	TenantID   string          `json:"tenantID"`
	// +optional
	DisplayName string      `json:"displayName,omitempty"`
	ClusterType string      `json:"clusterType,omitempty"`
	OSType      OSType      `json:"osType,omitempty"`
	CRIType     CRIType     `json:"criType,omitempty"`
	NetworkType NetworkType `json:"networkType,omitempty"`
	Version     string      `json:"version,omitempty"`
	// +optional
	NetworkDevice string `json:"networkDevice,omitempty"`
	// +optional
	ClusterCIDR string `json:"clusterCIDR,omitempty"`
	// ServiceCIDR is used to set a separated CIDR for k8s service, it's exclusive with MaxClusterServiceNum.
	// +optional
	ServiceCIDR *string `json:"serviceCIDR,omitempty"`
	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string `json:"dnsDomain,omitempty"`
	// +optional
	PublicAlternativeNames []string `json:"publicAlternativeNames,omitempty"`
	// +optional
	Features ClusterFeature `json:"features,omitempty"`
	// +optional
	Properties ClusterProperty `json:"properties,omitempty"`
	// +optional
	Machines []*ClusterMachine `json:"machines,omitempty"`
	// +optional
	DockerExtraArgs map[string]string `json:"dockerExtraArgs,omitempty"`
	// +optional
	KubeletExtraArgs map[string]string `json:"kubeletExtraArgs,omitempty"`
	// +optional
	APIServerExtraArgs map[string]string `json:"apiServerExtraArgs,omitempty"`
	// +optional
	ControllerManagerExtraArgs map[string]string `json:"controllerManagerExtraArgs,omitempty"`
	// +optional
	SchedulerExtraArgs map[string]string `json:"schedulerExtraArgs,omitempty"`
	// Etcd holds configuration for etcd.
	Etcd *Etcd `json:"etcd,omitempty"`
	// Upgrade control upgrade process.
	// +optional
	Upgrade Upgrade `json:"upgrade,omitempty"`
	// +optional
	NetworkArgs map[string]string `json:"networkArgs,omitempty"`
	// +optional
	// Pause
	Pause bool `json:"pause,omitempty"`
}

// ClusterStatus represents information about the status of a cluster.
type ClusterStatus struct {
	// +optional
	Locked *bool `json:"locked,omitempty"`
	// +optional
	Version string `json:"version"`
	// +optional
	Phase ClusterPhase `json:"phase,omitempty"`
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []ClusterCondition `json:"conditions,omitempty"`
	// A human readable message indicating details about why the cluster is in this condition.
	// +optional
	Message string `json:"message,omitempty"`
	// A brief CamelCase message indicating details about why the cluster is in this state.
	// +optional
	Reason string `json:"reason,omitempty"`
	// List of addresses reachable to the cluster.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Addresses []ClusterAddress `json:"addresses,omitempty"`
	// +optional
	Resource ClusterResource `json:"resource,omitempty"`
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Components []ClusterComponent `json:"components,omitempty"`
	// +optional
	ServiceCIDR string `json:"serviceCIDR,omitempty"`
	// +optional
	NodeCIDRMaskSize int32 `json:"nodeCIDRMaskSize,omitempty" `
	// +optional
	DNSIP string `json:"dnsIP,omitempty"`
	// +optional
	RegistryIPs []string `json:"registryIPs,omitempty"`
	// +optional
	SecondaryServiceCIDR string `json:"secondaryServiceCIDR,omitempty"`
	// +optional
	ClusterCIDR string `json:"clusterCIDR,omitempty"`
	// +optional
	SecondaryClusterCIDR string `json:"secondaryClusterCIDR,omitempty" `
	// +optional
	NodeCIDRMaskSizeIPv4 int32 `json:"nodeCIDRMaskSizeIPv4,omitempty"`
	// +optional
	NodeCIDRMaskSizeIPv6 int32 `json:"nodeCIDRMaskSizeIPv6,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true

// Cluster is the Schema for the Cluster API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=vc
// +kubebuilder:printcolumn:name="DNSIP",type="string",JSONPath=".status.dnsIP",description="The cluster dnsIP."
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".status..version",description="The version of kubernetes."
// +kubebuilder:printcolumn:name="PHASE",type="string",JSONPath=".status.phase",description="The cluter phase."
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. "
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
