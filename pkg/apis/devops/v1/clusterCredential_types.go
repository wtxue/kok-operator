package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CredentialInfo struct {
	TenantID    string `json:"tenantID" protobuf:"bytes,2,opt,name=tenantID"`
	ClusterName string `json:"clusterName" protobuf:"bytes,3,opt,name=clusterName"`

	// For TKE in global reuse
	// +optional
	ETCDCACert []byte `json:"etcdCACert,omitempty" protobuf:"bytes,4,opt,name=etcdCACert"`
	// +optional
	ETCDCAKey []byte `json:"etcdCAKey,omitempty" protobuf:"bytes,5,opt,name=etcdCAKey"`
	// +optional
	ETCDAPIClientCert []byte `json:"etcdAPIClientCert,omitempty" protobuf:"bytes,6,opt,name=etcdAPIClientCert"`
	// +optional
	ETCDAPIClientKey []byte `json:"etcdAPIClientKey,omitempty" protobuf:"bytes,7,opt,name=etcdAPIClientKey"`

	// For connect the cluster
	// +optional
	CACert []byte `json:"caCert,omitempty" protobuf:"bytes,8,opt,name=caCert"`
	// +optional
	CAKey []byte `json:"caKey,omitempty" protobuf:"bytes,9,opt,name=caKey"`
	// For kube-apiserver X509 auth
	// +optional
	ClientCert []byte `json:"clientCert,omitempty" protobuf:"bytes,10,opt,name=clientCert"`
	// For kube-apiserver X509 auth
	// +optional
	ClientKey []byte `json:"clientKey,omitempty" protobuf:"bytes,11,opt,name=clientKey"`
	// For kube-apiserver token auth
	// +optional
	Token *string `json:"token,omitempty" protobuf:"bytes,12,opt,name=token"`
	// For kubeadm init or join
	// +optional
	BootstrapToken *string `json:"bootstrapToken,omitempty" protobuf:"bytes,13,opt,name=bootstrapToken"`
	// For kubeadm init or join
	// +optional
	CertificateKey *string `json:"certificateKey,omitempty" protobuf:"bytes,14,opt,name=certificateKey"`
}

// +kubebuilder:object:root=true

// ClusterCredential records the credential information needed to access the cluster.
type ClusterCredential struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	CredentialInfo `json:",inline"`
}

// +kubebuilder:object:root=true

// ClusterCredentialList is the whole list of all ClusterCredential which owned by a tenant.
type ClusterCredentialList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of clusters
	Items []ClusterCredential `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&ClusterCredential{}, &ClusterCredentialList{})
}
