package kubeadm

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"

	kubeadmv1beta2 "github.com/wtxue/kok-operator/pkg/apis/kubeadm/v1beta2"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"

	"github.com/wtxue/kok-operator/pkg/k8sutil"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"github.com/wtxue/kok-operator/pkg/provider/phases/certs"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	"github.com/wtxue/kok-operator/pkg/util/template"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

const (
	joinControlPlaneCmd = `kubeadm join {{.ControlPlaneEndpoint}} \
--node-name={{.NodeName}} --token={{.BootstrapToken}} \
--control-plane --certificate-key={{.CertificateKey}} \
--skip-phases=control-plane-join/mark-control-plane \
--discovery-token-unsafe-skip-ca-verification \
--ignore-preflight-errors=ImagePull \
--ignore-preflight-errors=Port-10250 \
--ignore-preflight-errors=NumCPU \
--ignore-preflight-errors=FileContent--proc-sys-net-bridge-bridge-nf-call-iptables \
--ignore-preflight-errors=DirAvailable--etc-kubernetes-manifests \
--ignore-preflight-errors=FileAvailable--etc-kubernetes-kubelet.conf \
-v 9
`
	joinNodeCmd = `kubeadm join {{.ControlPlaneEndpoint}} \
--node-name={{.NodeName}} \
--token={{.BootstrapToken}} \
--discovery-token-unsafe-skip-ca-verification \
--ignore-preflight-errors=ImagePull \
--ignore-preflight-errors=Port-10250 \
--ignore-preflight-errors=NumCPU \
--ignore-preflight-errors=FileContent--proc-sys-net-bridge-bridge-nf-call-iptables \
-v 9
`
)

type InitOption struct {
	KubeadmConfigFileName string
	NodeName              string
	BootstrapToken        string
	CertificateKey        string

	ETCDImageTag         string
	CoreDNSImageTag      string
	KubernetesVersion    string
	ControlPlaneEndpoint string

	DNSDomain             string
	ServiceSubnet         string
	NodeCIDRMaskSize      int32
	ClusterCIDR           string
	ServiceClusterIPRange string
	CertSANs              []string

	APIServerExtraArgs         map[string]string
	ControllerManagerExtraArgs map[string]string
	SchedulerExtraArgs         map[string]string

	ImageRepository string
	ClusterName     string

	KubeProxyMode string
}

func Init(s ssh.Interface, kubeadmConfig *Config, extraCmd string) error {
	configData, err := kubeadmConfig.Marshal()
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(configData), constants.KubeadmConfigFileName)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("kubeadm init phase %s --config=%s -v 9", extraCmd, constants.KubeadmConfigFileName)
	klog.Infof("init cmd: %s", cmd)
	out, err := s.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("exec %q error: %w", cmd, err)
	}
	klog.Info(string(out))

	return nil
}

func InitCerts(cfg *Config, c *common.Cluster, isHosted bool) error {
	var lastCACert *certs.CaAll
	cfgMaps := make(map[string][]byte)

	warp := &kubeadmv1beta2.WarpperConfiguration{
		InitConfiguration:    cfg.InitConfiguration,
		ClusterConfiguration: cfg.ClusterConfiguration,
		IPs:                  c.IPs(),
	}

	var certList certs.Certificates
	if !isHosted {
		certList = certs.GetDefaultCertList()
	} else {
		certList = certs.GetCertsWithoutEtcd()
	}

	for _, cert := range certList {
		if cert.CAName == "" {
			ret, err := certs.CreateCACertAndKeyFiles(cert, warp, cfgMaps)
			if err != nil {
				return err
			}
			lastCACert = ret
		} else {
			if lastCACert == nil {
				return fmt.Errorf("not hold CertificateAuthority by create cert: %s", cert.Name)
			}
			err := certs.CreateCertAndKeyFilesWithCA(cert, lastCACert, warp, cfgMaps)
			if err != nil {
				return errors.Wrapf(err, "create cert: %s", cert.Name)
			}
		}
	}

	err := certs.CreateServiceAccountKeyAndPublicKeyFiles(cfg.ClusterConfiguration.CertificatesDir, x509.RSA, cfgMaps)
	if err != nil {
		return errors.Wrapf(err, "create sa public key")
	}

	if len(cfgMaps) == 0 {
		return fmt.Errorf("no cert build")
	}

	if c.ClusterCredential.CertsBinaryData == nil {
		c.ClusterCredential.CertsBinaryData = make(map[string][]byte)
	}

	for pathFile, v := range cfgMaps {
		if pathFile == constants.CACertName {
			c.ClusterCredential.CACert = v
		}

		if pathFile == constants.CAKeyName {
			c.ClusterCredential.CAKey = v
		}

		if pathFile == constants.EtcdCACertName {
			c.ClusterCredential.ETCDCACert = v
		}

		if pathFile == constants.EtcdCAKeyName {
			c.ClusterCredential.ETCDCAKey = v
		}

		if pathFile == constants.APIServerEtcdClientCertName {
			c.ClusterCredential.ETCDAPIClientCert = v
		}

		if pathFile == constants.APIServerEtcdClientKeyName {
			c.ClusterCredential.ETCDAPIClientKey = v
		}

		c.ClusterCredential.CertsBinaryData[pathFile] = v
	}

	return nil
}

