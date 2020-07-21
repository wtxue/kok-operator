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

package cluster

import (
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/server/mux"

	clusterprovider "github.com/wtxue/kube-on-kube-operator/pkg/provider/cluster"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/common"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/validation"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/config"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/pointer"
	"k8s.io/klog"
)

func Add(mgr *clusterprovider.CpManager, cfg *config.Config) error {
	p, err := NewProvider(mgr, cfg)
	if err != nil {
		klog.Errorf("init cluster provider error: %s", err)
		return err
	}
	mgr.Register(p.Name(), p)
	return nil
}

type Provider struct {
	*clusterprovider.DelegateProvider
	Mgr *clusterprovider.CpManager
	Cfg *config.Config
}

var _ clusterprovider.Provider = &Provider{}

func NewProvider(mgr *clusterprovider.CpManager, cfg *config.Config) (*Provider, error) {
	p := &Provider{
		Mgr: mgr,
		Cfg: cfg,
	}

	p.DelegateProvider = &clusterprovider.DelegateProvider{
		ProviderName: "Hosted",
		CreateHandlers: []clusterprovider.Handler{
			p.EnsureCopyFiles,
			p.EnsurePreInstallHook,
			p.EnsureClusterComplete,
			p.EnsureEtcd,
			p.EnsureCerts,
			p.EnsureKubeconfig,
			p.EnsureKubeMaster,

			p.EnsurePostInstallHook,
		},
		UpdateHandlers: []clusterprovider.Handler{
			p.EnsureKubeMaster,
			p.EnsureExtKubeconfig,
			p.EnsureAddons,
			p.EnsureFlannel,
			p.EnsureMetricsServer,
		},
	}

	return p, nil
}

func (p *Provider) RegisterHandler(mux *mux.PathRecorderMux) {
	prefix := "/provider/" + strings.ToLower(p.Name())

	mux.HandleFunc(path.Join(prefix, "ping"), p.ping)
}

func (p *Provider) Validate(cluster *common.Cluster) field.ErrorList {
	return validation.ValidateCluster(cluster)
}

func (p *Provider) PreCreate(cluster *common.Cluster) error {
	if cluster.Spec.Version == "" {
		cluster.Spec.Version = constants.K8sVersions[0]
	}
	if cluster.Spec.ClusterCIDR == "" {
		cluster.Spec.ClusterCIDR = "10.244.0.0/16"
	}
	if cluster.Spec.NetworkDevice == "" {
		cluster.Spec.NetworkDevice = "eth0"
	}

	if cluster.Spec.Features.IPVS == nil {
		cluster.Spec.Features.IPVS = pointer.ToBool(true)
	}

	if cluster.Spec.Properties.MaxClusterServiceNum == nil && cluster.Spec.ServiceCIDR == nil {
		cluster.Spec.Properties.MaxClusterServiceNum = pointer.ToInt32(256)
	}
	if cluster.Spec.Properties.MaxNodePodNum == nil {
		cluster.Spec.Properties.MaxNodePodNum = pointer.ToInt32(256)
	}
	if cluster.Spec.Features.SkipConditions == nil {
		cluster.Spec.Features.SkipConditions = p.Cfg.Feature.SkipConditions
	}

	if cluster.Spec.Etcd == nil {
		cluster.Spec.Etcd = &devopsv1.Etcd{Local: &devopsv1.LocalEtcd{}}
	}

	return nil
}
