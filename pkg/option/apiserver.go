package option

import (
	"github.com/spf13/pflag"
)

type ApiServerOption struct {
	IsLocalKube bool

	// local /  external
	IsLocalEtcd bool

	BaseBinDir string

	// DataDir is the directory etcd will place its data.
	// Defaults to "k8sdata/etcd".
	DataDir string `json:"dataDir" protobuf:"bytes,1,opt,name=dataDir"`

	// Endpoints of etcd members. Required for ExternalEtcd.
	Endpoints []string `json:"endpoints" protobuf:"bytes,1,rep,name=endpoints"`

	// CAFile is an SSL Certificate Authority file used to secure etcd communication.
	// Required if using a TLS connection.
	CAFile string `json:"caFile" protobuf:"bytes,2,opt,name=caFile"`

	// CertFile is an SSL certification file used to secure etcd communication.
	// Required if using a TLS connection.
	CertFile string `json:"certFile" protobuf:"bytes,3,opt,name=certFile"`

	// KeyFile is an SSL key file used to secure etcd communication.
	// Required if using a TLS connection.
	KeyFile string `json:"keyFile" protobuf:"bytes,4,opt,name=keyFile"`
}

func DefaultApiServerOption() *ApiServerOption {
	return &ApiServerOption{
		IsLocalKube: true,
		IsLocalEtcd: true,
		BaseBinDir:  "k8s/bin",
		DataDir:     "k8s/data/etcd",
	}
}

func (o *ApiServerOption) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.IsLocalKube, "is-local-kube", o.IsLocalKube, "enable local mock kube api server")
	fs.BoolVar(&o.IsLocalEtcd, "is-local-etcd", o.IsLocalEtcd, "when enable local mock kube use loacl etcd cluster")
}
