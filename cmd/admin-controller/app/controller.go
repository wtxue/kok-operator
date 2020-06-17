package app

import (
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/wtxue/kube-on-kube-operator/cmd/admin-controller/app/app_option"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers"
	"github.com/wtxue/kube-on-kube-operator/pkg/k8sclient"
	"github.com/wtxue/kube-on-kube-operator/pkg/static"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/k8sutil"
	ctrlmanager "sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func NewControllerCmd(opt *app_option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ctrl",
		Short: "Manage controller Component",
		Run: func(cmd *cobra.Command, args []string) {
			PrintFlags(cmd.Flags())

			cfg, err := opt.Global.GetK8sConfig()
			if err != nil {
				klog.Fatalf("unable to get cfg err: %v", err)
			}

			if opt.Ctrl.EnableManagerCrds {
				crds, err := static.LoadCRDs()
				if err != nil {
					klog.Fatalf("unable to get cfg err: %v", err)
				}

				err = k8sutil.ReconcileCrds(cfg, crds)
				if err != nil {
					klog.Fatalf("failed to reconcile crd err: %v", err)
				}
			}

			// Adjust our client's rate limits based on the number of controllers we are running.
			cfg.QPS = float32(2) * cfg.QPS
			cfg.Burst = 2 * cfg.Burst

			mgr, err := ctrlmanager.New(cfg, ctrlmanager.Options{
				Scheme:                  k8sclient.GetScheme(),
				LeaderElection:          opt.Global.EnableLeaderElection,
				LeaderElectionNamespace: opt.Global.LeaderElectionNamespace,
				SyncPeriod:              &opt.Global.ResyncPeriod,
				MetricsBindAddress:      "0",
				HealthProbeBindAddress:  ":8090",
				// Port:               9443,
			})
			if err != nil {
				klog.Fatalf("unable to new manager err: %v", err)
			}

			// Setup all Controllers
			klog.Info("Setting up controller")
			if err := controllers.AddToManager(mgr, opt.Ctrl); err != nil {
				klog.Fatalf("unable to register controllers to the manager err: %v", err)
			}

			klog.Info("starting manager")
			stopCh := signals.SetupSignalHandler()
			if err := mgr.Start(stopCh); err != nil {
				klog.Fatalf("problem start running manager err: %v", err)
			}
		},
	}

	opt.Ctrl.AddFlags(cmd.Flags())
	return cmd
}
