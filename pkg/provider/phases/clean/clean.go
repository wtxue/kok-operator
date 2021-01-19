package clean

import (
	"os"

	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	"k8s.io/klog/v2"
)

func CleanNode(s ssh.Interface) error {
	cmd := "kubeadm reset -f && rm -rf /var/lib/etcd /var/lib/kubelet /var/lib/dockershim /var/run/kubernetes /var/lib/cni /etc/kubernetes /etc/cni /root/.kube /opt/k8s && ipvsadm --clear"
	klog.Infof("start exec node: %s cmd: %s", s.HostIP(), cmd)
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		klog.Errorf("cmd: %s exit: %q err: %+v", cmd, exit, err)
		return errors.Wrapf(err, "node: %s exec: \n%s", s.HostIP(), cmd)
	}

	return nil
}
