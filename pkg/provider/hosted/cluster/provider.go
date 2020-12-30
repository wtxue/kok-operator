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
		ProviderName: "hosted",
		CreateHandlers: []clusterprovider.Handler{
			p.EnsureCopyFiles,
			p.EnsurePreInstallHook,
			p.EnsureClusterComplete,
			p.EnsureEtcd,
			p.EnsureCerts,
			p.EnsureKubeMisc,
			p.EnsureKubeMaster,

			p.EnsureExtKubeconfig,
			p.EnsurePostInstallHook,
		},
		UpdateHandlers: []clusterprovider.Handler{
			p.EnsureExtKubeconfig,
			p.EnsureKubeMaster,
			p.EnsureAddons,
			p.EnsureCni,
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
