package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
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

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[string]resource.Quantity

// ResourceRequirements describes the compute resource requirements.
type ResourceRequirements struct {
	Limits   ResourceList `json:"limits,omitempty" protobuf:"bytes,1,rep,name=limits,casttype=ResourceList"`
	Requests ResourceList `json:"requests,omitempty" protobuf:"bytes,2,rep,name=requests,casttype=ResourceList"`
}

// ClusterResource records the current available and maximum resource quota
// information for the cluster.
type ClusterResource struct {
	// Capacity represents the total resources of a cluster.
	// +optional
	Capacity ResourceList `json:"capacity,omitempty" protobuf:"bytes,1,rep,name=capacity,casttype=ResourceList"`
	// Allocatable represents the resources of a cluster that are available for scheduling.
	// Defaults to Capacity.
	// +optional
	Allocatable ResourceList `json:"allocatable,omitempty" protobuf:"bytes,2,rep,name=allocatable,casttype=ResourceList"`
	// +optional
	Allocated ResourceList `json:"allocated,omitempty" protobuf:"bytes,3,rep,name=allocated,casttype=ResourceList"`
}

// ClusterComponent records the number of copies of each component of the
// cluster master.
type ClusterComponent struct {
	Type     string                   `json:"type" protobuf:"bytes,1,opt,name=type"`
	Replicas ClusterComponentReplicas `json:"replicas" protobuf:"bytes,2,opt,name=replicas,casttype=ClusterComponentReplicas"`
}

// ClusterComponentReplicas records the number of copies of each state of each
// component of the cluster master.
type ClusterComponentReplicas struct {
	Desired   int32 `json:"desired" protobuf:"varint,1,name=desired"`
	Current   int32 `json:"current" protobuf:"varint,2,name=current"`
	Available int32 `json:"available" protobuf:"varint,3,name=available"`
	Updated   int32 `json:"updated" protobuf:"varint,4,name=updated"`
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
)

