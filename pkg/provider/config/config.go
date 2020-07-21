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

package config

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"strings"
)

type Config struct {
	Registry       Registry
	Audit          Audit
	Feature        Feature
	CustomRegistry string
	CustomeCert    bool
	CustomeImages  bool
}

type Registry struct {
	Prefix    string
	IP        string
	Domain    string
	Namespace string
}

type Audit struct {
	Address string
}

type Feature struct {
	SkipConditions []string
}

func NewDefaultConfig() (*Config, error) {
	config := &Config{
		Registry: Registry{
			Prefix: "registry.cn-hangzhou.aliyuncs.com/wtxue",
			// Prefix: "registry.aliyuncs.com/google_containers",
		},
		CustomRegistry: "registry.cn-hangzhou.aliyuncs.com/wtxue",
	}

	s := strings.Split(config.Registry.Prefix, "/")
	if len(s) != 2 {
		return nil, errors.New("invalid registry prefix")
	}
	config.Registry.Domain = s[0]
	config.Registry.Namespace = s[1]
	config.CustomeCert = true
	config.CustomeImages = true
	return config, nil
}

func (r *Config) NeedSetHosts() bool {
	return r.Registry.IP != ""
}

func (r *Config) ImageFullName(name, tag string) string {
	b := new(bytes.Buffer)
	b.WriteString(name)
	if tag != "" {
		if !strings.Contains(tag, "v") {
			b.WriteString(":" + "v" + tag)
		} else {
			b.WriteString(":" + tag)
		}
	}

	return path.Join(r.Registry.Domain, r.Registry.Namespace, b.String())
}

func (r *Config) KubeAllImageFullName(name, tag string) string {
	if !strings.Contains(tag, "v") {
		tag = "v" + tag
	}

	return fmt.Sprintf("%s/%s:%s", r.CustomRegistry, name, tag)
}

func (r *Config) KubeProxyImagesName(tag string) string {
	if !strings.Contains(tag, "v") {
		tag = "v" + tag
	}

	return fmt.Sprintf("%s/%s:%s", r.CustomRegistry, "kube-proxy", tag)
}
