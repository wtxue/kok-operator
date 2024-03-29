package kubeadm

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/imdario/mergo"
	"github.com/wtxue/kok-operator/pkg/apis"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"

	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubeproxyv1alpha1 "k8s.io/kube-proxy/config/v1alpha1"
	kubeletv1beta1 "k8s.io/kubelet/config/v1beta1"
	bootstraptokenv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/bootstraptoken/v1"
	kubeadmv1beta3 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	utilsnet "k8s.io/utils/net"
)

type Config struct {
	InitConfiguration      *kubeadmv1beta3.InitConfiguration
	ClusterConfiguration   *kubeadmv1beta3.ClusterConfiguration
	JoinConfiguration      *kubeadmv1beta3.JoinConfiguration
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

func GetKubeadmConfigByMaster0(ctx *common.ClusterContext, cfg *config.Config) *Config {
	controlPlaneEndpoint := fmt.Sprintf("%s:6443", ctx.Cluster.Spec.Machines[0].IP)
	return GetKubeadmConfig(ctx, cfg, controlPlaneEndpoint)
}

func GetKubeadmConfig(ctx *common.ClusterContext, cfg *config.Config, controlPlaneEndpoint string) *Config {
	return &Config{
		InitConfiguration:      GetInitConfiguration(ctx, cfg),
		ClusterConfiguration:   GetClusterConfiguration(ctx, cfg, controlPlaneEndpoint),
		KubeletConfiguration:   GetFullKubeletConfiguration(ctx),
		KubeProxyConfiguration: GetKubeProxyConfiguration(ctx),
	}
}

func GetInitConfiguration(ctx *common.ClusterContext, cfg *config.Config) *kubeadmv1beta3.InitConfiguration {
	token, _ := bootstraptokenv1.NewBootstrapTokenString(*ctx.Credential.BootstrapToken)

	initCfg := &kubeadmv1beta3.InitConfiguration{
		BootstrapTokens: []bootstraptokenv1.BootstrapToken{
			{
				Token:       token,
				Description: "kubeadm bootstrap token",
				TTL:         &metav1.Duration{Duration: 0},
			},
		},
		CertificateKey: *ctx.Credential.CertificateKey,
	}

	kubeletExtraArgs := map[string]string{}
	utilruntime.Must(mergo.Merge(&kubeletExtraArgs, ctx.Cluster.Spec.KubeletExtraArgs))
	utilruntime.Must(mergo.Merge(&kubeletExtraArgs, cfg.Kubelet.ExtraArgs))

	if len(ctx.Cluster.Spec.Machines) > 0 {
		initCfg.NodeRegistration = kubeadmv1beta3.NodeRegistrationOptions{
			Name: ctx.Cluster.Spec.Machines[0].IP,
		}

		initCfg.LocalAPIEndpoint = kubeadmv1beta3.APIEndpoint{
			AdvertiseAddress: ctx.Cluster.Spec.Machines[0].IP,
			BindPort:         6443,
		}
	}

	if ctx.Cluster.Spec.CRIType != devopsv1.DockerCRI {
		initCfg.NodeRegistration.CRISocket = "unix:///run/containerd/containerd.sock"
	}

	return initCfg
}

func GetClusterConfiguration(ctx *common.ClusterContext, cfg *config.Config, controlPlaneEndpoint string) *kubeadmv1beta3.ClusterConfiguration {
	ctx.Logger.Info("GetClusterConfiguration", "CustomRegistry", cfg.CustomRegistry)

	kubernetesVolume := kubeadmv1beta3.HostPathMount{
		Name:      "vol-dir-0",
		HostPath:  "/etc/kubernetes",
		MountPath: "/etc/kubernetes",
	}

	auditVolume := kubeadmv1beta3.HostPathMount{
		Name:      "audit-dir-0",
		HostPath:  "/var/log/kubernetes",
		MountPath: "/var/log/kubernetes",
		PathType:  corev1.HostPathDirectoryOrCreate,
	}

	kubeadmCfg := &kubeadmv1beta3.ClusterConfiguration{
		CertificatesDir: constants.CertificatesDir,
		Networking: kubeadmv1beta3.Networking{
			DNSDomain:     ctx.Cluster.Spec.DNSDomain,
			ServiceSubnet: ctx.Cluster.Status.ServiceCIDR,
		},
		KubernetesVersion:    ctx.Cluster.Spec.Version,
		ControlPlaneEndpoint: controlPlaneEndpoint,
		APIServer: kubeadmv1beta3.APIServer{
			ControlPlaneComponent: kubeadmv1beta3.ControlPlaneComponent{
				ExtraArgs:    GetAPIServerExtraArgs(ctx),
				ExtraVolumes: []kubeadmv1beta3.HostPathMount{kubernetesVolume, auditVolume},
			},
			CertSANs: k8sutil.GetAPIServerCertSANs(ctx.Cluster),
		},
		ControllerManager: kubeadmv1beta3.ControlPlaneComponent{
			ExtraArgs:    GetControllerManagerExtraArgs(ctx),
			ExtraVolumes: []kubeadmv1beta3.HostPathMount{kubernetesVolume},
		},
		Scheduler: kubeadmv1beta3.ControlPlaneComponent{
			ExtraArgs:    GetSchedulerExtraArgs(ctx),
			ExtraVolumes: []kubeadmv1beta3.HostPathMount{kubernetesVolume},
		},
		DNS: kubeadmv1beta3.DNS{
			// Type: kubeadmv1beta3.CoreDNS,
		},
		ImageRepository: cfg.CustomRegistry,
		ClusterName:     ctx.Cluster.Name,
	}

	return kubeadmCfg
}

func GetKubeProxyConfiguration(ctx *common.ClusterContext) *kubeproxyv1alpha1.KubeProxyConfiguration {
	c := &kubeproxyv1alpha1.KubeProxyConfiguration{}
	c.Mode = "iptables"
	if ctx.Cluster.Spec.Features.IPVS != nil && *ctx.Cluster.Spec.Features.IPVS {
		c.Mode = "ipvs"
		c.ClusterCIDR = ctx.Cluster.Spec.ClusterCIDR
		if ctx.Cluster.Spec.Features.HA != nil {
			if ctx.Cluster.Spec.Features.HA.KubeHA != nil {
				c.IPVS.ExcludeCIDRs = []string{fmt.Sprintf("%s/32", ctx.Cluster.Spec.Features.HA.KubeHA.VIP)}
			}
			if ctx.Cluster.Spec.Features.HA.ThirdPartyHA != nil {
				c.IPVS.ExcludeCIDRs = []string{fmt.Sprintf("%s/32", ctx.Cluster.Spec.Features.HA.ThirdPartyHA.VIP)}
			}
		}
	}

	if utilsnet.IsIPv6CIDRString(ctx.Cluster.Spec.ClusterCIDR) {
		c.BindAddress = "::"
	}
	return c
}

func GetFullKubeletConfiguration(ctx *common.ClusterContext) *kubeletv1beta1.KubeletConfiguration {
	containerLogMaxFiles := int32(5)

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

		ClusterDNS:    []string{ctx.Cluster.Status.DNSIP},
		ClusterDomain: ctx.Cluster.Spec.DNSDomain,

		CgroupDriver:         "systemd",
		ContainerLogMaxSize:  "100Mi",
		ContainerLogMaxFiles: &containerLogMaxFiles,

		KubeReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},

		SystemReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
		MaxPods: *ctx.Cluster.Spec.Properties.MaxNodePodNum,
	}
}

func GetAPIServerExtraArgs(ctx *common.ClusterContext) map[string]string {
	args := map[string]string{
		"token-auth-file": constants.TokenFile,
	}

	for k, v := range ctx.Cluster.Spec.APIServerExtraArgs {
		args[k] = v
	}

	return args
}

func GetControllerManagerExtraArgs(ctx *common.ClusterContext) map[string]string {
	args := map[string]string{}

	if len(ctx.Cluster.Spec.ClusterCIDR) > 0 {
		args["allocate-node-cidrs"] = "true"
		args["cluster-cidr"] = ctx.Cluster.Spec.ClusterCIDR
		args["node-cidr-mask-size"] = fmt.Sprintf("%v", ctx.Cluster.Status.NodeCIDRMaskSize)
	}

	for k, v := range ctx.Cluster.Spec.ControllerManagerExtraArgs {
		args[k] = v
	}

	return args
}

func GetSchedulerExtraArgs(ctx *common.ClusterContext) map[string]string {
	args := map[string]string{}

	// args["use-legacy-policy-config"] = "true"
	// args["policy-config-file"] = constants.SchedulerPolicyConfigFile

	for k, v := range ctx.Cluster.Spec.SchedulerExtraArgs {
		args[k] = v
	}

	return args
}
