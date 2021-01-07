package machine

import (
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/provider/baremetal/validation"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	machineprovider "github.com/wtxue/kok-operator/pkg/provider/machine"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func Add(mgr *machineprovider.MpManager, cfg *config.Config) error {
	p, err := NewProvider(mgr, cfg)
	if err != nil {
		return err
	}
	mgr.Register(p.Name(), p)
	return nil
}

type Provider struct {
	*machineprovider.DelegateProvider
	Mgr *machineprovider.MpManager
	Cfg *config.Config
}

func NewProvider(mgr *machineprovider.MpManager, cfg *config.Config) (*Provider, error) {
	p := &Provider{
		Mgr: mgr,
		Cfg: cfg,
	}

	p.DelegateProvider = &machineprovider.DelegateProvider{
		ProviderName: "Hosted",
		CreateHandlers: []machineprovider.Handler{
			p.EnsureCopyFiles,
			p.EnsurePreInstallHook,
			p.EnsureClean,
			p.EnsureRegistryHosts,

			p.EnsureEth,
			p.EnsureSystem,
			p.EnsureK8sComponent,
			p.EnsurePreflight, // wait basic setting done

			p.EnsureJoinNode,
			p.EnsureKubeconfig,
			p.EnsureMarkNode,
			p.EnsureCni,
			p.EnsureNodeReady,

			p.EnsurePostInstallHook,
		},
		UpdateHandlers: []machineprovider.Handler{
			p.EnsureCni,
			p.EnsurePostInstallHook,
			p.EnsureRegistryHosts,
		},
	}

	return p, nil
}

var _ machineprovider.Provider = &Provider{}

func (p *Provider) Validate(machine *devopsv1.Machine) field.ErrorList {
	return validation.ValidateMachine(machine)
}
