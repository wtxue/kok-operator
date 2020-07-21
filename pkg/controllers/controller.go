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

package controllers

import (
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/cluster"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/k8smanager"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/machine"
	"github.com/wtxue/kube-on-kube-operator/pkg/gmanager"
	"github.com/wtxue/kube-on-kube-operator/pkg/option"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

var AddToManagerWithProviderFuncs []func(manager.Manager, *gmanager.GManager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, opt *option.ControllersManagerOption) error {
	if opt.EnableCluster {
		AddToManagerWithProviderFuncs = append(AddToManagerWithProviderFuncs, cluster.Add)
	}

	if opt.EnableMachine {
		AddToManagerWithProviderFuncs = append(AddToManagerWithProviderFuncs, machine.Add)
	}

	pMgr, err := provider.NewProvider()
	if err != nil {
		klog.Errorf("NewProvider err: %v", err)
		return err
	}

	k8sMgr, _ := k8smanager.NewManager(k8smanager.MasterClient{
		Manager: m,
	})

	var gMgr = &gmanager.GManager{
		ProviderManager: pMgr,
		ClusterManager:  k8sMgr,
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
