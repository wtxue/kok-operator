package apis

import (
	kubeadmv1beta3 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// WarpperConfiguration contains a list of elements that InitConfiguration and ClusterConfiguration
type WarpperConfiguration struct {
	*kubeadmv1beta3.InitConfiguration    `json:"-"`
	*kubeadmv1beta3.ClusterConfiguration `json:"-"`
	IPs                                  []string
}
