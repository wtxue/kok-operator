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

const (
	ComponentNameCrd = "crds"
	CreatedByLabel   = "k8s.io/created-by"
	CreatedBy        = "operator"

	KubeApiServer         = "kube-apiserver"
	KubeKubeScheduler     = "kube-scheduler"
	KubeControllerManager = "kube-controller-manager"
	KubeApiServerCerts    = "kube-apiserver-certs"
	KubeApiServerConfig   = "kube-apiserver-config"
	KubeApiServerAudit    = "kube-apiserver-audit"
	KubeMasterManifests   = "kube-master-manifests"
)

const (
	ClusterAnnoApplySep      = "k8s.io/apply.step"
	ClusterPhaseRestore      = "k8s.io/step.restore"
	ClusterApiSvcType        = "k8s.io/apiserver.type"
	ClusterApiSvcVip         = "k8s.io/apiserver.vip"
	ClusterAnnoLocalDebugDir = "k8s.io/local.dir"
)

var KubeApiServerLabels = map[string]string{
	"component": KubeApiServer,
}

var KubeKubeSchedulerLabels = map[string]string{
	"component": KubeKubeScheduler,
}

var KubeControllerManagerLabels = map[string]string{
	"component": KubeControllerManager,
}

var CtrlLabels = map[string]string{
	"createBy": "controller",
}

func GetAnnotationKey(annotation map[string]string, key string) string {
	if k, ok := annotation[key]; ok {
		return k
	}

	return ""
}
