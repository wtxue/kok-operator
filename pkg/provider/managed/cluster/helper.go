package cluster

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/k8sutil"

	"github.com/pkg/errors"
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

func GenLabels(clusterID, component string) map[string]string {
	return map[string]string{
		"component": component,
		"clusterID": clusterID,
	}
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
		ObjectMeta: k8sutil.ObjectMeta(constants.GenComponentName(ctx.Cluster.GetName(), constants.KubeApiServerCerts), constants.CtrlLabels, ctx.Cluster),
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
		ObjectMeta: k8sutil.ObjectMeta(constants.GenComponentName(ctx.Cluster.GetName(), constants.KubeApiServerConfig), constants.CtrlLabels, ctx.Cluster),
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
	id := r.Ctx.GetClusterID()
	lb := GenLabels(id, constants.KubeApiServer)
	intPort := GetPodBindPort(r.Ctx)
	intstrPort := intstr.FromInt(int(intPort))
	enableHostNetwork := false
	DNSPolicy := corev1.DNSClusterFirst
	if r.Provider.Cfg.EnableHostNetwork {
		enableHostNetwork = true
		DNSPolicy = corev1.DNSClusterFirstWithHostNet
	}

	svcCidr := "10.96.0.0/16"
	if r.Ctx.Cluster.Spec.ServiceCIDR != nil {
		svcCidr = *r.Ctx.Cluster.Spec.ServiceCIDR
	}

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
						Name: constants.GenComponentName(id, constants.KubeApiServerCerts),
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
						Name: constants.GenComponentName(id, constants.KubeApiServerConfig),
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
		{
			Name: constants.KubeApiServerAudit,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: fmt.Sprintf("/var/audit-kube-apiserver/%s", id),
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
		"--tls-cert-file=/etc/kubernetes/pki/apiserver.crt",
		"--tls-private-key-file=/etc/kubernetes/pki/apiserver.key",
		"--token-auth-file=/etc/kubernetes/known_tokens.csv",
		"--service-account-issuer=https://kubernetes.default.svc.cluster.local",
		"--service-account-signing-key-file=/etc/kubernetes/pki/sa.key",
		"--service-account-key-file=/etc/kubernetes/pki/sa.pub",
		"--insecure-port=0",
		"--enable-aggregator-routing=true",
		"--bind-address=0.0.0.0",
	}

	advertiseAddress := GetAdvertiseAddress(r.Ctx)
	cmds = append(cmds, fmt.Sprintf("--secure-port=%d", intPort))
	cmds = append(cmds, fmt.Sprintf("--advertise-address=%s", advertiseAddress))
	if r.Ctx.Cluster.Spec.APIServerExtraArgs != nil {
		extraArgs := []string{}
		for k, v := range r.Ctx.Cluster.Spec.APIServerExtraArgs {
			extraArgs = append(extraArgs, fmt.Sprintf("--%s=%s", k, v))
		}
		sort.Strings(extraArgs)
		cmds = append(cmds, extraArgs...)
	}

	cmds = append(cmds, fmt.Sprintf("--service-cluster-ip-range=%s", svcCidr))
	if r.Ctx.Cluster.Spec.Etcd != nil && r.Ctx.Cluster.Spec.Etcd.External != nil {
		cmds = append(cmds, fmt.Sprintf("--etcd-servers=%s", strings.Join(r.Ctx.Cluster.Spec.Etcd.External.Endpoints, ",")))
		// TODO check
		if strings.Contains(r.Ctx.Cluster.Spec.Etcd.External.Endpoints[0], "https") {
			cmds = append(cmds, fmt.Sprintf("--etcd-cafile=%s", r.Ctx.Cluster.Spec.Etcd.External.CAFile))
			cmds = append(cmds, fmt.Sprintf("--etcd-certfile=%s", r.Ctx.Cluster.Spec.Etcd.External.CertFile))
			cmds = append(cmds, fmt.Sprintf("--etcd-keyfile=%s", r.Ctx.Cluster.Spec.Etcd.External.KeyFile))
		}
	} else {
		cmds = append(cmds, fmt.Sprintf("--etcd-servers=%s", "http://etcd-0.etcd:2379,http://etcd-1.etcd:2379,http://etcd-2.etcd:2379"))
	}

	KubeApiserverContainer := corev1.Container{
		Name:            constants.KubeApiServer,
		Image:           r.Provider.Cfg.KubeAllImageFullName(constants.KubeApiServer, r.Ctx.Cluster.Spec.Version),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         cmds,
		Ports: []corev1.ContainerPort{
			{
				Name:          "https",
				HostPort:      intPort,
				ContainerPort: intPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/readyz",
					Host:   "127.0.0.1",
					Port:   intstrPort,
					Scheme: corev1.URISchemeHTTPS,
				},
			},

			InitialDelaySeconds: 10,
			PeriodSeconds:       10,
			TimeoutSeconds:      15,
			FailureThreshold:    3,
			SuccessThreshold:    1,
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/livez",
					Host:   "127.0.0.1",
					Port:   intstrPort,
					Scheme: corev1.URISchemeHTTPS,
				},
			},

			InitialDelaySeconds: 12,
			PeriodSeconds:       10,
			TimeoutSeconds:      15,
			FailureThreshold:    8,
			SuccessThreshold:    1,
		},
		StartupProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/livez",
					Host:   "127.0.0.1",
					Port:   intstrPort,
					Scheme: corev1.URISchemeHTTPS,
				},
			},

			InitialDelaySeconds: 10,
			PeriodSeconds:       10,
			TimeoutSeconds:      15,
			FailureThreshold:    24,
			SuccessThreshold:    1,
		},
		Env: common.ComponentEnv(r.Ctx),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.1"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},

		VolumeMounts: vms,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: k8sutil.ObjectMeta(constants.GenComponentName(id, constants.KubeApiServer), lb, r.Ctx.Cluster),
		Spec: appsv1.DeploymentSpec{
			Replicas: k8sutil.IntPointer(2),
			Strategy: common.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: lb,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: lb,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						KubeApiserverContainer,
					},
					Volumes:     volumes,
					Affinity:    common.KubeAPIServerAffinity(r.Ctx.GetNamespace(), id, lb),
					Tolerations: common.KubeAPIServerTolerations(id),
					HostNetwork: enableHostNetwork,
					DNSPolicy:   DNSPolicy,
				},
			},
		},
	}

	return deployment
}

