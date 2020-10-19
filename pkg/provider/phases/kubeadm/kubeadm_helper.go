package kubeadm

import (
	"bytes"
	"reflect"

	"fmt"

	"github.com/wtxue/kok-operator/pkg/apis"
	kubeadmv1beta2 "github.com/wtxue/kok-operator/pkg/apis/kubeadm/v1beta2"
	kubeletv1beta1 "github.com/wtxue/kok-operator/pkg/apis/kubelet/config/v1beta1"
	kubeproxyv1alpha1 "github.com/wtxue/kok-operator/pkg/apis/kubeproxy/config/v1alpha1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"github.com/wtxue/kok-operator/pkg/util/json"
	"github.com/wtxue/kok-operator/pkg/util/k8sutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

type Config struct {
	InitConfiguration      *kubeadmv1beta2.InitConfiguration
	ClusterConfiguration   *kubeadmv1beta2.ClusterConfiguration
	JoinConfiguration      *kubeadmv1beta2.JoinConfiguration
	KubeletConfiguration   *kubeletv1beta1.KubeletConfiguration
	KubeProxyConfiguration *kubeproxyv1alpha1.KubeProxyConfiguration
}

func (c *Config) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	v := reflect.ValueOf(*c)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsNil() {
			continue
		}
		obj, ok := v.Field(i).Interface().(runtime.Object)
		if !ok {
			panic("no runtime.Object")
		}
		gvks, _, err := apis.GetScheme().ObjectKinds(obj)
		if err != nil {
			return nil, err
		}

		yamlData, err := apis.MarshalToYAML(obj, gvks[0].GroupVersion())
		if err != nil {
			return nil, err
		}
		buf.WriteString("---\n")
		buf.Write(yamlData)
	}

	return buf.Bytes(), nil
}

func GetKubeadmConfigByMaster0(c *common.Cluster, cfg *config.Config) *Config {
	controlPlaneEndpoint := fmt.Sprintf("%s:6443", c.Spec.Machines[0].IP)
	return GetKubeadmConfig(c, cfg, controlPlaneEndpoint)
}

func GetKubeadmConfig(c *common.Cluster, cfg *config.Config, controlPlaneEndpoint string) *Config {
	cc := new(Config)
	cc.InitConfiguration = GetInitConfiguration(c)
	cc.ClusterConfiguration = GetClusterConfiguration(c, cfg, controlPlaneEndpoint)
	cc.KubeProxyConfiguration = GetKubeProxyConfiguration(c)
	cc.KubeletConfiguration = GetKubeletConfiguration(c)

	return cc
}

func GetInitConfiguration(c *common.Cluster) *kubeadmv1beta2.InitConfiguration {
	token, _ := kubeadmv1beta2.NewBootstrapTokenString(*c.ClusterCredential.BootstrapToken)

	initCfg := &kubeadmv1beta2.InitConfiguration{
		BootstrapTokens: []kubeadmv1beta2.BootstrapToken{
			{
				Token:       token,
				Description: "dke kubeadm bootstrap token",
				TTL:         &metav1.Duration{Duration: 0},
			},
		},
		CertificateKey: *c.ClusterCredential.CertificateKey,
	}

	if len(c.Cluster.Spec.Machines) > 0 {
		initCfg.NodeRegistration = kubeadmv1beta2.NodeRegistrationOptions{
			Name: c.Spec.Machines[0].IP,
		}

		initCfg.LocalAPIEndpoint = kubeadmv1beta2.APIEndpoint{
			AdvertiseAddress: c.Spec.Machines[0].IP,
			BindPort:         6443,
		}
	}

	return initCfg
}

