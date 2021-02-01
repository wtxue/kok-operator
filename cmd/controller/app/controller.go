package app

import (
	"github.com/spf13/cobra"
	"github.com/wtxue/kok-operator/cmd/controller/app/app_option"
	"github.com/wtxue/kok-operator/pkg/controllers"
	"github.com/wtxue/kok-operator/pkg/k8sclient"
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	"github.com/wtxue/kok-operator/pkg/static"
	"k8s.io/klog/v2"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func NewControllerCmd(opt *app_option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ctrl",
		Short: "Manage controller Component",
		Run: func(cmd *cobra.Command, args []string) {
			PrintFlags(cmd.Flags())
			opt.Global.SetupLogger()

			cfg, err := opt.Global.GetK8sConfig()
			if err != nil {
				klog.Fatalf("unable to get cfg err: %v", err)
			}

			if opt.Ctrl.EnableManagerCrds {
				crds, err := static.LoadCRDs()
				if err != nil {
					klog.Fatalf("unable to get cfg err: %v", err)
				}

				err = k8sutil.ReconcileCRDs(cfg, crds)
				if err != nil {
					klog.Fatalf("failed to reconcile crd err: %v", err)
				}
			}

			// Adjust our client's rate limits based on the number of controllers we are running.
			cfg.QPS = float32(2) * cfg.QPS
			cfg.Burst = 2 * cfg.Burst

			mgr, err := manager.New(cfg, manager.Options{
				Scheme:                  k8sclient.GetScheme(),
				LeaderElection:          opt.Global.EnableLeaderElection,
				LeaderElectionNamespace: opt.Global.LeaderElectionNamespace,
				SyncPeriod:              &opt.Global.ResyncPeriod,
				MetricsBindAddress:      "0", // disable metrics with manager, use our observe
				HealthProbeBindAddress:  "0", // disable health probe with manager, use our observe
			})
			if err != nil {
				klog.Fatalf("unable to new manager err: %v", err)
			}

			// Setup all Controllers
			ctrlrt.Log.Info("Setting up controller")
			if err := controllers.AddToManager(mgr, opt.Ctrl, opt.Provider); err != nil {
				klog.Fatalf("unable to register controllers to the manager err: %v", err)
			}

			ctrlrt.Log.Info("starting manager")
			stopCh := signals.SetupSignalHandler()
			if err := mgr.Start(stopCh); err != nil {
				klog.Fatalf("problem start running manager err: %v", err)
			}
		},
	}

	opt.Ctrl.AddFlags(cmd.Flags())
	opt.Provider.AddFlags(cmd.Flags())
	return cmd
}
