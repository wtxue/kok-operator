package k8sutil

import (
	"io"

	"github.com/pkg/errors"

	"bufio"
	"bytes"
	"strings"

	"github.com/wtxue/kube-on-kube-operator/pkg/k8sclient"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
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

func LoadObjs(f io.Reader) ([]runtime.Object, error) {
	var b bytes.Buffer

	var yamls []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			yamls = append(yamls, b.String())
			b.Reset()
		} else {
			if _, err := b.WriteString(line); err != nil {
				return nil, err
			}
			if _, err := b.WriteString("\n"); err != nil {
				return nil, err
			}
		}
	}
	if s := strings.TrimSpace(b.String()); s != "" {
		yamls = append(yamls, s)
	}

	objs := make([]runtime.Object, 0)
	for _, yaml := range yamls {
		if len(yaml) < 10 {
			continue
		}

		s := json.NewSerializerWithOptions(json.DefaultMetaFactory,
			k8sclient.GetScheme(), k8sclient.GetScheme(), json.SerializerOptions{Yaml: true})

		obj, gvk, err := s.Decode([]byte(yaml), nil, nil)
		if err != nil {
			continue
		}

		if gvk != nil {
			key, err := runtimeClient.ObjectKeyFromObject(obj)
			if err != nil {
				continue
			}
			klog.Infof("append gvk: %s, key: %s", gvk.String(), key.String())
		}
		objs = append(objs, obj)

	}

	return objs, nil
}
