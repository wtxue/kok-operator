package kubeconfig

import (
	"bytes"

	"github.com/wtxue/kube-on-kube-operator/pkg/util/ssh"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
)

type Option struct {
	MasterEndpoint string
	ClusterName    string
	CACert         []byte
	Token          string
}

// Install creates all the requested kubeconfig files.
func Install(s ssh.Interface, option *Option) error {
	config := CreateWithToken(option.MasterEndpoint, option.ClusterName, "kubernetes-admin", option.CACert, option.Token)
	data, err := runtime.Encode(clientcmdlatest.Codec, config)
	if err != nil {
		return err
	}
	err = s.WriteFile(bytes.NewReader(data), "/root/.kube/config") // fixme ssh not support $HOME or ~
	if err != nil {
		return err
	}

	return nil
}
