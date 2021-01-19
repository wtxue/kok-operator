package static

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/k8sclient"
	crdgenerated "github.com/wtxue/kok-operator/pkg/static/crds/generated"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func load(f io.Reader) ([]*apiextensionsv1.CustomResourceDefinition, error) {
	var b bytes.Buffer

	var yamls []string

	crds := make([]*apiextensionsv1.CustomResourceDefinition, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			yamls = append(yamls, b.String())
			b.Reset()
		} else {
			if _, err := b.WriteString(line); err != nil {
				return crds, err
			}
			if _, err := b.WriteString("\n"); err != nil {
				return crds, err
			}
		}
	}
	if s := strings.TrimSpace(b.String()); s != "" {
		yamls = append(yamls, s)
	}

	for _, yaml := range yamls {
		s := json.NewSerializerWithOptions(json.DefaultMetaFactory,
			k8sclient.GetScheme(), k8sclient.GetScheme(), json.SerializerOptions{Yaml: true})

		obj, _, err := s.Decode([]byte(yaml), nil, nil)
		if err != nil {
			continue
		}

		var crd *apiextensionsv1.CustomResourceDefinition
		var ok bool
		if crd, ok = obj.(*apiextensionsv1.CustomResourceDefinition); !ok {
			continue
		}

		crd.Status = apiextensionsv1.CustomResourceDefinitionStatus{}
		crd.SetGroupVersionKind(schema.GroupVersionKind{})
		if crd.Labels == nil {
			crd.Labels = make(map[string]string)
		}
		crd.Labels[constants.CreatedByLabel] = constants.CreatedBy
		crds = append(crds, crd)
	}

	return crds, nil
}

func LoadCRDs() ([]*apiextensionsv1.CustomResourceDefinition, error) {
	crds := make([]*apiextensionsv1.CustomResourceDefinition, 0)
	dir, err := crdgenerated.CRDs.Open("/")
	if err != nil {
		return crds, err
	}

	dirFiles, err := dir.Readdir(-1)
	if err != nil {
		return crds, err
	}
	for _, file := range dirFiles {
		f, err := crdgenerated.CRDs.Open(file.Name())
		if err != nil {
			return crds, err
		}

		tmp, err := load(f)
		if err != nil {
			return crds, err
		}

		crds = append(crds, tmp...)
	}

	return crds, nil
}
