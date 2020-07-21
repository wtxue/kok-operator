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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CredentialInfo struct {
	TenantID    string `json:"tenantID"`
	ClusterName string `json:"clusterName"`

	// For TKE in global reuse
	// +optional
	ETCDCACert []byte `json:"etcdCACert,omitempty"`
	// +optional
	ETCDCAKey []byte `json:"etcdCAKey,omitempty"`
	// +optional
	ETCDAPIClientCert []byte `json:"etcdAPIClientCert,omitempty"`
	// +optional
	ETCDAPIClientKey []byte `json:"etcdAPIClientKey,omitempty"`

	// For connect the cluster
	// +optional
	CACert []byte `json:"caCert,omitempty"`
	// +optional
	CAKey []byte `json:"caKey,omitempty"`
	// For kube-apiserver X509 auth
	// +optional
	ClientCert []byte `json:"clientCert,omitempty"`
	// For kube-apiserver X509 auth
	// +optional
	ClientKey []byte `json:"clientKey,omitempty"`
	// For kube-apiserver token auth
	// +optional
	Token *string `json:"token,omitempty"`
	// For kubeadm init or join
	// +optional
	BootstrapToken *string `json:"bootstrapToken,omitempty"`
	// For kubeadm init or join
	// +optional
	CertificateKey *string `json:"certificateKey,omitempty"`

	ExtData         map[string]string `json:"extData,omitempty"`
	KubeData        map[string]string `json:"kubeData,omitempty"`
	ManifestsData   map[string]string `json:"manifestsData,omitempty"`
	CertsBinaryData map[string][]byte `json:"certsBinaryData,omitempty"`
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
