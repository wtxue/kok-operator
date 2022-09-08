package apis

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubeproxyv1alpha1 "k8s.io/kube-proxy/config/v1alpha1"
	kubeletv1beta1 "k8s.io/kubelet/config/v1beta1"
	kubeadmv1beta2 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2"
	kubeadmv1beta3 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

var Scheme = runtime.NewScheme()

var Codecs = serializer.NewCodecFactory(Scheme)

var localSchemeBuilder = runtime.SchemeBuilder{
	kubeadmv1beta2.AddToScheme,
	kubeadmv1beta3.AddToScheme,
	kubeletv1beta1.AddToScheme,
	kubeproxyv1alpha1.AddToScheme,
}

var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	utilruntime.Must(AddToScheme(Scheme))
}

// MarshalToYAML marshals an object into yaml.
func MarshalToYAML(obj runtime.Object, gv schema.GroupVersion) ([]byte, error) {
	const mediaType = runtime.ContentTypeYAML
	info, ok := runtime.SerializerInfoForMediaType(Codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return []byte{}, errors.Errorf("unsupported media type %q", mediaType)
	}
	encoder := Codecs.EncoderForVersion(info.Serializer, gv)
	return runtime.Encode(encoder, obj)
}

// GetScheme gets an initialized runtime.Scheme with kubeadm, kubelet, kubeproxy
func GetScheme() *runtime.Scheme {
	return Scheme
}
