package clean

import (
	"os"

	"github.com/wtxue/kok-operator/pkg/util/ssh"

	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func CleanNode(s ssh.Interface) error {
	cmd := "kubeadm reset -f && rm -rf /var/lib/etcd /var/lib/kubelet /var/lib/dockershim /var/run/kubernetes /var/lib/cni /etc/kubernetes /etc/cni /root/.kube /opt/k8s && ipvsadm --clear"
	logf.Log.V(4).Info("start exec", "node", s.HostIP(), "cmd", cmd)
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		logf.Log.Error(err, "exec err", "node", s.HostIP(), "cmd", cmd, "exit", exit)
		return errors.Wrapf(err, "node: %s exec: \n%s", s.HostIP(), cmd)
	}

	return nil
}
