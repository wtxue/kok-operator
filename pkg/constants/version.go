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

package constants

import (
	"github.com/thoas/go-funk"
	"k8s.io/klog"
)

var (
	OSs              = []string{"linux"}
	K8sVersions      = []string{"v1.16.13", "v1.18.5"}
	K8sVersionsWithV = funk.Map(K8sVersions, func(s string) string {
		return "v" + s
	}).([]string)
	K8sVersionConstraint = ">= 1.10"
	DockerVersions       = []string{"18.09.9"}
)

func IsK8sSupport(version string) bool {
	for _, v := range K8sVersions {
		if v == version {
			return true
		}
	}

	klog.Errorf("k8s version only support: %#v", K8sVersions)
	return false
}