func (r *Reconciler) apiServerSvc() client.Object {
	id := r.Ctx.GetClusterID()
	lb := GenLabels(id, constants.KubeApiServer)
	svc := &corev1.Service{
		ObjectMeta: k8sutil.ObjectMeta(constants.GenComponentName(r.Ctx.GetClusterID(), constants.KubeApiServer), lb, r.Ctx.Cluster),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: lb,
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       GetPodBindPort(r.Ctx),
					TargetPort: intstr.FromString("https"),
				},
			},
		},
	}

	if r.Provider.Cfg.EnableHostNetwork {
		svc.Spec.Type = corev1.ServiceTypeClusterIP
	} else {
		svc.Spec.Type = corev1.ServiceTypeNodePort
		svc.Spec.Ports[0].NodePort = GetSvcNodePort(r.Ctx)
	}

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
						Name: constants.GenComponentName(r.Ctx.GetClusterID(), constants.KubeApiServerCerts),
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
						Name: constants.GenComponentName(r.Ctx.GetClusterID(), constants.KubeApiServerConfig),
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
		Image:           r.Provider.Cfg.KubeAllImageFullName(constants.KubeControllerManager, r.Ctx.Cluster.Spec.Version),
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

	id := r.Ctx.GetClusterID()
	lb := GenLabels(id, constants.KubeControllerManager)
	deployment := &appsv1.Deployment{
		ObjectMeta: k8sutil.ObjectMeta(constants.GenComponentName(id, constants.KubeControllerManager), lb, r.Ctx.Cluster),
		Spec: appsv1.DeploymentSpec{
			Replicas: k8sutil.IntPointer(3),
			Strategy: common.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: lb,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: lb,
				},
				Spec: corev1.PodSpec{
					Containers:  containers,
					Volumes:     volumes,
					Affinity:    common.ComponentAffinity(r.Ctx.GetNamespace(), lb),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}

func (r *Reconciler) schedulerDeployment() client.Object {
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
						Name: constants.GenComponentName(r.Ctx.GetClusterID(), constants.KubeApiServerConfig),
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
		Image:           r.Provider.Cfg.KubeAllImageFullName(constants.KubeKubeScheduler, r.Ctx.Cluster.Spec.Version),
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

	containers := []corev1.Container{}
	containers = append(containers, c)

	id := r.Ctx.GetClusterID()
	lb := GenLabels(id, constants.KubeKubeScheduler)
	deployment := &appsv1.Deployment{
		ObjectMeta: k8sutil.ObjectMeta(constants.GenComponentName(id, constants.KubeKubeScheduler), lb, r.Ctx.Cluster),
		Spec: appsv1.DeploymentSpec{
			Replicas: k8sutil.IntPointer(3),
			Strategy: common.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: lb,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: lb,
				},
				Spec: corev1.PodSpec{
					Containers:  containers,
					Volumes:     volumes,
					Affinity:    common.ComponentAffinity(r.Ctx.GetNamespace(), lb),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}
