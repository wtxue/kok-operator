package app_option

import (
	"github.com/spf13/pflag"
	"github.com/wtxue/kok-operator/pkg/option"
	"github.com/wtxue/kok-operator/pkg/provider/config"
)

// Options ...
type Options struct {
	Global   *option.GlobalManagerOption
	Ctrl     *option.ControllersManagerOption
	Provider *config.Config
}

// NewOptions creates a new Options with a default config.
func NewOptions() *Options {
	return &Options{
		Global:   option.DefaultGlobalManagetOption(),
		Ctrl:     option.DefaultControllersManagerOption(),
		Provider: config.NewDefaultConfig(),
	}
}

// AddFlags adds flags for a specific server to the specified FlagSet object.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Global.AddFlags(fs)
	o.Ctrl.AddFlags(fs)
	o.Provider.AddFlags(fs)
}
