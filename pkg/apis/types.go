package apis

import (
	kubeadmv1beta2 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2"
)

// WarpperConfiguration contains a list of elements that InitConfiguration and ClusterConfiguration
type WarpperConfiguration struct {
	*kubeadmv1beta2.InitConfiguration    `json:"-"`
	*kubeadmv1beta2.ClusterConfiguration `json:"-"`
	IPs                                  []string
}
