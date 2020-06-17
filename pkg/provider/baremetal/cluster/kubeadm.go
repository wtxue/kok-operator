package cluster

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	kubeadmv1beta2 "github.com/wtxue/kube-on-kube-operator/pkg/apis/kubeadm/v1beta2"
	kubeletv1beta1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/kubelet/config/v1beta1"
	kubeproxyv1alpha1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/kubeproxy/config/v1alpha1"
	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/phases/kubeadm"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/json"
	corev1 "k8s.io/api/core/v1"
)

func (p *Provider) getKubeadmConfig(c *provider.Cluster) *kubeadm.Config {
	config := new(kubeadm.Config)
	config.InitConfiguration = p.getInitConfiguration(c)
	config.ClusterConfiguration = p.getClusterConfiguration(c)
	config.KubeProxyConfiguration = p.getKubeProxyConfiguration(c)
	config.KubeletConfiguration = p.getKubeletConfiguration(c)

	return config
}

func (p *Provider) getInitConfiguration(c *provider.Cluster) *kubeadmv1beta2.InitConfiguration {
	token, _ := kubeadmv1beta2.NewBootstrapTokenString(*c.ClusterCredential.BootstrapToken)

	return &kubeadmv1beta2.InitConfiguration{
		BootstrapTokens: []kubeadmv1beta2.BootstrapToken{
			{
				Token:       token,
				Description: "dke kubeadm bootstrap token",
				TTL:         &metav1.Duration{Duration: 0},
			},
		},
		NodeRegistration: kubeadmv1beta2.NodeRegistrationOptions{
			Name: c.Spec.Machines[0].IP,
		},
		LocalAPIEndpoint: kubeadmv1beta2.APIEndpoint{
			AdvertiseAddress: c.Spec.Machines[0].IP,
			BindPort:         6443,
		},
		CertificateKey: *c.ClusterCredential.CertificateKey,
	}
}

func (p *Provider) getClusterConfiguration(c *provider.Cluster) *kubeadmv1beta2.ClusterConfiguration {
	controlPlaneEndpoint := fmt.Sprintf("%s:6443", c.Spec.Machines[0].IP)

	// //  use vip
	// addr := c.Address(devopsv1.AddressAdvertise)
	// if addr != nil {
	// 	controlPlaneEndpoint = fmt.Sprintf("%s:%d", addr.Host, addr.Port)
	// }

	kubernetesVolume := kubeadmv1beta2.HostPathMount{
		Name:      "vol-dir-0",
		HostPath:  "/etc/kubernetes",
		MountPath: "/etc/kubernetes",
	}

	auditVolume := kubeadmv1beta2.HostPathMount{
		Name:      "audit-dir-0",
		HostPath:  "/var/log/kubernetes",
		MountPath: "/var/log/kubernetes",
		PathType:  corev1.HostPathDirectoryOrCreate,
	}

	config := &kubeadmv1beta2.ClusterConfiguration{
		CertificatesDir: constants.CertificatesDir,
		Networking: kubeadmv1beta2.Networking{
			DNSDomain:     c.Spec.DNSDomain,
			ServiceSubnet: c.Status.ServiceCIDR,
		},
		KubernetesVersion:    c.Spec.Version,
		ControlPlaneEndpoint: controlPlaneEndpoint,
		APIServer: kubeadmv1beta2.APIServer{
			ControlPlaneComponent: kubeadmv1beta2.ControlPlaneComponent{
				ExtraArgs:    p.getAPIServerExtraArgs(c),
				ExtraVolumes: []kubeadmv1beta2.HostPathMount{kubernetesVolume, auditVolume},
			},
			CertSANs: GetAPIServerCertSANs(c.Cluster),
		},
		ControllerManager: kubeadmv1beta2.ControlPlaneComponent{
			ExtraArgs:    p.getControllerManagerExtraArgs(c),
			ExtraVolumes: []kubeadmv1beta2.HostPathMount{kubernetesVolume},
		},
		Scheduler: kubeadmv1beta2.ControlPlaneComponent{
			ExtraArgs:    p.getSchedulerExtraArgs(c),
			ExtraVolumes: []kubeadmv1beta2.HostPathMount{kubernetesVolume},
		},
		DNS: kubeadmv1beta2.DNS{
			Type: kubeadmv1beta2.CoreDNS,
		},
		ImageRepository: p.config.Registry.Prefix,
		ClusterName:     c.Name,
	}

	utilruntime.Must(json.Merge(&config.Etcd, &c.Spec.Etcd))

	return config
}

func (p *Provider) getKubeProxyConfiguration(c *provider.Cluster) *kubeproxyv1alpha1.KubeProxyConfiguration {
	kubeProxyMode := "iptables"
	if c.Spec.Features.IPVS != nil && *c.Spec.Features.IPVS {
		kubeProxyMode = "ipvs"
	}

	return &kubeproxyv1alpha1.KubeProxyConfiguration{
		Mode: kubeproxyv1alpha1.ProxyMode(kubeProxyMode),
	}
}

func (p *Provider) getKubeletConfiguration(c *provider.Cluster) *kubeletv1beta1.KubeletConfiguration {
	return &kubeletv1beta1.KubeletConfiguration{
		KubeReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
		SystemReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
	}
}

func (p *Provider) getAPIServerExtraArgs(c *provider.Cluster) map[string]string {
	args := map[string]string{
		"token-auth-file": constants.TokenFile,
	}

	for k, v := range c.Spec.APIServerExtraArgs {
		args[k] = v
	}

	return args
}

func (p *Provider) getControllerManagerExtraArgs(c *provider.Cluster) map[string]string {
	args := map[string]string{}

	if len(c.Spec.ClusterCIDR) > 0 {
		args["allocate-node-cidrs"] = "true"
		args["cluster-cidr"] = c.Spec.ClusterCIDR
		args["node-cidr-mask-size"] = fmt.Sprintf("%v", c.Status.NodeCIDRMaskSize)
	}

	for k, v := range c.Spec.ControllerManagerExtraArgs {
		args[k] = v
	}

	return args
}

func (p *Provider) getSchedulerExtraArgs(c *provider.Cluster) map[string]string {
	args := map[string]string{}

	// args["use-legacy-policy-config"] = "true"
	// args["policy-config-file"] = constants.SchedulerPolicyConfigFile

	for k, v := range c.Spec.SchedulerExtraArgs {
		args[k] = v
	}

	return args
}
