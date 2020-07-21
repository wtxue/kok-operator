package apis

import (
	"github.com/pkg/errors"
	kubeadmv1beta2 "github.com/wtxue/kok-operator/pkg/apis/kubeadm/v1beta2"
	kubeletv1beta1 "github.com/wtxue/kok-operator/pkg/apis/kubelet/config/v1beta1"
	kubeproxyv1alpha1 "github.com/wtxue/kok-operator/pkg/apis/kubeproxy/config/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var localSchemeBuilder = runtime.SchemeBuilder{
	kubeadmv1beta2.AddToScheme,
	kubeletv1beta1.AddToScheme,
	kubeproxyv1alpha1.AddToScheme,
}
var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	utilruntime.Must(AddToScheme(Scheme))
}

// MarshalToYaml marshals an object into yaml.
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
