package kubeadm

import (
	"bytes"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"

	"crypto/x509"

	"path/filepath"

	"os"

	kubeadmv1beta2 "github.com/wtxue/kube-on-kube-operator/pkg/apis/kubeadm/v1beta2"
	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/phases/kubeadm/helper"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/ssh"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/template"
	"k8s.io/klog"
)

const (
	kubeadmConfigFile  = "kubeadm/kubeadm-config.yaml"
	kubeadmKubeletConf = "/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf"

	joinControlPlaneCmd = `kubeadm join {{.ControlPlaneEndpoint}} \
--node-name={{.NodeName}} --token={{.BootstrapToken}} \
--control-plane --certificate-key={{.CertificateKey}} \
--skip-phases=control-plane-join/mark-control-plane \
--discovery-token-unsafe-skip-ca-verification \
--ignore-preflight-errors=ImagePull \
--ignore-preflight-errors=Port-10250 \
--ignore-preflight-errors=FileContent--proc-sys-net-bridge-bridge-nf-call-iptables \
--ignore-preflight-errors=DirAvailable--etc-kubernetes-manifests \
--ignore-preflight-errors=FileAvailable--etc-kubernetes-kubelet.conf
`
	joinNodeCmd = `kubeadm join {{.ControlPlaneEndpoint}} \
--node-name={{.NodeName}} \
--token={{.BootstrapToken}} \
--discovery-token-unsafe-skip-ca-verification \
--ignore-preflight-errors=ImagePull \
--ignore-preflight-errors=Port-10250 \
--ignore-preflight-errors=FileContent--proc-sys-net-bridge-bridge-nf-call-iptables
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

	cmd := fmt.Sprintf("kubeadm init phase %s --config=%s", extraCmd, constants.KubeadmConfigFileName)
	klog.Infof("init cmd: %s", cmd)
	out, err := s.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("exec %q error: %w", cmd, err)
	}
	klog.Info(string(out))

	return nil
}

func InitCustomCerts(cfg *Config, c *provider.Cluster) error {
	var lastCACert *helper.CaAll
	cfgMaps := make(map[string][]byte)

	warp := &kubeadmv1beta2.WarpperConfiguration{
		InitConfiguration:    cfg.InitConfiguration,
		ClusterConfiguration: cfg.ClusterConfiguration,
		IPs:                  c.IPs(),
	}

	for _, cert := range helper.GetDefaultCertList() {
		if cert.CAName == "" {
			ret, err := helper.CreateCACertAndKeyFiles(cert, warp, cfgMaps)
			if err != nil {
				return err
			}
			lastCACert = ret
		} else {
			if lastCACert == nil {
				return fmt.Errorf("not hold CertificateAuthority by create cert: %s", cert.Name)
			}
			err := helper.CreateCertAndKeyFilesWithCA(cert, lastCACert, warp, cfgMaps)
			if err != nil {
				return errors.Wrapf(err, "create cert: %s", cert.Name)
			}
		}
	}

	err := helper.CreateServiceAccountKeyAndPublicKeyFiles(cfg.ClusterConfiguration.CertificatesDir, x509.RSA, cfgMaps)
	if err != nil {
		return errors.Wrapf(err, "create sa public key")
	}

	if len(cfgMaps) == 0 {
		return fmt.Errorf("no cert build")
	}

	for _, machine := range c.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		klog.Infof("node: %s start write cert ...", machine.IP)
		for pathFile, v := range cfgMaps {
			err = sh.WriteFile(bytes.NewReader(v), pathFile)
			if err != nil {
				klog.Errorf("write kubeconfg: %s err: %+v", pathFile, err)
				return err
			}
		}
	}
	return nil
}

func InitCustomKubeconfig(cfg *Config, s ssh.Interface, c *provider.Cluster) error {
	warp := &kubeadmv1beta2.WarpperConfiguration{
		InitConfiguration:    cfg.InitConfiguration,
		ClusterConfiguration: cfg.ClusterConfiguration,
		IPs:                  c.IPs(),
	}

	cfgMaps, err := helper.CreateKubeConfigFile(c.ClusterCredential.CAKey,
		c.ClusterCredential.CACert, &kubeadmv1beta2.APIEndpoint{
			AdvertiseAddress: s.HostIP(),
			BindPort:         warp.LocalAPIEndpoint.BindPort,
		}, warp.ClusterName)
	if err != nil {
		klog.Errorf("create kubeconfg err: %+v", err)
		return err
	}

	klog.Infof("node: %s start write kubeconfig ...", s.HostIP())
	for noPathFile, v := range cfgMaps {
		by, err := helper.BuildKubeConfigByte(v)
		if err != nil {
			return err
		}

		pathFile := filepath.Join(constants.KubernetesDir, noPathFile)
		err = s.WriteFile(bytes.NewReader(by), pathFile)
		if err != nil {
			klog.Errorf("write kubeconfg: %s err: %+v", noPathFile, err)
			return err
		}
	}

	return nil
}

type JoinControlPlaneOption struct {
	NodeName             string
	BootstrapToken       string
	CertificateKey       string
	ControlPlaneEndpoint string
}

func JoinControlPlane(s ssh.Interface, option *JoinControlPlaneOption) error {
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
