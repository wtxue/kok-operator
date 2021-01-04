/*
Copyright 2020 wtxue.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package static

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/k8sclient"
	crdgenerated "github.com/wtxue/kok-operator/pkg/static/crds/generated"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
