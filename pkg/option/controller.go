package option

import (
	"github.com/spf13/pflag"
)

type ControllersManagerOption struct {
	EnableManagerCrds bool
	EnableCluster     bool
	EnableMachine     bool
	EnableAddons      bool
}

func DefaultControllersManagerOption() *ControllersManagerOption {
	return &ControllersManagerOption{
		EnableCluster:     true,
		EnableMachine:     true,
		EnableAddons:      true,
		EnableManagerCrds: false,
	}
}

func (o *ControllersManagerOption) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.EnableManagerCrds, "enable-manager-crds", o.EnableManagerCrds, "Enables to manager the associated crds")
	fs.BoolVar(&o.EnableCluster, "enable-cluster", o.EnableCluster, "Enables the Cluster controller manager")
	fs.BoolVar(&o.EnableMachine, "enable-machine", o.EnableMachine, "Enables the Machine controller manager")
	fs.BoolVar(&o.EnableMachine, "enable-addons", o.EnableAddons, "Enables the cluster addons controller manager")
}
