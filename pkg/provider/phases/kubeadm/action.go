package kubeadm

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	kubeadmv1beta2 "github.com/wtxue/kok-operator/pkg/apis/kubeadm/v1beta2"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"github.com/wtxue/kok-operator/pkg/provider/phases/certs"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	"github.com/wtxue/kok-operator/pkg/util/template"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
)

const (
	joinControlPlaneCmd = `kubeadm join {{.ControlPlaneEndpoint}} \
--apiserver-advertise-address={{.AdvertiseAddress }} \
--node-name={{.NodeName}} \
--token={{.BootstrapToken}} \
--control-plane --certificate-key={{.CertificateKey}} \
--skip-phases=control-plane-join/mark-control-plane \
--discovery-token-unsafe-skip-ca-verification \
--ignore-preflight-errors=ImagePull \
--ignore-preflight-errors=Port-10250 \
--ignore-preflight-errors=NumCPU \
--ignore-preflight-errors=FileContent--proc-sys-net-bridge-bridge-nf-call-iptables \
--ignore-preflight-errors=DirAvailable--etc-kubernetes-manifests \
--ignore-preflight-errors=FileAvailable--etc-kubernetes-kubelet.conf \
-v 4
`
	joinNodeCmd = `kubeadm join {{.ControlPlaneEndpoint}} \
--node-name={{.NodeName}} \
--token={{.BootstrapToken}} \
--discovery-token-unsafe-skip-ca-verification \
--ignore-preflight-errors=ImagePull \
--ignore-preflight-errors=Port-10250 \
--ignore-preflight-errors=NumCPU \
--ignore-preflight-errors=FileContent--proc-sys-net-bridge-bridge-nf-call-iptables \
-v 4
`
)

// ImagesPull ...
func ImagesPull(ctx *common.ClusterContext, s ssh.Interface, k8sVersion, imagesRepository string) error {
	cmd := fmt.Sprintf("kubeadm config images pull --kubernetes-version=%s", k8sVersion)
	if imagesRepository != "" {
		cmd = cmd + fmt.Sprintf(" --image-repository=%s", imagesRepository)
	}

	ctx.Info("ImagesPull", "node", s.HostIP(), "cmd", cmd)
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		ctx.Error(err, "exit", exit, "node", s.HostIP())
		return errors.Wrapf(err, "node: %s exec: %q", s.HostIP(), cmd)
	}
	return nil
}

// Init phase ...
func Init(ctx *common.ClusterContext, s ssh.Interface, kubeadmConfig *Config, extraCmd string) error {
	configData, err := kubeadmConfig.Marshal()
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(configData), constants.KubeadmConfigFileName)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("kubeadm init phase %s --config=%s -v 4", extraCmd, constants.KubeadmConfigFileName)
	ctx.Info("kubeadm", "cmd", cmd)
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		ctx.Error(err, "node", s.HostIP(), "exit", exit)
		return errors.Wrapf(err, "node: %s exec: %q", s.HostIP(), cmd)
	}

	return nil
}

