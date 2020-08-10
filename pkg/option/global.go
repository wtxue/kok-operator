package option

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/wtxue/kube-on-kube-operator/pkg/k8sclient"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

type GlobalManagerOption struct {
	Kubeconfig              string
	ConfigContext           string
	Namespace               string
	DefaultNamespace        string
	LoggerDevMode           bool
	Threadiness             int
	GoroutineThreshold      int
	ResyncPeriod            time.Duration
	LeaderElectionNamespace string
	EnableLeaderElection    bool
}

func DefaultGlobalManagetOption() *GlobalManagerOption {
	return &GlobalManagerOption{
		LoggerDevMode:           true,
		Threadiness:             1,
		GoroutineThreshold:      1000,
		ResyncPeriod:            60 * time.Minute,
		EnableLeaderElection:    false,
		LeaderElectionNamespace: "onkube-admin",
	}
}

func (o *GlobalManagerOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Kubeconfig, "kubeconfig", "", "Kubernetes configuration file")
	fs.StringVar(&o.ConfigContext, "context", "", "The name of the kubeconfig context to use")
	fs.StringVar(&o.Namespace, "namespace", "", "Config namespace")
	fs.BoolVar(&o.LoggerDevMode, "logger-dev-mode", o.LoggerDevMode, "Enables the Cluster controller manager")
	fs.IntVar(&o.Threadiness, "threadiness", o.Threadiness, "Enables the Machine controller manager")
	fs.IntVar(&o.GoroutineThreshold, "goroutine-threshold", o.GoroutineThreshold, "Enables the Machine controller manager")
}

func (o *GlobalManagerOption) GetK8sConfig() (*rest.Config, error) {
	cfg, err := k8sclient.GetConfigWithContext(o.Kubeconfig, o.ConfigContext)
	if err != nil {
		return nil, errors.Wrap(err, "could not get k8s config")
	}

	// Adjust our client's rate limits based on the number of controllers we are running.
	if cfg.QPS == 0.0 {
		cfg.QPS = 40.0
		cfg.Burst = 60
	}

	return cfg, nil
}

func (o *GlobalManagerOption) GetKubeInterface() (kubernetes.Interface, error) {
	cfg, err := o.GetK8sConfig()
	if err != nil {
		return nil, errors.Wrap(err, "could not get k8s config")
	}

	kubeCli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("failed to get kubernetes Clientset: %v", err)
	}

	return kubeCli, nil
}

func (o *GlobalManagerOption) GetKubeInterfaceOrDie() kubernetes.Interface {
	kubeCli, err := o.GetKubeInterface()
	if err != nil {
		klog.Fatalf("unable to get kube interface err: %v", err)
	}

	return kubeCli
}
