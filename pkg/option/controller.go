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
