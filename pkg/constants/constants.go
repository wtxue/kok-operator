package constants

import (
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
	KeepavlivedManifestFile              = KubeletPodManifestDir + "keepalived.yaml"

	DstTmpDir  = "/tmp/k8s/"
	DstBinDir  = "/usr/bin/"
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

	KubeadmConfigFileName = KubernetesDir + "kubeadm-config.yaml"

	// KubeletKubeConfigFileName defines the file name for the kubeconfig that the control-plane kubelet will use for talking
	// to the API server
	KubeletKubeConfigFileName = KubernetesDir + "kubelet.conf"

	// LabelNodeRoleMaster specifies that a node is a control-plane
	// This is a duplicate definition of the constant in pkg/controller/service/service_controller.go
	LabelNodeRoleMaster = "node-role.kubernetes.io/master"

	DNSIPIndex = 10

	// RenewCertsTimeThreshold control how long time left to renew certs
	RenewCertsTimeThreshold = 30 * 24 * time.Hour

	FlannelDirFile   = KubernetesDir + "flannel.yaml"
	CustomDir        = "/opt/k8s/"
	SystemInitFile   = CustomDir + "init.sh"
	CniHostLocalFile = CNIConfDIr + "/net.d/10-host-local.conf"
	CniLoopBack      = CNIConfDIr + "/net.d/99-loopback.conf"
)
