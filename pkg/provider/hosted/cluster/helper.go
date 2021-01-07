package cluster

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	appsv1 "k8s.io/api/apps/v1"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	Ctx     *common.ClusterContext
	dynamic dynamic.Interface
	*Provider
}

func GetPodBindPort(ctx *common.ClusterContext) int32 {
	var port int32
	port = 6443
	if ctx.Cluster.Spec.Features.HA != nil && ctx.Cluster.Spec.Features.HA.ThirdPartyHA != nil {
		port = ctx.Cluster.Spec.Features.HA.ThirdPartyHA.VPort
	}
	return port
}

func GetSvcNodePort(ctx *common.ClusterContext) int32 {
	port := GetPodBindPort(ctx)

	if port < 30000 {
		port = port + 30000
	}
	return port
}

func GetAdvertiseAddress(ctx *common.ClusterContext) string {
	advertiseAddress := "$(INSTANCE_IP)"
	if ctx.Cluster.Spec.Features.HA != nil && ctx.Cluster.Spec.Features.HA.ThirdPartyHA != nil {
		advertiseAddress = ctx.Cluster.Spec.Features.HA.ThirdPartyHA.VIP
	}

	return advertiseAddress
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

func ApplyCertsConfigmap(ctx *common.ClusterContext, pathCerts map[string][]byte) error {
	noPathCerts := make(map[string]string, len(pathCerts))
	for pathName, value := range pathCerts {
		splits := strings.Split(pathName, "/")
		noPathName := splits[len(splits)-1]
		noPathCerts[noPathName] = string(value)
		ctx.Info("add", "noPathName", noPathName)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: k8sutil.ObjectMeta(constants.KubeApiServerCerts, constants.CtrlLabels, ctx.Cluster),
		Data:       noPathCerts,
	}

	logger := ctx.WithValues("cluster", ctx.Cluster.Name)
	err := k8sutil.Reconcile(logger, ctx.Client, cm, k8sutil.DesiredStatePresent)
	if err != nil {
		return errors.Wrapf(err, "apply certs configmap err: %v", err)
	}
	return nil
}

func ApplyKubeMiscConfigmap(ctx *common.ClusterContext, pathKubeMisc map[string]string) error {
	noPathKubeMisc := make(map[string]string, len(pathKubeMisc))
	for pathName, value := range pathKubeMisc {
		splits := strings.Split(pathName, "/")
		noPathName := splits[len(splits)-1]
		noPathKubeMisc[noPathName] = value
		ctx.Info("add", "noPathName", noPathName)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: k8sutil.ObjectMeta(constants.KubeApiServerConfig, constants.CtrlLabels, ctx.Cluster),
		Data:       noPathKubeMisc,
	}

	logger := ctx.WithValues("cluster", ctx.Cluster.Name)
	err := k8sutil.Reconcile(logger, ctx.Client, cm, k8sutil.DesiredStatePresent)
	if err != nil {
		return errors.Wrapf(err, "apply kube misc configmap err: %v", err)
	}
	return nil
}

func (r *Reconciler) apiServerDeployment() client.Object {
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
					Path: fmt.Sprintf("/web/%s/kube-apiserver/audit", r.Ctx.Cluster.Name),
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
		"--token-auth-file=/etc/kubernetes/known_tokens.csv",
	}

	advertiseAddress := GetAdvertiseAddress(r.Ctx)
	cmds = append(cmds, fmt.Sprintf("--secure-port=%d", GetPodBindPort(r.Ctx)))
	cmds = append(cmds, fmt.Sprintf("--advertise-address=%s", advertiseAddress))
	if r.Ctx.Cluster.Spec.APIServerExtraArgs != nil {
		extraArgs := []string{}
		for k, v := range r.Ctx.Cluster.Spec.APIServerExtraArgs {
			extraArgs = append(extraArgs, fmt.Sprintf("--%s=%s", k, v))
		}
		sort.Strings(extraArgs)
		cmds = append(cmds, extraArgs...)
	}

	svcCidr := "10.96.0.0/16"
	if r.Ctx.Cluster.Spec.ServiceCIDR != nil {
		svcCidr = *r.Ctx.Cluster.Spec.ServiceCIDR
	}

	cmds = append(cmds, fmt.Sprintf("--service-cluster-ip-range=%s", svcCidr))
	if r.Ctx.Cluster.Spec.Etcd != nil && r.Ctx.Cluster.Spec.Etcd.External != nil {
		cmds = append(cmds, fmt.Sprintf("--etcd-servers=%s", strings.Join(r.Ctx.Cluster.Spec.Etcd.External.Endpoints, ",")))
		// tode check
		if strings.Contains(r.Ctx.Cluster.Spec.Etcd.External.Endpoints[0], "https") {
			cmds = append(cmds, fmt.Sprintf("--etcd-cafile=%s", r.Ctx.Cluster.Spec.Etcd.External.CAFile))
			cmds = append(cmds, fmt.Sprintf("--etcd-certfile=%s", r.Ctx.Cluster.Spec.Etcd.External.CertFile))
			cmds = append(cmds, fmt.Sprintf("--etcd-keyfile=%s", r.Ctx.Cluster.Spec.Etcd.External.KeyFile))
		}
	} else {
		cmds = append(cmds, fmt.Sprintf("--etcd-servers=%s", "http://etcd-0.etcd:2379,http://etcd-1.etcd:2379,http://etcd-2.etcd:2379"))
	}

	c := corev1.Container{
		Name:            constants.KubeApiServer,
		Image:           r.Provider.Cfg.KubeAllImageFullName(constants.KubernetesAllImageName, r.Ctx.Cluster.Spec.Version),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         cmds,
		Ports: []corev1.ContainerPort{
			{
				Name:          "https",
				ContainerPort: GetPodBindPort(r.Ctx),
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
		Env: common.ComponentEnv(r.Ctx),
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
		ObjectMeta: k8sutil.ObjectMeta(constants.KubeApiServer, constants.KubeApiServerLabels, r.Ctx.Cluster),
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
					Affinity:    common.ComponentAffinity(r.Ctx.Cluster.Namespace, constants.KubeApiServerLabels),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}

func (r *Reconciler) apiServerSvc() client.Object {
	svc := &corev1.Service{
		ObjectMeta: k8sutil.ObjectMeta(constants.KubeApiServer, constants.KubeApiServerLabels, r.Ctx.Cluster),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       GetPodBindPort(r.Ctx),
					NodePort:   GetSvcNodePort(r.Ctx),
					TargetPort: intstr.FromString("https"),
				},
			},

			Selector: constants.KubeApiServerLabels,
		},
	}
	svcType := corev1.ServiceTypeNodePort
	if constants.GetAnnotationKey(r.Ctx.Cluster.Annotations, constants.ClusterApiSvcType) == string(corev1.ServiceTypeLoadBalancer) {
		svcType = corev1.ServiceTypeLoadBalancer
		svc.Spec.LoadBalancerIP = constants.GetAnnotationKey(r.Ctx.Cluster.Annotations, constants.ClusterApiSvcVip)
	}
	svc.Spec.Type = svcType

	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}

	podPort := GetPodBindPort(r.Ctx)
	svc.Annotations["contour.heptio.com/upstream-protocol.tls"] = fmt.Sprintf("%d,https", podPort)
	return svc
}