type JoinControlPlaneOption struct {
	NodeName             string
	BootstrapToken       string
	CertificateKey       string
	ControlPlaneEndpoint string
}

func JoinControlPlane(s ssh.Interface, c *common.Cluster) error {
	option := &JoinControlPlaneOption{
		BootstrapToken:       *c.ClusterCredential.BootstrapToken,
		CertificateKey:       *c.ClusterCredential.CertificateKey,
		ControlPlaneEndpoint: fmt.Sprintf("%s:6443", c.Spec.Machines[0].IP),
		NodeName:             s.HostIP(),
	}

	cmd, err := template.ParseString(joinControlPlaneCmd, option)
	if err != nil {
		return errors.Wrap(err, "parse joinControlePlaneCmd error")
	}
	klog.Infof("node: %s join cmd: %s", option.NodeName, cmd)
	exit, err := s.ExecStream(string(cmd), os.Stdout, os.Stderr)
	if err != nil || exit != 0 {
		return fmt.Errorf("exec %q failed:exit %d error:%v", cmd, exit, err)
	}

	return nil
}

type JoinNodeOption struct {
	NodeName             string
	BootstrapToken       string
	ControlPlaneEndpoint string
}

func JoinNode(s ssh.Interface, option *JoinNodeOption) error {
	cmd, err := template.ParseString(joinNodeCmd, option)
	if err != nil {
		return errors.Wrap(err, "parse joinNodeCmd error")
	}
	exit, err := s.ExecStream(string(cmd), os.Stdout, os.Stderr)
	if err != nil || exit != 0 {
		_, _, _, _ = s.Exec("kubeadm reset -f")
		return fmt.Errorf("exec %q failed:exit %d error:%v", cmd, exit, err)
	}

	return nil
}

func RenewCerts(s ssh.Interface) error {
	err := fixKubeadmBug1753(s)
	if err != nil {
		return fmt.Errorf("fixKubeadmBug1753(https://github.com/kubernetes/kubeadm/issues/1753) error: %w", err)
	}

	cmd := fmt.Sprintf("kubeadm alpha certs renew all --config=%s", constants.KubeadmConfigFileName)
	_, err = s.CombinedOutput(cmd)
	if err != nil {
		return err
	}

	err = RestartControlPlane(s)
	if err != nil {
		return err
	}

	return nil
}

// https://github.com/kubernetes/kubeadm/issues/1753
func fixKubeadmBug1753(s ssh.Interface) error {
	needUpdate := false

	data, err := s.ReadFile(constants.KubeletKubeConfigFileName)
	if err != nil {
		return err
	}
	kubeletKubeconfig, err := clientcmd.Load(data)
	if err != nil {
		return err
	}
	for _, info := range kubeletKubeconfig.AuthInfos {
		if info.ClientKeyData == nil && info.ClientCertificateData == nil {
			continue
		}

		info.ClientKeyData = []byte{}
		info.ClientCertificateData = []byte{}
		info.ClientKey = constants.KubeletClientCurrent
		info.ClientCertificate = constants.KubeletClientCurrent

		needUpdate = true
	}

	if needUpdate {
		data, err := runtime.Encode(clientcmdlatest.Codec, kubeletKubeconfig)
		if err != nil {
			return err
		}
		err = s.WriteFile(bytes.NewReader(data), constants.KubeletKubeConfigFileName)
		if err != nil {
			return err
		}
	}

	return nil
}

func RestartControlPlane(s ssh.Interface) error {
	targets := []string{"kube-apiserver", "kube-controller-manager", "kube-scheduler"}
	for _, one := range targets {
		err := RestartContainerByFilter(s, DockerFilterForControlPlane(one))
		if err != nil {
			return err
		}
	}

	return nil
}

