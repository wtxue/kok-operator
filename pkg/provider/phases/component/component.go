package component

import (
	"fmt"
	"strings"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/common"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/ssh"
	"k8s.io/klog"
)

const (
	kubeletService = `
[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=https://kubernetes.io/docs/

[Service]
User=root
ExecStart=/usr/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
`

	KubeletServiceRunConfig = `
[Service]
Environment="KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf"
Environment="KUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yaml"
EnvironmentFile=-/var/lib/kubelet/kubeadm-flags.env
EnvironmentFile=-/etc/sysconfig/kubelet
ExecStart=
ExecStart=/usr/bin/kubelet $KUBELET_KUBECONFIG_ARGS $KUBELET_CONFIG_ARGS $KUBELET_KUBEADM_ARGS $KUBELET_EXTRA_ARGS
`
)

func Install(s ssh.Interface, c *common.Cluster) error {
	// dir := "k8s/linuxbin/" // local debug config dir
	var k8sDir string
	var otherDir string
	if dir := constants.GetAnnotationKey(c.Cluster.Annotations, constants.ClusterAnnoLocalDebugDir); len(dir) > 0 {
		k8sDir = dir
		otherDir = dir
	} else {
		k8sDir = fmt.Sprintf("/k8s-%s/bin/", c.Cluster.Spec.Version)
		otherDir = "/k8s/bin/"
	}

	var CopyList = []devopsv1.File{
		{
			Src: k8sDir + "kubectl",
			Dst: "/usr/local/bin/kubectl",
		},
		{
			Src: k8sDir + "kubeadm",
			Dst: "/usr/local/bin/kubeadm",
		},
		{
			Src: k8sDir + "kubelet",
			Dst: "/usr/bin/kubelet",
		},
		{
			Src: otherDir + "cni.tgz",
			Dst: "/opt/cni.tgz",
		},
	}

	for _, ls := range CopyList {
		if ok, err := s.Exist(ls.Dst); err == nil && ok {
			continue
		}

		err := s.CopyFile(ls.Src, ls.Dst)
		if err != nil {
			klog.Errorf("node: %s copy %s err: %v", s.HostIP(), ls.Src, err)
			return err
		}

		if strings.Contains(ls.Dst, "bin") {
			_, _, _, err = s.Execf("chmod a+x %s", ls.Dst)
			if err != nil {
				return err
			}
		}

		if strings.Contains(ls.Dst, "cni") {
			cmd := fmt.Sprintf("mkdir -p %s && tar -C %s -xzf /opt/cni.tgz", constants.CNIBinDir, constants.CNIBinDir)
			_, err := s.CombinedOutput(cmd)
			if err != nil {
				klog.Errorf("node: %s exec cmd %s err: %v", s.HostIP(), cmd, err)
				return err
			}
		}
		klog.Errorf("node: %s copy %s success", s.HostIP(), ls.Dst)
	}

	klog.Infof("node: %s start write %s ... ", s.HostIP(), constants.KubeletSystemdUnitFilePath)
	err := s.WriteFile(strings.NewReader(kubeletService), constants.KubeletSystemdUnitFilePath)
	if err != nil {
		return err
	}

	klog.Infof("node: %s start write %s ... ", s.HostIP(), constants.KubeletServiceRunConfig)
	err = s.WriteFile(strings.NewReader(KubeletServiceRunConfig), constants.KubeletServiceRunConfig)
	if err != nil {
		return err
	}

	unitName := fmt.Sprintf("%s.service", "kubelet")
	cmd := fmt.Sprintf("mkdir -p /etc/kubernetes/manifests/ && systemctl -f enable %s && systemctl daemon-reload && systemctl restart %s", unitName, unitName)
	if _, stderr, exit, err := s.Execf(cmd); err != nil || exit != 0 {
		cmd = fmt.Sprintf("journalctl --unit %s -n10 --no-pager", unitName)
		jStdout, _, jExit, jErr := s.Execf(cmd)
		if jErr != nil || jExit != 0 {
			return fmt.Errorf("exec %q:error %s", cmd, err)
		}
		klog.Infof("log:\n%s", jStdout)

		return fmt.Errorf("Exec %s failed:exit %d:stderr %s:error %s:log:\n%s", cmd, exit, stderr, err, jStdout)
	}

	cmd = fmt.Sprintf("kubectl completion bash > /etc/bash_completion.d/kubectl")
	_, err = s.CombinedOutput(cmd)
	if err != nil {
		return err
	}

	return nil
}
