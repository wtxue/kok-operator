package hosts

import (
	"bytes"

	"github.com/wtxue/kube-on-kube-operator/pkg/util/ssh"
)

// RemoteHosts for remote hosts
type RemoteHosts struct {
	Host string
	SSH  ssh.Interface
}

// Data return hosts data
func (h *RemoteHosts) Data() ([]byte, error) {
	return h.SSH.ReadFile(linuxHostfile)
}

// Set sets hosts
func (h *RemoteHosts) Set(ip string) error {
	data, err := h.Data()
	if err != nil {
		return err
	}
	data, err = setHosts(data, h.Host, ip)
	if err != nil {
		return err
	}

	return h.SSH.WriteFile(bytes.NewReader(data), linuxHostfile)
}
