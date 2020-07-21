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

package cluster

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/common"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/k8sutil"
	appsv1 "k8s.io/api/apps/v1"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	Obj     *common.Cluster
	dynamic dynamic.Interface
	*Provider
}

func GetPodBindPort() int32 {
	return 6443
}

// GetHPAReplicaCountOrDefault get desired replica count from HPA if exists, returns the given default otherwise
func GetHPAReplicaCountOrDefault(client client.Client, name types.NamespacedName, defaultReplicaCount int32) int32 {
	var hpa autoscalev2beta1.HorizontalPodAutoscaler
	err := client.Get(context.Background(), name, &hpa)
	if err != nil {
		return defaultReplicaCount
	}

	if hpa.Spec.MinReplicas != nil && hpa.Status.DesiredReplicas < *hpa.Spec.MinReplicas {
		return *hpa.Spec.MinReplicas
	}

	return hpa.Status.DesiredReplicas
}

func GetAdvertiseAddress(obj *common.Cluster) string {
	advertiseAddress := "$(INSTANCE_IP)"
	if obj.Cluster.Spec.Features.HA != nil && obj.Cluster.Spec.Features.HA.ThirdPartyHA != nil {
		advertiseAddress = obj.Cluster.Spec.Features.HA.ThirdPartyHA.VIP
	}

	return advertiseAddress
}

