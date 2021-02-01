package controllers

import (
	"github.com/wtxue/kok-operator/pkg/clustermanager"
	"github.com/wtxue/kok-operator/pkg/controllers/addons"
	"github.com/wtxue/kok-operator/pkg/controllers/cluster"
	"github.com/wtxue/kok-operator/pkg/controllers/machine"
	"github.com/wtxue/kok-operator/pkg/gmanager"
	"github.com/wtxue/kok-operator/pkg/option"
	"github.com/wtxue/kok-operator/pkg/provider"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

var AddToManagerWithProviderFuncs []func(manager.Manager, *gmanager.GManager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(mgr manager.Manager, opt *option.ControllersManagerOption, config *config.Config) error {
	if opt.EnableCluster {
		AddToManagerWithProviderFuncs = append(AddToManagerWithProviderFuncs, cluster.Add)
	}

	if opt.EnableMachine {
		AddToManagerWithProviderFuncs = append(AddToManagerWithProviderFuncs, machine.Add)
	}

	if opt.EnableAddons {
		AddToManagerWithProviderFuncs = append(AddToManagerWithProviderFuncs, addons.Add)
	}

	pMgr, err := provider.NewProvider(config)
	if err != nil {
		return err
	}

	k8sMgr, _ := clustermanager.NewManager(clustermanager.ControlCluster{
		Cluster: mgr,
	})

	gMgr := &gmanager.GManager{
		ProviderManager: pMgr,
		ClusterManager:  k8sMgr,
		Config:          config,
	}
	for _, f := range AddToManagerFuncs {
		if err := f(mgr); err != nil {
			return err
		}
	}

	for _, f := range AddToManagerWithProviderFuncs {
		if err := f(mgr, gMgr); err != nil {
			return err
		}
	}

	mgr.Add(gMgr.ClusterManager)
	return nil
}
