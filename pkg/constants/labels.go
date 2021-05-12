package constants

import (
	"fmt"
)

const (
	ComponentNameCrd = "crds"
	CreatedByLabel   = "fake.io/created-by"
	CreatedBy        = "operator"

	KubeApiServer         = "kube-apiserver"
	KubeKubeScheduler     = "kube-scheduler"
	KubeControllerManager = "kube-controller-manager"
	KubeApiServerCerts    = "kube-apiserver-certs"
	KubeApiServerConfig   = "kube-apiserver-config"
	KubeApiServerAudit    = "kube-apiserver-audit"
)

const (
	ClusterUpdateStep    = "fake.io/update.step"
	ClusterRestoreStep   = "fake.io/restore.step"
	ClusterApiserverType = "fake.io/apiserver.type"
	ClusterApiserverVip  = "fake.io/apiserver.vip"
	ClusterDebugLocalDir = "fake.io/debug.localdir"
)

var CtrlLabels = map[string]string{
	"createBy": "controller",
}

func GetMapKey(annotation map[string]string, key string) string {
	if k, ok := annotation[key]; ok {
		return k
	}

	return ""
}

// GenComponentName ...
func GenComponentName(clusterID, suffix string) string {
	return fmt.Sprintf("%s-%s", clusterID, suffix)
}
