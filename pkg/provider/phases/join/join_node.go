package join

import (
	"fmt"
	"os"

	"strings"

	"github.com/pkg/errors"
	kubeadmv1beta2 "github.com/wtxue/kok-operator/pkg/apis/kubeadm/v1beta2"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"github.com/wtxue/kok-operator/pkg/provider/phases/certs"
	"github.com/wtxue/kok-operator/pkg/provider/phases/kubeadm"
	"github.com/wtxue/kok-operator/pkg/util/pkiutil"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	"github.com/wtxue/kok-operator/pkg/util/template"
	"k8s.io/klog"
)

func ApplyPodManifest(hostIP string, ctx *common.ClusterContext, cfg *config.Config, pathName string, podManifest string, fileMaps map[string]string) error {
	opt := &kubeadm.Option{
		HostIP:           hostIP,
		Images:           cfg.KubeAllImageFullName(constants.KubernetesAllImageName, ctx.Cluster.Spec.Version),
		EtcdPeerCluster:  kubeadm.BuildMasterEtcdPeerCluster(ctx),
		TokenClusterName: ctx.Cluster.Name,
	}

	serialized, err := template.ParseString(podManifest, opt)
	if err != nil {
		return err
	}

	fileMaps[pathName] = string(serialized)
	return nil
}

func BuildKubeletKubeconfig(hostIP string, ctx *common.ClusterContext, apiserver string, fileMaps map[string]string) error {
	cfgMaps, err := certs.CreateKubeConfigFiles(ctx.Credential.CAKey, ctx.Credential.CACert,
		apiserver, hostIP, ctx.Cluster.Name, pkiutil.KubeletKubeConfigFileName)
	if err != nil {
		klog.Errorf("create node: %s kubelet kubeconfg err: %+v", hostIP, err)
		return err
	}

	var kubeletConf []byte
	for _, v := range cfgMaps {
		data, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			klog.Errorf("covert node: %s kubelet kubeconfg err: %+v", hostIP, err)
			return err
		}

		kubeletConf = data
		break
	}

	if kubeletConf == nil {
		return fmt.Errorf("node: %s can't build kubeletConf", hostIP)
	}

	fileMaps[constants.KubeletKubeConfigFileName] = string(kubeletConf)
	return nil
}

func JoinMasterNode(hostIP string, ctx *common.ClusterContext, cfg *config.Config, isMaster bool, fileMaps map[string]string) error {
	if !isMaster {
		fileMaps[constants.CACertName] = string(ctx.Credential.CACert)
		return nil
	}

	for pathName, va := range ctx.Credential.CertsBinaryData {
		fileMaps[pathName] = string(va)
	}

	for pathName, va := range ctx.Credential.KubeData {
		fileMaps[pathName] = va
	}

	for pathName, va := range ctx.Credential.ManifestsData {
		ApplyPodManifest(hostIP, ctx, cfg, pathName, va, fileMaps)
	}

	return nil
}

func JoinNodePhase(s ssh.Interface, cfg *config.Config, ctx *common.ClusterContext, apiserver string, isMaster bool) error {
	hostIP := s.HostIP()
	fileMaps := make(map[string]string)
	err := JoinMasterNode(hostIP, ctx, cfg, isMaster, fileMaps)
	if err != nil {
		return errors.Wrapf(err, "node: %s failed build misc file", hostIP)
	}

	err = BuildKubeletKubeconfig(hostIP, ctx, apiserver, fileMaps)
	if err != nil {
		return errors.Wrapf(err, "node: %s failed build kubelet file", hostIP)
	}

	nodeOpt := &kubeadmv1beta2.NodeRegistrationOptions{
		Name: hostIP,
	}
	flagsEnv := BuildKubeletDynamicEnvFile(cfg.Registry.Prefix, nodeOpt)
	fileMaps[constants.KubeletEnvFileName] = flagsEnv

	kubeletCfg := kubeadm.GetFullKubeletConfiguration(ctx)
	cfgYaml, err := KubeletMarshal(kubeletCfg)
	if err != nil {
		return errors.Wrapf(err, "node: %s failed marshal kubelet file", hostIP)
	}

	fileMaps[constants.KubeletConfigurationFileName] = string(cfgYaml)
	fileMaps[constants.KubeletServiceRunConfig] = kubeletEnvironmentTemplate

	for pathName, va := range fileMaps {
		klog.V(4).Infof("node: %s start write [%s] ...", hostIP, pathName)
		err = s.WriteFile(strings.NewReader(va), pathName)
		if err != nil {
			return errors.Wrapf(err, "node: %s failed to write for %s ", hostIP, pathName)
		}
	}

	klog.Infof("node: %s restart kubelet ... ", hostIP)
	cmd := fmt.Sprintf("mkdir -p /etc/kubernetes/manifests && systemctl enable kubelet && systemctl daemon-reload && systemctl restart kubelet")
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		klog.Errorf("%q %+v", exit, err)
		return err
	}
	return nil
}