func (r *Reconciler) apiServerDeployment() runtime.Object {
	containers := []corev1.Container{}
	vms := []corev1.VolumeMount{
		{
			Name:      constants.KubeApiServerCerts,
			MountPath: "/etc/kubernetes/pki/",
			ReadOnly:  true,
		},
		{
			Name:      constants.KubeApiServerConfig,
			MountPath: "/etc/kubernetes/",
		},
		{
			Name:      constants.KubeApiServerAudit,
			MountPath: "/var/log/kubernetes",
		},
	}
	hostPathType := corev1.HostPathDirectoryOrCreate
	volumes := []corev1.Volume{
		{
			Name: constants.KubeApiServerCerts,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: constants.KubeApiServerCerts,
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
		{
			Name: constants.KubeApiServerConfig,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: constants.KubeApiServerConfig,
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
		{
			Name: constants.KubeApiServerAudit,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: fmt.Sprintf("/web/%s/kube-apiserver/audit", r.Obj.Cluster.Name),
					Type: &hostPathType,
				},
			},
		},
	}

	cmds := []string{
		"kube-apiserver",
		"--allow-privileged=true",
		"--authorization-mode=Node,RBAC",
		"--client-ca-file=/etc/kubernetes/pki/ca.crt",
		"--enable-admission-plugins=NodeRestriction",
		"--enable-bootstrap-token-auth=true",
		"--kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt",
		"--kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key",
		"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
		"--proxy-client-cert-file=/etc/kubernetes/pki/front-proxy-client.crt",
		"--proxy-client-key-file=/etc/kubernetes/pki/front-proxy-client.key",
		"--requestheader-allowed-names=front-proxy-client",
		"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
		"--requestheader-extra-headers-prefix=X-Remote-Extra-",
		"--requestheader-group-headers=X-Remote-Group",
		"--requestheader-username-headers=X-Remote-User",
		"--service-account-key-file=/etc/kubernetes/pki/sa.pub",
		"--tls-cert-file=/etc/kubernetes/pki/apiserver.crt",
		"--tls-private-key-file=/etc/kubernetes/pki/apiserver.key",
		"--token-auth-file=/etc/kubernetes/pki/known_tokens.csv",
	}

	advertiseAddress := GetAdvertiseAddress(r.Obj)
	cmds = append(cmds, fmt.Sprintf("--secure-port=%d", GetPodBindPort()))
	cmds = append(cmds, fmt.Sprintf("--advertise-address=%s", advertiseAddress))
	if r.Obj.Cluster.Spec.APIServerExtraArgs != nil {
		extraArgs := []string{}
		for k, v := range r.Obj.Cluster.Spec.APIServerExtraArgs {
			extraArgs = append(extraArgs, fmt.Sprintf("--%s=%s", k, v))
		}
		sort.Strings(extraArgs)
		cmds = append(cmds, extraArgs...)
	}

	svcCidr := "10.96.0.0/16"
	if r.Obj.Cluster.Spec.ServiceCIDR != nil {
		svcCidr = *r.Obj.Cluster.Spec.ServiceCIDR
	}

	cmds = append(cmds, fmt.Sprintf("--service-cluster-ip-range=%s", svcCidr))
	if r.Obj.Cluster.Spec.Etcd != nil && r.Obj.Cluster.Spec.Etcd.External != nil {
		cmds = append(cmds, fmt.Sprintf("--etcd-servers=%s", strings.Join(r.Obj.Cluster.Spec.Etcd.External.Endpoints, ",")))
		// tode check
		if strings.Contains(r.Obj.Cluster.Spec.Etcd.External.Endpoints[0], "https") {
			cmds = append(cmds, fmt.Sprintf("--etcd-cafile=%s", r.Obj.Cluster.Spec.Etcd.External.CAFile))
			cmds = append(cmds, fmt.Sprintf("--etcd-certfile=%s", r.Obj.Cluster.Spec.Etcd.External.CertFile))
			cmds = append(cmds, fmt.Sprintf("--etcd-keyfile=%s", r.Obj.Cluster.Spec.Etcd.External.KeyFile))
		}
	} else {
		cmds = append(cmds, fmt.Sprintf("--etcd-servers=%s", "http://etcd-0.etcd:2379,http://etcd-1.etcd:2379,http://etcd-2.etcd:2379"))
	}

	c := corev1.Container{
		Name:            constants.KubeApiServer,
		Image:           r.Provider.Cfg.KubeAllImageFullName(constants.KubernetesAllImageName, r.Obj.Cluster.Spec.Version),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         cmds,
		Ports: []corev1.ContainerPort{
			{
				Name:          "https",
				ContainerPort: GetPodBindPort(),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/healthz",
					Port:   intstr.FromString("https"),
					Scheme: corev1.URISchemeHTTPS,
				},
			},

			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			TimeoutSeconds:      15,
			FailureThreshold:    8,
			SuccessThreshold:    1,
		},
		Env: common.ComponentEnv(r.Obj),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.1"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},

		VolumeMounts:             vms,
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}

	containers = append(containers, c)

	deployment := &appsv1.Deployment{
		ObjectMeta: k8sutil.ObjectMeta(constants.KubeApiServer, constants.KubeApiServerLabels, r.Obj.Cluster),
		Spec: appsv1.DeploymentSpec{
			Replicas: k8sutil.IntPointer(3),
			Strategy: common.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: constants.KubeApiServerLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: constants.KubeApiServerLabels,
				},
				Spec: corev1.PodSpec{
					Containers:  containers,
					Volumes:     volumes,
					Affinity:    common.ComponentAffinity(r.Obj.Cluster.Namespace, constants.KubeApiServerLabels),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}

func (r *Reconciler) apiServerSvc() runtime.Object {
	svc := &corev1.Service{
		ObjectMeta: k8sutil.ObjectMeta(constants.KubeApiServer, constants.KubeApiServerLabels, r.Obj.Cluster),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       GetPodBindPort(),
					NodePort:   30443,
					TargetPort: intstr.FromString("https"),
				},
			},

			Selector: constants.KubeApiServerLabels,
		},
	}
	svcType := corev1.ServiceTypeNodePort
	if constants.GetAnnotationKey(r.Obj.Annotations, constants.ClusterApiSvcType) == string(corev1.ServiceTypeLoadBalancer) {
		svcType = corev1.ServiceTypeLoadBalancer
		svc.Spec.LoadBalancerIP = constants.GetAnnotationKey(r.Obj.Annotations, constants.ClusterApiSvcVip)
	}
	svc.Spec.Type = svcType

	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}

	// svc.Annotations["contour.heptio.com/upstream-protocol.tls"] = "443,https"
	svc.Annotations["projectcontour.io/upstream-protocol.tls"] = "6443"
	svc.Annotations["gloo.solo.io/sslService.secret"] = constants.KubeApiServerCerts

	return svc
}