func DockerFilterForControlPlane(name string) string {
	return fmt.Sprintf("label=io.kubernetes.container.name=%s", name)
}

func RestartContainerByFilter(s ssh.Interface, filter string) error {
	cmd := fmt.Sprintf("docker rm -f $(docker ps -q -f '%s')", filter)
	klog.V(4).Infof("node: %s, cmd: %s", s.HostIP(), cmd)
	_, err := s.CombinedOutput(cmd)
	if err != nil {
		return err
	}

	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		cmd = fmt.Sprintf("docker ps -q -f '%s'", filter)
		klog.V(4).Infof("wait node: %s, cmd: %s", s.HostIP(), cmd)
		output, err := s.CombinedOutput(cmd)
		if err != nil {
			return false, nil
		}
		if len(output) == 0 {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("restart container(%s) error: %w", filter, err)
	}

	return nil
}

type Option struct {
	HostIP           string
	Images           string
	EtcdPeerCluster  string
	TokenClusterName string
}

func BuildMasterEtcdPeerCluster(c *common.Cluster) string {
	etcdPeerEndpoints := []string{}

	for _, machine := range c.Spec.Machines {
		etcdPeerEndpoints = append(etcdPeerEndpoints, fmt.Sprintf("%s=https://%s:2380", machine.IP, machine.IP))
	}

	return strings.Join(etcdPeerEndpoints, ",")
}

func ApplyCustomComponent(s ssh.Interface, c *common.Cluster, image string, podManifest string) error {
	// var err error
	// var podBytes []byte
	// if podManifest == constants.EtcdPodManifestFile {
	// 	opt := &Option{
	// 		HostIP:           s.HostIP(),
	// 		Images:           image,
	// 		EtcdPeerCluster:  BuildMasterEtcdPeerCluster(c),
	// 		TokenClusterName: c.Cluster.Name,
	// 	}
	//
	// 	podBytes, err = template.ParseString(staticPodEtcdTemplate, opt)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	c.ClusterCredential.ManifestsData[podManifest] = staticPodEtcdTemplate
	// } else {
	// 	podBytes, err = s.ReadFile(podManifest)
	// 	if err != nil {
	// 		return fmt.Errorf("node: %s ReadFile: %s failed error: %v", s.HostIP(), podManifest, err)
	// 	}
	//
	// 	replaceStr := strings.Replace(string(podBytes), s.HostIP(), "{{ .HostIP }}", -1)
	// 	c.ClusterCredential.ManifestsData[podManifest] = replaceStr
	// }

	podBytes, err := s.ReadFile(podManifest)
	if err != nil {
		return fmt.Errorf("node: %s ReadFile: %s failed error: %v", s.HostIP(), podManifest, err)
	}

	obj, err := k8sutil.UnmarshalFromYaml(podBytes, corev1.SchemeGroupVersion)
	if err != nil {
		return fmt.Errorf("node: %s marshalling %s failed error: %v", s.HostIP(), podManifest, err)
	}

	switch obj.(type) {
	case *corev1.Pod:
		ins := obj.(*corev1.Pod)
		if len(ins.Spec.Containers) > 0 {
			ins.Spec.Containers[0].Image = image
		}
	default:
		return fmt.Errorf("unknown type")
	}

	serialized, err := k8sutil.MarshalToYaml(obj, corev1.SchemeGroupVersion)
	if err != nil {
		return errors.Wrapf(err, "node: %s failed to marshal manifest for %s to YAML", s.HostIP(), podManifest)
	}

	err = s.WriteFile(bytes.NewReader(serialized), podManifest)
	if err != nil {
		return errors.Wrapf(err, "node: %s failed to write manifest for %s ", s.HostIP(), podManifest)
	}

	return nil
}

func RebuildMasterManifestFile(s ssh.Interface, c *common.Cluster, cfg *config.Config) error {
	if c.ClusterCredential.ManifestsData == nil {
		c.ClusterCredential.ManifestsData = make(map[string]string)
	}

	manifestFileList := []string{
		constants.EtcdPodManifestFile,
		constants.KubeAPIServerPodManifestFile,
		constants.KubeControllerManagerPodManifestFile,
		constants.KubeSchedulerPodManifestFile,
	}

	images := cfg.KubeAllImageFullName(constants.KubernetesAllImageName, c.Cluster.Spec.Version)
	for _, name := range manifestFileList {
		err := ApplyCustomComponent(s, c, images, name)
		if err != nil {
			klog.Errorf("applyCustomComponent %s err: %v", name, err)
			return err
		}
	}

	return nil
}
