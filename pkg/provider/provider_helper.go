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

package provider

import (
	baremetalcluster "github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/cluster"
	baremetalmachine "github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/machine"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/cluster"
	clusterprovider "github.com/wtxue/kube-on-kube-operator/pkg/provider/cluster"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/config"
	hostedcluster "github.com/wtxue/kube-on-kube-operator/pkg/provider/hosted/cluster"
	hostedmachine "github.com/wtxue/kube-on-kube-operator/pkg/provider/hosted/machine"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/machine"
	machineprovider "github.com/wtxue/kube-on-kube-operator/pkg/provider/machine"
)

type ProviderManager struct {
	*cluster.CpManager
	*machine.MpManager
	Cfg *config.Config
}

var AddToCpManagerFuncs []func(*clusterprovider.CpManager, *config.Config) error
var AddToMpManagerFuncs []func(*machineprovider.MpManager, *config.Config) error

func NewProvider() (*ProviderManager, error) {
	AddToCpManagerFuncs = append(AddToCpManagerFuncs, baremetalcluster.Add)
	AddToCpManagerFuncs = append(AddToCpManagerFuncs, hostedcluster.Add)

	AddToMpManagerFuncs = append(AddToMpManagerFuncs, baremetalmachine.Add)
	AddToMpManagerFuncs = append(AddToMpManagerFuncs, hostedmachine.Add)

	cfg, _ := config.NewDefaultConfig()
	mgr := &ProviderManager{
		CpManager: cluster.New(),
		MpManager: machine.New(),
		Cfg:       cfg,
	}

	for _, f := range AddToCpManagerFuncs {
		if err := f(mgr.CpManager, mgr.Cfg); err != nil {
			return nil, err
		}
	}

	for _, f := range AddToMpManagerFuncs {
		if err := f(mgr.MpManager, mgr.Cfg); err != nil {
			return nil, err
		}
	}

	return mgr, nil
}
