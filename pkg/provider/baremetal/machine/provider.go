package machine

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/config"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/validation"
	machineprovider "github.com/wtxue/kube-on-kube-operator/pkg/provider/machine"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/containerregistry"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"k8s.io/klog"
)

func init() {
	p, err := NewProvider()
	if err != nil {
		klog.Errorf("init machine provider error: %s", err)
		return
	}
	machineprovider.Register(p.Name(), p)
}

type Provider struct {
	*machineprovider.DelegateProvider

	config *config.Config
}

func NewProvider() (*Provider, error) {
	p := new(Provider)

	cfg, err := config.NewDefaultConfig()
	if err != nil {
		return nil, err
	}
	p.config = cfg

	containerregistry.Init(cfg.Registry.Domain, cfg.Registry.Namespace)

	p.DelegateProvider = &machineprovider.DelegateProvider{
		ProviderName: "Baremetal",
		CreateHandlers: []machineprovider.Handler{
			p.EnsureCopyFiles,
			p.EnsurePreInstallHook,

			p.EnsureClean,
			p.EnsureRegistryHosts,

			p.EnsureSystem,
			p.EnsurePreflight, // wait basic setting done

			p.EnsureJoinNode,
			p.EnsureKubeconfig,
			p.EnsureMarkNode,
			p.EnsureNodeReady,

			p.EnsurePostInstallHook,
		},
		UpdateHandlers: []machineprovider.Handler{
			p.EnsurePostInstallHook,
		},
	}

	return p, nil
}

var _ machineprovider.Provider = &Provider{}

func (p *Provider) Validate(machine *devopsv1.Machine) field.ErrorList {
	return validation.ValidateMachine(machine)
}