func (r *Reconciler) controllerManagerDeployment() runtime.Object {
	containers := []corev1.Container{}
	vms := []corev1.VolumeMount{
		{
			Name:      constants.KubeApiServerCerts,
			MountPath: "/etc/kubernetes/pki/",
			ReadOnly:  true,
		},
		{
			Name:      constants.KubeApiServerConfig,
			MountPath: "/etc/kubernetes/",
		},
	}
	volumes := []corev1.Volume{
		{
			Name: constants.KubeApiServerCerts,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: constants.KubeApiServerCerts,
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
		{
			Name: constants.KubeApiServerConfig,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: constants.KubeApiServerConfig,
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
	}

	cmds := []string{
		"kube-controller-manager",
		"--authentication-kubeconfig=/etc/kubernetes/controller-manager.conf",
		"--authorization-kubeconfig=/etc/kubernetes/controller-manager.conf",
		"--client-ca-file=/etc/kubernetes/pki/ca.crt",
		"--cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt",
		"--cluster-signing-key-file=/etc/kubernetes/pki/ca.key",
		"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
		"--kubeconfig=/etc/kubernetes/controller-manager.conf",
		"--leader-elect=true",
		"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
		"--root-ca-file=/etc/kubernetes/pki/ca.crt",
		"--service-account-private-key-file=/etc/kubernetes/pki/sa.key",
		"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
		"--use-service-account-credentials=true",
	}

	if r.Obj.Cluster.Spec.ControllerManagerExtraArgs != nil {
		extraArgs := []string{}
		for k, v := range r.Obj.Cluster.Spec.ControllerManagerExtraArgs {
			extraArgs = append(extraArgs, fmt.Sprintf("--%s=%s", k, v))
		}
		sort.Strings(extraArgs)
		cmds = append(cmds, extraArgs...)
	}

	if r.Obj.Cluster.Status.NodeCIDRMaskSize > 0 {
		cmds = append(cmds, "--allocate-node-cidrs=true")
		cmds = append(cmds, fmt.Sprintf("--cluster-cidr=%s", r.Obj.Cluster.Spec.ClusterCIDR))
		cmds = append(cmds, fmt.Sprintf("--cluster-name=%s", r.Obj.Cluster.Name))
		cmds = append(cmds, fmt.Sprintf("--node-cidr-mask-size=%d", r.Obj.Cluster.Status.NodeCIDRMaskSize))
	}

	healthPortName := "https-healthz"
	c := corev1.Container{
		Name:            constants.KubeControllerManager,
		Image:           r.Provider.Cfg.KubeAllImageFullName(constants.KubernetesAllImageName, r.Obj.Cluster.Spec.Version),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         cmds,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 10252,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          healthPortName,
				ContainerPort: 10257,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/healthz",
					Port:   intstr.FromString(healthPortName),
					Scheme: corev1.URISchemeHTTPS,
				},
			},

			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			TimeoutSeconds:      15,
			FailureThreshold:    8,
			SuccessThreshold:    1,
		},
		Env: common.ComponentEnv(r.Obj),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.1"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},

		VolumeMounts:             vms,
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}

	containers = append(containers, c)

	deployment := &appsv1.Deployment{
		ObjectMeta: k8sutil.ObjectMeta(constants.KubeControllerManager, constants.KubeControllerManagerLabels, r.Obj.Cluster),
		Spec: appsv1.DeploymentSpec{
			Replicas: k8sutil.IntPointer(3),
			Strategy: common.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: constants.KubeControllerManagerLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: constants.KubeControllerManagerLabels,
				},
				Spec: corev1.PodSpec{
					Containers:  containers,
					Volumes:     volumes,
					Affinity:    common.ComponentAffinity(r.Obj.Cluster.Namespace, constants.KubeApiServerLabels),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}

func (r *Reconciler) schedulerDeployment() runtime.Object {
	containers := []corev1.Container{}
	vms := []corev1.VolumeMount{
		{
			Name:      constants.KubeApiServerConfig,
			MountPath: "/etc/kubernetes/",
		},
	}

	volumes := []corev1.Volume{

		{
			Name: constants.KubeApiServerConfig,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: constants.KubeApiServerConfig,
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
	}

	cmds := []string{
		"kube-scheduler",
		"--authentication-kubeconfig=/etc/kubernetes/scheduler.conf",
		"--authorization-kubeconfig=/etc/kubernetes/scheduler.conf",
		"--bind-address=0.0.0.0",
		"--kubeconfig=/etc/kubernetes/scheduler.conf",
		"--leader-elect=true",
	}

	healthPortName := "https-healthz"
	c := corev1.Container{
		Name:            constants.KubeKubeScheduler,
		Image:           r.Provider.Cfg.KubeAllImageFullName(constants.KubernetesAllImageName, r.Obj.Cluster.Spec.Version),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         cmds,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 10251,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          healthPortName,
				ContainerPort: 10259,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/healthz",
					Port:   intstr.FromString(healthPortName),
					Scheme: corev1.URISchemeHTTPS,
				},
			},

			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			TimeoutSeconds:      15,
			FailureThreshold:    8,
			SuccessThreshold:    1,
		},
		Env: common.ComponentEnv(r.Obj),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.1"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},

		VolumeMounts:             vms,
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}

	containers = append(containers, c)

	deployment := &appsv1.Deployment{
		ObjectMeta: k8sutil.ObjectMeta(constants.KubeKubeScheduler, constants.KubeKubeSchedulerLabels, r.Obj.Cluster),
		Spec: appsv1.DeploymentSpec{
			Replicas: k8sutil.IntPointer(3),
			Strategy: common.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: constants.KubeKubeSchedulerLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: constants.KubeKubeSchedulerLabels,
				},
				Spec: corev1.PodSpec{
					Containers:  containers,
					Volumes:     volumes,
					Affinity:    common.ComponentAffinity(r.Obj.Cluster.Namespace, constants.KubeApiServerLabels),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}