// ClusterCondition contains details for the current condition of this cluster.
type ClusterCondition struct {
	// Type is the type of the condition.
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

type HookType string

const (
	HookPreInstall     HookType = "PreInstall"
	HookPostInstall    HookType = "PostInstall"
	HookPostCniInstall HookType = "PostCniInstall"
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

// ClusterAddress contains information for the cluster's address.
type ClusterAddress struct {
	// Cluster address type, one of Public, ExternalIP or InternalIP.
	Type AddressType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=AddressType"`
	// The cluster address.
	Host string `json:"host" protobuf:"bytes,2,opt,name=host"`
	Port int32  `json:"port" protobuf:"varint,3,name=port"`
}

// LocalEtcd describes that kubeadm should run an etcd cluster locally
type LocalEtcd struct {
	// DataDir is the directory etcd will place its data.
	// Defaults to "/var/lib/etcd".
	DataDir string `json:"dataDir" protobuf:"bytes,1,opt,name=dataDir"`

	// ExtraArgs are extra arguments provided to the etcd binary
	// when run inside a static pod.
	ExtraArgs map[string]string `json:"extraArgs,omitempty" protobuf:"bytes,2,rep,name=extraArgs"`

	// ServerCertSANs sets extra Subject Alternative Names for the etcd server signing cert.
	ServerCertSANs []string `json:"serverCertSANs,omitempty" protobuf:"bytes,3,rep,name=serverCertSANs"`
	// PeerCertSANs sets extra Subject Alternative Names for the etcd peer signing cert.
	PeerCertSANs []string `json:"peerCertSANs,omitempty" protobuf:"bytes,4,rep,name=peerCertSANs"`
}

type HA struct {
	DKEHA        *DKEHA        `json:"dke,omitempty" protobuf:"bytes,1,opt,name=tke"`
	ThirdPartyHA *ThirdPartyHA `json:"thirdParty,omitempty" protobuf:"bytes,2,opt,name=thirdParty"`
}

type DKEHA struct {
	VIP string `json:"vip" protobuf:"bytes,1,name=vip"`
}

type ThirdPartyHA struct {
	VIP   string `json:"vip" protobuf:"bytes,1,name=vip"`
	VPort int32  `json:"vport" protobuf:"bytes,2,name=vport"`
}

type File struct {
	Src string `json:"src" protobuf:"bytes,1,name=src"` // Only support regular file
	Dst string `json:"dst" protobuf:"bytes,2,name=dst"`
}

// ClusterFeature records the features that are enabled by the cluster.
type ClusterFeature struct {
	// +optional
	IPVS *bool `json:"ipvs,omitempty" protobuf:"varint,1,opt,name=ipvs"`
	// +optional
	PublicLB *bool `json:"publicLB,omitempty" protobuf:"varint,2,opt,name=publicLB"`
	// +optional
	InternalLB *bool `json:"internalLB,omitempty" protobuf:"varint,3,opt,name=internalLB"`
	// +optional
	EnableMasterSchedule bool `json:"enableMasterSchedule,omitempty" protobuf:"bytes,5,opt,name=enableMasterSchedule"`
	// +optional
	HA *HA `json:"ha,omitempty" protobuf:"bytes,6,opt,name=ha"`
	// +optional
	SkipConditions []string `json:"skipConditions,omitempty" protobuf:"bytes,7,opt,name=skipConditions"`
	// +optional
	Files []File `json:"files,omitempty" protobuf:"bytes,8,opt,name=files"`
	// +optional
	Hooks map[HookType]string `json:"hooks,omitempty" protobuf:"bytes,9,opt,name=hooks"`
}

// ClusterProperty records the attribute information of the cluster.
type ClusterProperty struct {
	// +optional
	MaxClusterServiceNum *int32 `json:"maxClusterServiceNum,omitempty" protobuf:"bytes,1,opt,name=maxClusterServiceNum"`
	// +optional
	MaxNodePodNum *int32 `json:"maxNodePodNum,omitempty" protobuf:"bytes,2,opt,name=maxNodePodNum"`
	// +optional
	OversoldRatio map[string]string `json:"oversoldRatio,omitempty" protobuf:"bytes,3,opt,name=oversoldRatio"`
}

// ExternalEtcd describes an external etcd cluster.
// Kubeadm has no knowledge of where certificate files live and they must be supplied.
type ExternalEtcd struct {
	// Endpoints of etcd members. Required for ExternalEtcd.
	Endpoints []string `json:"endpoints" protobuf:"bytes,1,rep,name=endpoints"`

	// CAFile is an SSL Certificate Authority file used to secure etcd communication.
	// Required if using a TLS connection.
	CAFile string `json:"caFile" protobuf:"bytes,2,opt,name=caFile"`

	// CertFile is an SSL certification file used to secure etcd communication.
	// Required if using a TLS connection.
	CertFile string `json:"certFile" protobuf:"bytes,3,opt,name=certFile"`

	// KeyFile is an SSL key file used to secure etcd communication.
	// Required if using a TLS connection.
	KeyFile string `json:"keyFile" protobuf:"bytes,4,opt,name=keyFile"`
}

// Etcd contains elements describing Etcd configuration.
type Etcd struct {

	// Local provides configuration knobs for configuring the local etcd instance
	// Local and External are mutually exclusive
	Local *LocalEtcd `json:"local,omitempty" protobuf:"bytes,1,opt,name=local"`

	// External describes how to connect to an external etcd cluster
	// Local and External are mutually exclusive
	External *ExternalEtcd `json:"external,omitempty" protobuf:"bytes,2,opt,name=external"`
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// Finalizers is an opaque list of values that must be empty to permanently remove object from storage.
	// +optional
	Finalizers []FinalizerName `json:"finalizers,omitempty" protobuf:"bytes,1,rep,name=finalizers,casttype=FinalizerName"`
	TenantID   string          `json:"tenantID" protobuf:"bytes,2,opt,name=tenantID"`
	// +optional
	DisplayName string `json:"displayName" protobuf:"bytes,3,opt,name=displayName"`
	Type        string `json:"type" protobuf:"bytes,4,opt,name=type"`
	Version     string `json:"version" protobuf:"bytes,5,opt,name=version"`
	// +optional
	NetworkType NetworkType `json:"networkType,omitempty" protobuf:"bytes,6,opt,name=networkType,casttype=NetworkType"`
	// +optional
	NetworkDevice string `json:"networkDevice,omitempty" protobuf:"bytes,7,opt,name=networkDevice"`
	// +optional
	ClusterCIDR string `json:"clusterCIDR,omitempty" protobuf:"bytes,8,opt,name=clusterCIDR"`
	// ServiceCIDR is used to set a separated CIDR for k8s service, it's exclusive with MaxClusterServiceNum.
	// +optional
	ServiceCIDR *string `json:"serviceCIDR,omitempty" protobuf:"bytes,19,opt,name=serviceCIDR"`
	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string `json:"dnsDomain,omitempty" protobuf:"bytes,9,opt,name=dnsDomain"`
	// +optional
	PublicAlternativeNames []string `json:"publicAlternativeNames,omitempty" protobuf:"bytes,10,opt,name=publicAlternativeNames"`
	// +optional
	Features ClusterFeature `json:"features,omitempty" protobuf:"bytes,11,opt,name=features,casttype=ClusterFeature"`
	// +optional
	Properties ClusterProperty `json:"properties,omitempty" protobuf:"bytes,12,opt,name=properties,casttype=ClusterProperty"`
	// +optional
	Machines []*ClusterMachine `json:"machines,omitempty" protobuf:"bytes,13,rep,name=addresses"`
	// +optional
	DockerExtraArgs map[string]string `json:"dockerExtraArgs,omitempty" protobuf:"bytes,14,name=dockerExtraArgs"`
	// +optional
	KubeletExtraArgs map[string]string `json:"kubeletExtraArgs,omitempty" protobuf:"bytes,15,name=kubeletExtraArgs"`
	// +optional
	APIServerExtraArgs map[string]string `json:"apiServerExtraArgs,omitempty" protobuf:"bytes,16,name=apiServerExtraArgs"`
	// +optional
	ControllerManagerExtraArgs map[string]string `json:"controllerManagerExtraArgs,omitempty" protobuf:"bytes,17,name=controllerManagerExtraArgs"`
	// +optional
	SchedulerExtraArgs map[string]string `json:"schedulerExtraArgs,omitempty" protobuf:"bytes,18,name=schedulerExtraArgs"`
	// Etcd holds configuration for etcd.
	Etcd *Etcd `json:"etcd,omitempty" protobuf:"bytes,21,opt,name=etcd"`
	//
	Pause bool `json:"pause,omitempty"`
}

// ClusterStatus represents information about the status of a cluster.
type ClusterStatus struct {
	// +optional
	Locked *bool `json:"locked,omitempty" protobuf:"varint,1,opt,name=locked"`
	// +optional
	Version string `json:"version" protobuf:"bytes,2,opt,name=version"`
	// +optional
	Phase ClusterPhase `json:"phase,omitempty" protobuf:"bytes,3,opt,name=phase,casttype=ClusterPhase"`
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []ClusterCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,4,rep,name=conditions"`
	// A human readable message indicating details about why the cluster is in this condition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
	// A brief CamelCase message indicating details about why the cluster is in this state.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,6,opt,name=reason"`
	// List of addresses reachable to the cluster.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Addresses []ClusterAddress `json:"addresses,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,7,rep,name=addresses"`
	// +optional
	Resource ClusterResource `json:"resource,omitempty" protobuf:"bytes,9,opt,name=resource,casttype=ClusterResource"`
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Components []ClusterComponent `json:"components,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,10,rep,name=components"`
	// +optional
	ServiceCIDR string `json:"serviceCIDR,omitempty" protobuf:"bytes,11,opt,name=serviceCIDR"`
	// +optional
	NodeCIDRMaskSize int32 `json:"nodeCIDRMaskSize,omitempty" protobuf:"varint,12,opt,name=nodeCIDRMaskSize"`
	// +optional
	DNSIP string `json:"dnsIP,omitempty" protobuf:"bytes,13,opt,name=dnsIP"`
	// +optional
	RegistryIPs []string `json:"registryIPs,omitempty" protobuf:"bytes,14,opt,name=registryIPs"`
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
