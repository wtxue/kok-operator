package commands

import (
	// monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	apiregistrationscheme "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/scheme"
	// metricsscheme "k8s.io/metrics/pkg/client/clientset/versioned/scheme"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextensionsscheme.AddToScheme(scheme)
	_ = apiregistrationscheme.AddToScheme(scheme)
	// _ = monitorv1.AddToScheme(scheme)
	// _ = metricsscheme.AddToScheme(scheme)

}

// GetScheme gets an initialized runtime.Scheme with default control cluster
func GetScheme() *runtime.Scheme {
	return scheme
}