func (r *Reconciler) controllerManagerDeployment() client.Object {
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

	if r.Ctx.Cluster.Spec.ControllerManagerExtraArgs != nil {
		extraArgs := []string{}
		for k, v := range r.Ctx.Cluster.Spec.ControllerManagerExtraArgs {
			extraArgs = append(extraArgs, fmt.Sprintf("--%s=%s", k, v))
		}
		sort.Strings(extraArgs)
		cmds = append(cmds, extraArgs...)
	}

	if r.Ctx.Cluster.Status.NodeCIDRMaskSize > 0 {
		cmds = append(cmds, "--allocate-node-cidrs=true")
		cmds = append(cmds, fmt.Sprintf("--cluster-cidr=%s", r.Ctx.Cluster.Spec.ClusterCIDR))
		cmds = append(cmds, fmt.Sprintf("--cluster-name=%s", r.Ctx.Cluster.Name))
		cmds = append(cmds, fmt.Sprintf("--node-cidr-mask-size=%d", r.Ctx.Cluster.Status.NodeCIDRMaskSize))
	}

	healthPortName := "https-healthz"
	c := corev1.Container{
		Name:            constants.KubeControllerManager,
		Image:           r.Provider.Cfg.KubeAllImageFullName(constants.KubernetesAllImageName, r.Ctx.Cluster.Spec.Version),
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
		Env: common.ComponentEnv(r.Ctx),
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
		ObjectMeta: k8sutil.ObjectMeta(constants.KubeControllerManager, constants.KubeControllerManagerLabels, r.Ctx.Cluster),
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
					Affinity:    common.ComponentAffinity(r.Ctx.Cluster.Namespace, constants.KubeApiServerLabels),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}

func (r *Reconciler) schedulerDeployment() client.Object {
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
		Image:           r.Provider.Cfg.KubeAllImageFullName(constants.KubernetesAllImageName, r.Ctx.Cluster.Spec.Version),
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
		Env: common.ComponentEnv(r.Ctx),
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
		ObjectMeta: k8sutil.ObjectMeta(constants.KubeKubeScheduler, constants.KubeKubeSchedulerLabels, r.Ctx.Cluster),
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
					Affinity:    common.ComponentAffinity(r.Ctx.Cluster.Namespace, constants.KubeApiServerLabels),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}
