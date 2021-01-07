package component

import (
	"fmt"
	"strings"

	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

var logger = log.Log.WithName("component")

func Install(ctx *common.ClusterContext, s ssh.Interface) error {
	// dir := "bin/linux/" // local debug config dir
	k8sDir := fmt.Sprintf("/k8s-%s/bin/", ctx.Cluster.Spec.Version)
	otherDir := "/k8s/bin/"
	if dir := constants.GetAnnotationKey(ctx.Cluster.Annotations, constants.ClusterAnnoLocalDebugDir); len(dir) > 0 {
		k8sDir = dir + k8sDir
		otherDir = dir + otherDir
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
			logger.Info("file exist ignoring", "node", s.HostIP(), "dst", ls.Dst)
			continue
		}

		err := s.CopyFile(ls.Src, ls.Dst)
		if err != nil {
			logger.Error(err, "CopyFile", "node", s.HostIP(), "src", ls.Src)
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

	logger.Info("write kubelet systemd unit file", "node", s.HostIP(), "dst", constants.KubeletSystemdUnitFilePath)
	err := s.WriteFile(strings.NewReader(kubeletService), constants.KubeletSystemdUnitFilePath)
	if err != nil {
		return err
	}

	logger.Info("write kubelet systemd service run config", "node", s.HostIP(), "path", constants.KubeletServiceRunConfig)
	err = s.WriteFile(strings.NewReader(KubeletServiceRunConfig), constants.KubeletServiceRunConfig)
	if err != nil {
		return err
	}

	cmd := "mkdir -p /etc/kubernetes/manifests/ && systemctl enable kubelet && systemctl daemon-reload && systemctl restart kubelet"
	if _, stderr, exit, err := s.Execf(cmd); err != nil || exit != 0 {
		cmd = "journalctl --unit kubelet -n10 --no-pager"
		jStdout, _, jExit, jErr := s.Execf(cmd)
		if jErr != nil || jExit != 0 {
			return fmt.Errorf("exec %q:error %s", cmd, err)
		}
		klog.Infof("log:\n%s", jStdout)

		return fmt.Errorf("Exec %s failed:exit %d:stderr %s:error %s:log:\n%s", cmd, exit, stderr, err, jStdout)
	}
	logger.Info("exec successfully", "node", s.HostIP(), "cmd", cmd)

	cmd = fmt.Sprintf("kubectl completion bash > /etc/bash_completion.d/kubectl")
	_, err = s.CombinedOutput(cmd)
	if err != nil {
		return err
	}

	logger.Info("exec successfully", "node", s.HostIP(), "cmd", cmd)
	return nil
}

func InstallCRI(ctx *common.ClusterContext, s ssh.Interface) error {
	// dir := "bin/linux/" // local debug config dir
	otherDir := "/k8s/bin/"
	if dir := constants.GetAnnotationKey(ctx.Cluster.Annotations, constants.ClusterAnnoLocalDebugDir); len(dir) > 0 {
		otherDir = dir + otherDir
	}

	var CopyList = []devopsv1.File{
		{
			Src: otherDir + "containerd.tar.gz",
			Dst: "/opt/k8s/containerd.tar.gz",
		},
	}

	for _, ls := range CopyList {
		if ok, err := s.Exist(ls.Dst); err == nil && ok {
			logger.Info("file exist ignoring", "node", s.HostIP(), "dst", ls.Dst)
			continue
		}

		err := s.CopyFile(ls.Src, ls.Dst)
		if err != nil {
			logger.Error(err, "CopyFile", "node", s.HostIP(), "src", ls.Src)
			return err
		}

		if strings.Contains(ls.Dst, "containerd") {
			cmd := "mkdir -p /usr/local/bin /usr/local/sbin /etc/systemd/system /opt/k8s/containerd && " +
				"tar -C /opt/k8s/containerd -xzf /opt/k8s/containerd.tar.gz && " +
				"cp -rf /opt/k8s/containerd/usr/local/sbin/* /usr/local/sbin/ && " +
				"cp -rf /opt/k8s/containerd/usr/local/bin/* /usr/local/bin/ && " +
				"cp -rf /opt/k8s/containerd/etc/crictl.yaml /etc/ && " +
				"cp -rf /opt/k8s/containerd/etc/systemd/system/containerd.service /etc/systemd/system/ && " +
				"rm -f /usr/local/bin/critest &&" +
				"rm -rf /opt/k8s/containerd"

			_, err := s.CombinedOutput(cmd)
			if err != nil {
				klog.Errorf("node: %s exec cmd %s err: %v", s.HostIP(), cmd, err)
				return err
			}
		}
		logger.Info("copy successfully", "node", s.HostIP(), "path", ls.Dst)
	}

	// mkdir /etc/containerd && containerd config default > /etc/containerd/config.toml
	// systemctl enable containerd && systemctl daemon-reload && systemctl restart containerd
	cmd := "systemctl enable containerd && systemctl daemon-reload && systemctl restart containerd"
	if _, stderr, exit, err := s.Execf(cmd); err != nil || exit != 0 {
		cmd = "journalctl --unit containerd -n10 --no-pager"
		jStdout, _, jExit, jErr := s.Execf(cmd)
		if jErr != nil || jExit != 0 {
			return fmt.Errorf("exec %q:error %s", cmd, err)
		}
		klog.Infof("log:\n%s", jStdout)

		return fmt.Errorf("Exec %s failed:exit %d:stderr %s:error %s:log:\n%s", cmd, exit, stderr, err, jStdout)
	}

	logger.Info("exec successfully", "node", s.HostIP(), "cmd", cmd)
	return nil
}