func GetClusterConfiguration(c *common.Cluster, cfg *config.Config, controlPlaneEndpoint string) *kubeadmv1beta2.ClusterConfiguration {
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

	kubeadmCfg := &kubeadmv1beta2.ClusterConfiguration{
		CertificatesDir: constants.CertificatesDir,
		Networking: kubeadmv1beta2.Networking{
			DNSDomain:     c.Spec.DNSDomain,
			ServiceSubnet: c.Cluster.Status.ServiceCIDR,
		},
		KubernetesVersion:    c.Spec.Version,
		ControlPlaneEndpoint: controlPlaneEndpoint,
		APIServer: kubeadmv1beta2.APIServer{
			ControlPlaneComponent: kubeadmv1beta2.ControlPlaneComponent{
				ExtraArgs:    GetAPIServerExtraArgs(c),
				ExtraVolumes: []kubeadmv1beta2.HostPathMount{kubernetesVolume, auditVolume},
			},
			CertSANs: k8sutil.GetAPIServerCertSANs(c.Cluster),
		},
		ControllerManager: kubeadmv1beta2.ControlPlaneComponent{
			ExtraArgs:    GetControllerManagerExtraArgs(c),
			ExtraVolumes: []kubeadmv1beta2.HostPathMount{kubernetesVolume},
		},
		Scheduler: kubeadmv1beta2.ControlPlaneComponent{
			ExtraArgs:    GetSchedulerExtraArgs(c),
			ExtraVolumes: []kubeadmv1beta2.HostPathMount{kubernetesVolume},
		},
		DNS: kubeadmv1beta2.DNS{
			Type: kubeadmv1beta2.CoreDNS,
		},
		ImageRepository: cfg.Registry.Prefix,
		ClusterName:     c.Name,
	}

	utilruntime.Must(json.Merge(&kubeadmCfg.Etcd, &c.Spec.Etcd))

	return kubeadmCfg
}

func GetKubeProxyConfiguration(c *common.Cluster) *kubeproxyv1alpha1.KubeProxyConfiguration {
	kubeProxyMode := "iptables"
	if c.Spec.Features.IPVS != nil && *c.Spec.Features.IPVS {
		kubeProxyMode = "ipvs"
	}

	return &kubeproxyv1alpha1.KubeProxyConfiguration{
		Mode: kubeproxyv1alpha1.ProxyMode(kubeProxyMode),
	}
}

func GetKubeletConfiguration(c *common.Cluster) *kubeletv1beta1.KubeletConfiguration {
	return &kubeletv1beta1.KubeletConfiguration{
		KubeReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
		SystemReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
		MaxPods: *c.Spec.Properties.MaxNodePodNum,
	}
}

func GetFullKubeletConfiguration(c *common.Cluster) *kubeletv1beta1.KubeletConfiguration {
	return &kubeletv1beta1.KubeletConfiguration{
		StaticPodPath: constants.KubeletPodManifestDir,
		Authentication: kubeletv1beta1.KubeletAuthentication{
			X509: kubeletv1beta1.KubeletX509Authentication{
				ClientCAFile: constants.CACertName,
			},
			Webhook: kubeletv1beta1.KubeletWebhookAuthentication{
				Enabled: k8sutil.BoolPointer(true),
			},
			Anonymous: kubeletv1beta1.KubeletAnonymousAuthentication{
				Enabled: k8sutil.BoolPointer(false),
			},
		},
		Authorization: kubeletv1beta1.KubeletAuthorization{
			Mode:    kubeletv1beta1.KubeletAuthorizationModeWebhook,
			Webhook: kubeletv1beta1.KubeletWebhookAuthorization{},
		},
		ClusterDNS:    []string{c.Cluster.Status.DNSIP},
		ClusterDomain: c.Cluster.Spec.DNSDomain,

		KubeReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
		SystemReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
		MaxPods: *c.Spec.Properties.MaxNodePodNum,
	}
}

func GetAPIServerExtraArgs(c *common.Cluster) map[string]string {
	args := map[string]string{
		"token-auth-file": constants.TokenFile,
	}

	for k, v := range c.Spec.APIServerExtraArgs {
		args[k] = v
	}

	return args
}

func GetControllerManagerExtraArgs(c *common.Cluster) map[string]string {
	args := map[string]string{}

	if len(c.Spec.ClusterCIDR) > 0 {
		args["allocate-node-cidrs"] = "true"
		args["cluster-cidr"] = c.Spec.ClusterCIDR
		args["node-cidr-mask-size"] = fmt.Sprintf("%v", c.Cluster.Status.NodeCIDRMaskSize)
	}

	for k, v := range c.Spec.ControllerManagerExtraArgs {
		args[k] = v
	}

	return args
}

func GetSchedulerExtraArgs(c *common.Cluster) map[string]string {
	args := map[string]string{}

	// args["use-legacy-policy-config"] = "true"
	// args["policy-config-file"] = constants.SchedulerPolicyConfigFile

	for k, v := range c.Spec.SchedulerExtraArgs {
		args[k] = v
	}

	return args
}
