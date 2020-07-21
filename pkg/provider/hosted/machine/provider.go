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

package machine

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/validation"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/config"
	machineprovider "github.com/wtxue/kube-on-kube-operator/pkg/provider/machine"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"k8s.io/klog"
)

func Add(mgr *machineprovider.MpManager, cfg *config.Config) error {
	p, err := NewProvider(mgr, cfg)
	if err != nil {
		klog.Errorf("init cluster provider error: %s", err)
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
