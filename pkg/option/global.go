package option

import (
	"time"

	"github.com/wtxue/kok-operator/pkg/k8sclient"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	ctrlrt "sigs.k8s.io/controller-runtime"
)

type GlobalManagerOption struct {
	Kubeconfig              string
	ConfigContext           string
	Namespace               string
	DefaultNamespace        string
	Threadiness             int
	GoroutineThreshold      int
	ResyncPeriod            time.Duration
	LeaderElectionNamespace string
	EnableLeaderElection    bool
	EnableDevLogging        bool
	LogLevel                string
}

func DefaultGlobalManagetOption() *GlobalManagerOption {
	return &GlobalManagerOption{
		Threadiness:             1,
		GoroutineThreshold:      1000,
		ResyncPeriod:            60 * time.Minute,
		EnableLeaderElection:    false,
		LeaderElectionNamespace: "kok-system",
		EnableDevLogging:        true,
		LogLevel:                "info",
	}
}

func (o *GlobalManagerOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Kubeconfig, "kubeconfig", "", "Kubernetes configuration file")
	fs.StringVar(&o.ConfigContext, "context", "", "The name of the kubeconfig context to use")
	fs.StringVar(&o.Namespace, "namespace", "", "Config namespace")
	fs.IntVar(&o.Threadiness, "threadiness", o.Threadiness, "Enables the Machine controller manager")
	fs.IntVar(&o.GoroutineThreshold, "goroutine-threshold", o.GoroutineThreshold, "Enables the Machine controller manager")
	fs.BoolVar(&o.EnableDevLogging, "enable-dev-logging", o.EnableDevLogging,
		"Configures the logger to use a Zap development config (encoder=consoleEncoder,logLevel=Debug,stackTraceLevel=Warn, no sampling), "+
			"otherwise a Zap production config will be used (encoder=jsonEncoder,logLevel=Info,stackTraceLevel=Error), sampling).")
	fs.StringVar(&o.LogLevel, "log-level", o.LogLevel,
		"The log level. Default is info. We use logr interface which only supports info and debug level",
	)
}

func (o *GlobalManagerOption) GetK8sConfig() (*rest.Config, error) {
	cfg, err := k8sclient.GetConfigWithContext(o.Kubeconfig, o.ConfigContext)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get k8s config")
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
		return nil, errors.Wrap(err, "failed to get k8s config")
	}

	kubeCli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("failed to get kubernetes clientset: %v", err)
	}

	return kubeCli, nil
}

func (o *GlobalManagerOption) GetKubeInterfaceOrDie() kubernetes.Interface {
	kubeCli, err := o.GetKubeInterface()
	if err != nil {
		klog.Fatalf("failed to get kube interface err: %v", err)
	}

	return kubeCli
}

// SetupLogger initializes the logger used in the service controller
func (o *GlobalManagerOption) SetupLogger() {
	// var lvl zapcore.LevelEnabler
	//
	// switch o.LogLevel {
	// case "debug":
	// 	lvl = zapcore.DebugLevel
	// default:
	// 	lvl = zapcore.InfoLevel
	// }
	//
	// zapOptions := &zap.Options{
	// 	Development: o.EnableDevLogging,
	// 	Level:       lvl,
	// }
	// ctrlrt.SetLogger(zap.New(zap.UseFlagOptions(zapOptions)))
	ctrlrt.SetLogger(klogr.New())
}
