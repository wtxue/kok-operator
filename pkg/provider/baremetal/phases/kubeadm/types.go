package kubeadm

import (
	"bytes"
	"reflect"

	"github.com/pkg/errors"
	kubeadmv1beta2 "github.com/wtxue/kube-on-kube-operator/pkg/apis/kubeadm/v1beta2"
	kubeletv1beta1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/kubelet/config/v1beta1"
	kubeproxyv1alpha1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/kubeproxy/config/v1alpha1"
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

type Config struct {
	InitConfiguration      *kubeadmv1beta2.InitConfiguration
	ClusterConfiguration   *kubeadmv1beta2.ClusterConfiguration
	JoinConfiguration      *kubeadmv1beta2.JoinConfiguration
	KubeletConfiguration   *kubeletv1beta1.KubeletConfiguration
	KubeProxyConfiguration *kubeproxyv1alpha1.KubeProxyConfiguration
}

func (c *Config) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	v := reflect.ValueOf(*c)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsNil() {
			continue
		}
		obj, ok := v.Field(i).Interface().(runtime.Object)
		if !ok {
			panic("no runtime.Object")
		}
		gvks, _, err := Scheme.ObjectKinds(obj)
		if err != nil {
			return nil, err
		}

		yamlData, err := MarshalToYAML(obj, gvks[0].GroupVersion())
		if err != nil {
			return nil, err
		}
		buf.WriteString("---\n")
		buf.Write(yamlData)
	}

	return buf.Bytes(), nil
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
