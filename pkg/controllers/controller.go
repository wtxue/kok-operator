package controllers

import (
	"github.com/wtxue/kok-operator/pkg/clustermanager"
	"github.com/wtxue/kok-operator/pkg/controllers/cluster"
	"github.com/wtxue/kok-operator/pkg/controllers/machine"
	"github.com/wtxue/kok-operator/pkg/gmanager"
	"github.com/wtxue/kok-operator/pkg/option"
	"github.com/wtxue/kok-operator/pkg/provider"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

var AddToManagerWithProviderFuncs []func(manager.Manager, *gmanager.GManager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, opt *option.ControllersManagerOption, config *config.Config) error {
	if opt.EnableCluster {
		AddToManagerWithProviderFuncs = append(AddToManagerWithProviderFuncs, cluster.Add)
	}

	if opt.EnableMachine {
		AddToManagerWithProviderFuncs = append(AddToManagerWithProviderFuncs, machine.Add)
	}

	if opt.EnableAddons {
		AddToManagerWithProviderFuncs = append(AddToManagerWithProviderFuncs, machine.Add)
	}

	pMgr, err := provider.NewProvider(config)
	if err != nil {
		klog.Errorf("NewProvider err: %v", err)
		return err
	}

	k8sMgr, _ := clustermanager.NewManager(clustermanager.MasterClient{
		Manager: m,
	})

	var gMgr = &gmanager.GManager{
		ProviderManager: pMgr,
		ClusterManager:  k8sMgr,
		Config:          config,
	}
	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}

	for _, f := range AddToManagerWithProviderFuncs {
		if err := f(m, gMgr); err != nil {
			return err
		}
	}

	m.Add(gMgr.ClusterManager)
	return nil
}
