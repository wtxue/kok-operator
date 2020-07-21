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

package option

import (
	"github.com/spf13/pflag"
)

type ControllersManagerOption struct {
	EnableCluster     bool
	EnableMachine     bool
	EnableManagerCrds bool
}

func DefaultControllersManagerOption() *ControllersManagerOption {
	return &ControllersManagerOption{
		EnableCluster:     true,
		EnableMachine:     true,
		EnableManagerCrds: false,
	}
}

func (o *ControllersManagerOption) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.EnableCluster, "enable-cluster", o.EnableCluster, "Enables the Cluster controller manager")
	fs.BoolVar(&o.EnableMachine, "enable-machine", o.EnableMachine, "Enables the Machine controller manager")
	fs.BoolVar(&o.EnableManagerCrds, "enable-manager-crds", o.EnableManagerCrds, "Enables to manager the associated crds")
}
