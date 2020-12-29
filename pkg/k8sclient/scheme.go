package k8sclient

import (
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	workloadv1 "github.com/wtxue/kok-operator/pkg/apis/workload/v1"

	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	apiregistrationscheme "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/scheme"
	//  monitorv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextensionsscheme.AddToScheme(scheme)
	_ = apiregistrationscheme.AddToScheme(scheme)
	_ = devopsv1.AddToScheme(scheme)
	_ = workloadv1.AddToScheme(scheme)
	// _ = monitorv1.AddToScheme(scheme)
}

// GetScheme gets an initialized runtime.Scheme with k8s core added by default
func GetScheme() *runtime.Scheme {
	return scheme
}
