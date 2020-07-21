/*
Copyright 2020 wtxue.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package option

import (
	"github.com/spf13/pflag"
)

type ApiServerOption struct {
	IsLocalKube bool

	// local /  external
	IsLocalEtcd bool

	BaseBinDir string

	RootDir string

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
		BaseBinDir:  "",
		RootDir:     "/k8s",
	}
}

func (o *ApiServerOption) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.IsLocalKube, "is-local-kube", o.IsLocalKube, "enable local mock kube api server")
	fs.BoolVar(&o.IsLocalEtcd, "is-local-etcd", o.IsLocalEtcd, "when enable local mock kube use loacl etcd cluster")
	fs.StringVar(&o.BaseBinDir, "baseBinDir", o.BaseBinDir, "the base bin dir")
	fs.StringVar(&o.RootDir, "rootDir", o.RootDir, "the root bin dir")
}
