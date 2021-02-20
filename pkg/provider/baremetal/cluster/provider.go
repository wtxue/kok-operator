package cluster

import (
	"path"
	"strings"

	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/provider/baremetal"
	"github.com/wtxue/kok-operator/pkg/provider/baremetal/validation"
	clusterprovider "github.com/wtxue/kok-operator/pkg/provider/cluster"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"github.com/wtxue/kok-operator/pkg/util/pointer"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/server/mux"
)

func Add(mgr *clusterprovider.CpManager, cfg *config.Config) error {
	p, err := NewProvider(mgr, cfg)
	if err != nil {
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
		ProviderName: baremetal.ProviderName,
		CreateHandlers: []clusterprovider.Handler{
			p.EnsureCopyFiles,
			p.EnsurePreInstallHook,
			p.EnsureRegistryHosts,
			p.EnsureNvidiaDriver,
			p.EnsureNvidiaContainerRuntime,
			p.EnsureSystem,
			p.EnsureEth,
			p.EnsureCRI,
			p.EnsureK8sComponent,
			p.EnsurePreflight,
			p.EnsureClusterComplete,

			p.EnsureImagesPull,
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

	if ctx.Cluster.Spec.CRIType == "" {
		ctx.Cluster.Spec.CRIType = devopsv1.ContainerdCRI
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