// InitCerts ...
func InitCerts(ctx *common.ClusterContext, cfg *Config, isManaged bool) error {
	var lastCACert *certs.CaAll
	cfgMaps := make(map[string][]byte)

	warp := &kubeadmv1beta2.WarpperConfiguration{
		InitConfiguration:    cfg.InitConfiguration,
		ClusterConfiguration: cfg.ClusterConfiguration,
		IPs:                  ctx.IPs(),
	}

	var certList certs.Certificates
	if !isManaged {
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

	if ctx.Credential.CertsBinaryData == nil {
		ctx.Credential.CertsBinaryData = make(map[string][]byte)
	}

	for pathFile, v := range cfgMaps {
		if pathFile == constants.CACertName {
			ctx.Credential.CACert = v
		}

		if pathFile == constants.CAKeyName {
			ctx.Credential.CAKey = v
		}

		if pathFile == constants.EtcdCACertName {
			ctx.Credential.ETCDCACert = v
		}

		if pathFile == constants.EtcdCAKeyName {
			ctx.Credential.ETCDCAKey = v
		}

		if pathFile == constants.APIServerEtcdClientCertName {
			ctx.Credential.ETCDAPIClientCert = v
		}

		if pathFile == constants.APIServerEtcdClientKeyName {
			ctx.Credential.ETCDAPIClientKey = v
		}

		ctx.Credential.CertsBinaryData[pathFile] = v
	}

	return nil
}

// JoinControlPlaneOption ...
type JoinControlPlaneOption struct {
	NodeName             string
	AdvertiseAddress     string
	BootstrapToken       string
	CertificateKey       string
	ControlPlaneEndpoint string
}

// JoinControlPlane ...
func JoinControlPlane(ctx *common.ClusterContext, s ssh.Interface) error {
	option := &JoinControlPlaneOption{
		BootstrapToken:       *ctx.Credential.BootstrapToken,
		CertificateKey:       *ctx.Credential.CertificateKey,
		ControlPlaneEndpoint: fmt.Sprintf("%s:6443", ctx.Cluster.Spec.Machines[0].IP),
		NodeName:             s.HostIP(),
		AdvertiseAddress:     s.HostIP(),
	}

	cmd, err := template.ParseString(joinControlPlaneCmd, option)
	if err != nil {
		return errors.Wrap(err, "template parse joinControlePlaneCmd")
	}
	ctx.Info("Join ControlPlane", "node", option.NodeName, "cmd", cmd)
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
	// cmd := fmt.Sprintf("docker rm -f $(docker ps -q -f '%s')", filter)
	// _, err := s.CombinedOutput(cmd)
	// if err != nil {
	// 	return err
	// }
	//
	// err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
	// 	cmd = fmt.Sprintf("docker ps -q -f '%s'", filter)
	// 	output, err := s.CombinedOutput(cmd)
	// 	if err != nil {
	// 		return false, nil
	// 	}
	// 	if len(output) == 0 {
	// 		return false, nil
	// 	}
	// 	return true, nil
	// })
	// if err != nil {
	// 	return fmt.Errorf("restart container(%s) error: %w", filter, err)
	// }

	return nil
}

type Option struct {
	HostIP           string
	Images           string
	EtcdPeerCluster  string
	TokenClusterName string
}

func BuildMasterEtcdPeerCluster(ctx *common.ClusterContext) string {
	etcdPeerEndpoints := []string{}

	for _, machine := range ctx.Cluster.Spec.Machines {
		etcdPeerEndpoints = append(etcdPeerEndpoints, fmt.Sprintf("%s=https://%s:2380", machine.IP, machine.IP))
	}

	return strings.Join(etcdPeerEndpoints, ",")
}

func ApplyCustomComponent(ctx *common.ClusterContext, s ssh.Interface, image string, podManifest string) error {
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
		return errors.Wrapf(err, "node: %s ReadFile: %s failed", s.HostIP(), podManifest)
	}

	obj, err := k8sutil.UnmarshalFromYaml(podBytes, corev1.SchemeGroupVersion)
	if err != nil {
		return errors.Wrapf(err, "node: %s marshalling %s failed", s.HostIP(), podManifest)
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

func RebuildMasterManifestFile(ctx *common.ClusterContext, s ssh.Interface, cfg *config.Config) error {
	if ctx.Credential.ManifestsData == nil {
		ctx.Credential.ManifestsData = make(map[string]string)
	}

	manifestFileList := []string{
		constants.EtcdPodManifestFile,
		constants.KubeAPIServerPodManifestFile,
		constants.KubeControllerManagerPodManifestFile,
		constants.KubeSchedulerPodManifestFile,
	}

	images := cfg.KubeAllImageFullName(constants.KubernetesAllImageName, ctx.Cluster.Spec.Version)
	for _, name := range manifestFileList {
		err := ApplyCustomComponent(ctx, s, images, name)
		if err != nil {
			return errors.Wrap(err, "ApplyCustomComponent")
		}
	}

	return nil
}
