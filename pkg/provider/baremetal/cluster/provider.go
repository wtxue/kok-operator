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

	clusterprovider "github.com/wtxue/kok-operator/pkg/provider/cluster"

	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/provider/baremetal/validation"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"github.com/wtxue/kok-operator/pkg/util/pointer"
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
		ProviderName: "baremetal",
		CreateHandlers: []clusterprovider.Handler{
			p.EnsureCopyFiles,
			p.EnsurePreInstallHook,
			p.EnsureEth,
			p.EnsureSystem,
			p.EnsureCRI,
			p.EnsureComponent,
			p.EnsurePreflight,
			p.EnsureClusterComplete,

			p.EnsureCerts,
			p.EnsureKubeadmInitKubeletStartPhase, // start kubelet
			p.EnsureBuildLocalKubeconfig,
			p.EnsureKubeMiscPhase,
			p.EnsureKubeadmInitControlPlanePhase,
			p.EnsureKubeadmInitEtcdPhase,

			p.EnsureKubeadmInitWaitControlPlanePhase,
			p.EnsureKubeadmInitUploadConfigPhase,
			p.EnsureKubeadmInitUploadCertsPhase,
			p.EnsureKubeadmInitBootstrapTokenPhase,
			p.EnsureKubeadmInitAddonPhase,
			p.EnsureJoinControlePlane,
			p.EnsureMarkControlPlane,
			p.EnsureRebuildEtcd,

			p.EnsureDeployCni,
			p.EnsureRebuildControlPlane,
			p.EnsureExtKubeconfig,
			p.EnsurePostInstallHook,
		},
		UpdateHandlers: []clusterprovider.Handler{
			p.EnsureExtKubeconfig,
			p.EnsureMasterNode,
			p.EnsureDeployCni,
			p.EnsureRebuildEtcd,
			p.EnsureRebuildControlPlane,
			p.EnsureRenewCerts,
			p.EnsureAPIServerCert,
			p.EnsureMetricsServer,
		},
	}

	return p, nil
}

func (p *Provider) RegisterHandler(mux *mux.PathRecorderMux) {
	prefix := "/provider/" + strings.ToLower(p.Name())

	mux.HandleFunc(path.Join(prefix, "ping"), p.ping)
}

func (p *Provider) Validate(ctx *common.ClusterContext) field.ErrorList {
	return validation.ValidateCluster(ctx)
}

func (p *Provider) PreCreate(ctx *common.ClusterContext) error {
	if ctx.Cluster.Spec.Version == "" {
		ctx.Cluster.Spec.Version = constants.K8sVersions[0]
	}
	if ctx.Cluster.Spec.ClusterCIDR == "" {
		ctx.Cluster.Spec.ClusterCIDR = "10.244.0.0/16"
	}
	if ctx.Cluster.Spec.NetworkDevice == "" {
		ctx.Cluster.Spec.NetworkDevice = "eth0"
	}

	if ctx.Cluster.Spec.Features.IPVS == nil {
		ctx.Cluster.Spec.Features.IPVS = pointer.ToBool(true)
	}

	if ctx.Cluster.Spec.Properties.MaxClusterServiceNum == nil && ctx.Cluster.Spec.ServiceCIDR == nil {
		ctx.Cluster.Spec.Properties.MaxClusterServiceNum = pointer.ToInt32(256)
	}
	if ctx.Cluster.Spec.Properties.MaxNodePodNum == nil {
		ctx.Cluster.Spec.Properties.MaxNodePodNum = pointer.ToInt32(256)
	}
	if ctx.Cluster.Spec.Features.SkipConditions == nil {
		ctx.Cluster.Spec.Features.SkipConditions = p.Cfg.Feature.SkipConditions
	}

	if ctx.Cluster.Spec.Etcd == nil {
		ctx.Cluster.Spec.Etcd = &devopsv1.Etcd{Local: &devopsv1.LocalEtcd{}}
	}

	return nil
}
