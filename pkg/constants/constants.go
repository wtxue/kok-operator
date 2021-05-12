package constants

import (
	"fmt"
	"strings"
	"time"
)

const (
	// KubernetesDir is the directory Kubernetes owns for storing various configuration files
	KubernetesDir         = "/etc/kubernetes/"
	KubeletPodManifestDir = KubernetesDir + "manifests/"

	SchedulerPolicyConfigFile = KubernetesDir + "scheduler-policy-config.json"
	AuditWebhookConfigFile    = KubernetesDir + "audit-api-client-config.yaml"
	AuditPolicyConfigFile     = KubernetesDir + "audit-policy.yaml"

	EtcdPodManifestFile                  = KubeletPodManifestDir + "etcd.yaml"
	KubeAPIServerPodManifestFile         = KubeletPodManifestDir + "kube-apiserver.yaml"
	KubeControllerManagerPodManifestFile = KubeletPodManifestDir + "kube-controller-manager.yaml"
	KubeSchedulerPodManifestFile         = KubeletPodManifestDir + "kube-scheduler.yaml"

	DstTmpDir  = "/tmp/k8s/"
	DstBinDir  = "/usr/local/bin/"
	CNIBinDir  = "/opt/cni/bin/"
	CNIDataDir = "/var/lib/cni/"
	CNIConfDIr = "/etc/cni"

	CertificatesDir = KubernetesDir + "pki/"
	EtcdDataDir     = "/var/lib/etcd"

	TokenFile = KubernetesDir + "known_tokens.csv"

	KubectlConfigFile    = "/root/.kube/config"
	CACertAndKeyBaseName = "ca"

	// CACertName defines certificate name
	CACertName = CertificatesDir + "ca.crt"
	// CAKeyName defines certificate name
	CAKeyName = CertificatesDir + "ca.key"
	// APIServerCertName defines API's server certificate name
	APIServerCertName = CertificatesDir + "apiserver.crt"
	// APIServerKeyName defines API's server key name
	APIServerKeyName = CertificatesDir + "apiserver.key"
	// KubeletClientCurrent defines kubelet rotate certificates
	KubeletClientCurrent = "/var/lib/kubelet/pki/kubelet-client-current.pem"
	// EtcdCACertName defines etcd's CA certificate name
	EtcdCACertName = CertificatesDir + "etcd/ca.crt"
	// EtcdCAKeyName defines etcd's CA key name
	EtcdCAKeyName = CertificatesDir + "etcd/ca.key"
	// EtcdListenClientPort defines the port etcd listen on for client traffic
	EtcdListenClientPort = 2379
	// EtcdListenPeerPort defines the port etcd listen on for peer traffic
	EtcdListenPeerPort = 2380
	// APIServerEtcdClientCertName defines apiserver's etcd client certificate name
	APIServerEtcdClientCertName = CertificatesDir + "apiserver-etcd-client.crt"
	// APIServerEtcdClientKeyName defines apiserver's etcd client key name
	APIServerEtcdClientKeyName = CertificatesDir + "apiserver-etcd-client.key"

	// KubeletKubeConfigFileName defines the file name for the kubeconfig that the control-plane kubelet will use for talking
	// to the API server
	KubeletKubeConfigFileName    = KubernetesDir + "kubelet.conf"
	KubeletRunDirectory          = "/var/lib/kubelet/"
	DefaultSystemdUnitFilePath   = "/usr/lib/systemd/system/"
	KubeletSystemdUnitFilePath   = DefaultSystemdUnitFilePath + "kubelet.service"
	KubeletServiceRunConfigPath  = DefaultSystemdUnitFilePath + "kubelet.service.d/10-kubeadm.conf"
	KubeletConfigurationFileName = KubeletRunDirectory + "config.yaml"
	KubeletEnvFileName           = KubeletRunDirectory + "kubeadm-flags.env"
	KubeletEnvFileVariableName   = "KUBELET_KUBEADM_ARGS"

	// LabelNodeRoleMaster specifies that a node is a control-plane
	// This is a duplicate definition of the constant in pkg/controller/service/service_controller.go
	LabelNodeRoleMaster = "node-role.kubernetes.io/master"

	DNSIPIndex = 10

	// RenewCertsTimeThreshold control how long time left to renew certs
	RenewCertsTimeThreshold = 30 * 24 * time.Hour

	FlannelDirFile    = KubernetesDir + "flannel.yaml"
	CustomDir         = "/opt/k8s/"
	SystemInitFile    = CustomDir + "init.sh"
	SystemInitCniFile = CustomDir + "initCni.sh"
	CniHostLocalFile  = CNIConfDIr + "/net.d/10-host-local.conf"
	CniLoopBack       = CNIConfDIr + "/net.d/99-loopback.conf"

	ProviderDir  = "provider/baremetal/"
	ManifestsDir = ProviderDir + "manifests/"
)

const (
	// PauseVersion indicates the default pause image version for kubeadm
	PauseVersion = "3.5"

	// CoreDNSConfigMap specifies in what ConfigMap in the kube-system namespace the CoreDNS config should be stored
	CoreDNSConfigMap = "coredns"

	// CoreDNSDeploymentName specifies the name of the Deployment for CoreDNS add-on
	CoreDNSDeploymentName = "coredns"

	// CoreDNSImageName specifies the name of the image for CoreDNS add-on
	CoreDNSImageName = "coredns"

	// CoreDNSVersion is the version of CoreDNS to be deployed if it is used
	CoreDNSVersion = "1.7.0"

	KubeProxyImageName = "kube-proxy"

	// KubeProxyConfigMap specifies in what ConfigMap in the kube-system namespace the kube-proxy configuration should be stored
	KubeProxyConfigMap = "kube-proxy"

	// KubeProxyConfigMapKey specifies in what ConfigMap key the component config of kube-proxy should be stored
	KubeProxyConfigMapKey = "config.conf"

	// NodeBootstrapTokenAuthGroup specifies which group a Node Bootstrap Token should be authenticated in
	NodeBootstrapTokenAuthGroup = "system:bootstrappers:kubeadm:default-node-token"

	KubernetesAllImageName = "kubernetes"
)

// GetGenericImage generates and returns a platform agnostic image (backed by manifest list)
func GetGenericImage(prefix, image, tag string) string {
	if strings.HasPrefix(image, "kube") {
		if !strings.Contains(tag, "v") {
			tag = "v" + tag
		}
	}
	return fmt.Sprintf("%s/%s:%s", prefix, image, tag)
}

// GetKubeImage base centes images all kube
func GetKubeImage(prefix, image, tag string) string {
	if strings.HasPrefix(image, "kube") {
		if !strings.Contains(tag, "v") {
			tag = "v" + tag
		}
	}
	return fmt.Sprintf("%s/%s:%s", prefix, image, tag)
}
