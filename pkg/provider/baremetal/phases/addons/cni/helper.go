package cni

import (
	"bytes"

	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/ssh"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/template"
)

const (
	hostLocalTemplate = `
{
 "cniVersion": "{{ default "0.3.1" .CniVersion}}",
 "name": "dke-cni",
 "type": "bridge",
 "bridge": "cni0",
 "forceAddress": false,
 "ipMasq": true,
 "hairpinMode": true,
 "ipam": {
  "type": "host-local",
  "ranges": [
   [
    {
     "subnet": "{{ .SubnetCidr}}",
     "rangeStart": "{{ .StartIP }}",
     "rangeEnd": "{{ .EndIP }}",
     "gateway": "{{ .Gw }}"
    }
   ]
  ],
  "routes": [
   {
    "dst": "0.0.0.0/0"
   },
   {
    "dst": "{{ .RouterDst }}",
    "gw": "{{ .RouterGw }}"
   }
  ],
  "dataDir": "/opt/k8s/data/cni"
 }
}
`
	loopbackTemplate = `
{
 "cniVersion": "{{ default "0.3.1" .CniVersion}}",
 "name": "lo",
 "type": "loopback"
}
`
)

type Option struct {
	CniVersion string
	SubnetCidr string
	StartIP    string
	EndIP      string
	Gw         string
	RouterDst  string
	RouterGw   string
}

func Install(s ssh.Interface, c *provider.Cluster) error {
	opt := &Option{
		SubnetCidr: "10.49.255.0/24",
		StartIP:    "10.49.255.1",
		EndIP:      "10.49.255.40",
		Gw:         "10.49.255.254",
		RouterDst:  "10.27.248.0/24",
		RouterGw:   "10.28.252.241",
	}
	localByte, err := template.ParseString(hostLocalTemplate, opt)
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(localByte), constants.CniHostLocalFile)
	if err != nil {
		return err
	}

	loopByte, err := template.ParseString(loopbackTemplate, opt)
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(loopByte), constants.CniLoopBack)
	if err != nil {
		return err
	}
	return nil
}
