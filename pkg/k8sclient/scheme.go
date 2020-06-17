package k8sclient

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//  monitorv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextensionsv1beta1.AddToScheme(scheme)

	_ = devopsv1.AddToScheme(scheme)
	// _ = monitorv1.AddToScheme(scheme)
}

// GetScheme gets an initialized runtime.Scheme with k8s core added by default
func GetScheme() *runtime.Scheme {
	return scheme
}
