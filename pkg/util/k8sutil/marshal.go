package k8sutil

import (
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

// MarshalToYaml marshals an object into yaml.
func MarshalToYaml(obj runtime.Object, gv schema.GroupVersion) ([]byte, error) {
	return MarshalToYamlForCodecs(obj, gv, clientsetscheme.Codecs)
}

// MarshalToYamlForCodecs marshals an object into yaml using the specified codec
// TODO: Is specifying the gv really needed here?
// TODO: Can we support json out of the box easily here?
func MarshalToYamlForCodecs(obj runtime.Object, gv schema.GroupVersion, codecs serializer.CodecFactory) ([]byte, error) {
	const mediaType = runtime.ContentTypeYAML
	info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return []byte{}, errors.Errorf("unsupported media type %q", mediaType)
	}

	encoder := codecs.EncoderForVersion(info.Serializer, gv)
	return runtime.Encode(encoder, obj)
}

// UnmarshalFromYaml unmarshals yaml into an object.
func UnmarshalFromYaml(buffer []byte, gv schema.GroupVersion) (runtime.Object, error) {
	return UnmarshalFromYamlForCodecs(buffer, gv, clientsetscheme.Codecs)
}

// UnmarshalFromYamlForCodecs unmarshals yaml into an object using the specified codec
// TODO: Is specifying the gv really needed here?
// TODO: Can we support json out of the box easily here?
func UnmarshalFromYamlForCodecs(buffer []byte, gv schema.GroupVersion, codecs serializer.CodecFactory) (runtime.Object, error) {
	const mediaType = runtime.ContentTypeYAML
	info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return nil, errors.Errorf("unsupported media type %q", mediaType)
	}

	decoder := codecs.DecoderToVersion(info.Serializer, gv)
	return runtime.Decode(decoder, buffer)
}

// GroupVersionKindsHasKind returns whether the following gvk slice contains the kind given as a parameter
func GroupVersionKindsHasKind(gvks []schema.GroupVersionKind, kind string) bool {
	for _, gvk := range gvks {
		if gvk.Kind == kind {
			return true
		}
	}
	return false
}
