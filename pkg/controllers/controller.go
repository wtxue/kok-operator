package controllers

import (
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/cluster"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/machine"
	"github.com/wtxue/kube-on-kube-operator/pkg/option"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, opt *option.ControllersManagerOption) error {
	if opt.EnableCluster {
		AddToManagerFuncs = append(AddToManagerFuncs, cluster.Add)
	}

	if opt.EnableMachine {
		AddToManagerFuncs = append(AddToManagerFuncs, machine.Add)
	}

	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}

	return nil
}
