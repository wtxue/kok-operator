package provider

import (
	baremetalcluster "github.com/wtxue/kok-operator/pkg/provider/baremetal/cluster"
	baremetalmachine "github.com/wtxue/kok-operator/pkg/provider/baremetal/machine"
	"github.com/wtxue/kok-operator/pkg/provider/cluster"
	clusterprovider "github.com/wtxue/kok-operator/pkg/provider/cluster"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"github.com/wtxue/kok-operator/pkg/provider/machine"
	machineprovider "github.com/wtxue/kok-operator/pkg/provider/machine"
	managedcluster "github.com/wtxue/kok-operator/pkg/provider/managed/cluster"
	managedmachine "github.com/wtxue/kok-operator/pkg/provider/managed/machine"
)

type ProviderManager struct {
	*cluster.CpManager
	*machine.MpManager
	Cfg *config.Config
}

var AddToCpManagerFuncs []func(*clusterprovider.CpManager, *config.Config) error
var AddToMpManagerFuncs []func(*machineprovider.MpManager, *config.Config) error

func NewProvider(config *config.Config) (*ProviderManager, error) {
	AddToCpManagerFuncs = append(AddToCpManagerFuncs, baremetalcluster.Add)
	AddToCpManagerFuncs = append(AddToCpManagerFuncs, managedcluster.Add)

	AddToMpManagerFuncs = append(AddToMpManagerFuncs, baremetalmachine.Add)
	AddToMpManagerFuncs = append(AddToMpManagerFuncs, managedmachine.Add)

	mgr := &ProviderManager{
		CpManager: cluster.New(),
		MpManager: machine.New(),
		Cfg:       config,
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
