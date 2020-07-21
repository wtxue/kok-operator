package kubeproxy

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/apis"
	kubeproxyv1alpha1 "github.com/wtxue/kok-operator/pkg/apis/kubeproxy/config/v1alpha1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"github.com/wtxue/kok-operator/pkg/provider/phases/certs"
	"github.com/wtxue/kok-operator/pkg/provider/phases/kubemisc"
	"github.com/wtxue/kok-operator/pkg/util/template"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/klog"
)

const (
	// KubeProxyConfigMap19 is the proxy ConfigMap manifest for Kubernetes 1.9 and above
	KubeProxyConfigMap19 = `
kind: ConfigMap
apiVersion: v1
metadata:
  name: {{ .ProxyConfigMap }}
  namespace: kube-system
  labels:
    app: kube-proxy
data:
  kubeconfig.conf: |-
    apiVersion: v1
    kind: Config
    clusters:
    - cluster:
        certificate-authority: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        server: {{ .ControlPlaneEndpoint }}
      name: default
    contexts:
    - context:
        cluster: default
        namespace: default
        user: default
      name: default
    current-context: default
    users:
    - name: default
      user:
        tokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
`

	// KubeProxyDaemonSet19 is the proxy DaemonSet manifest for Kubernetes 1.9 and above
	KubeProxyDaemonSet19 = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: kube-proxy
  name: kube-proxy
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: kube-proxy
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        k8s-app: kube-proxy
    spec:
      priorityClassName: system-node-critical
      containers:
      - name: kube-proxy
        image: {{ .Image }}
        imagePullPolicy: IfNotPresent
        command:
        - /usr/local/bin/kube-proxy
        - --config=/var/lib/kube-proxy/{{ .ProxyConfigMapKey }}
        - --hostname-override=$(NODE_NAME)
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /var/lib/kube-proxy
          name: kube-proxy
        - mountPath: /run/xtables.lock
          name: xtables-lock
          readOnly: false
        - mountPath: /lib/modules
          name: lib-modules
          readOnly: true
        env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
      hostNetwork: true
      serviceAccountName: kube-proxy
      volumes:
      - name: kube-proxy
        configMap:
          name: {{ .ProxyConfigMap }}
      - name: xtables-lock
        hostPath:
          path: /run/xtables.lock
          type: FileOrCreate
      - name: lib-modules
        hostPath:
          path: /lib/modules
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      - operator: Exists
      nodeSelector:
        kubernetes.io/os: linux
`
)

const (
	// KubeProxyClusterRoleName sets the name for the kube-proxy ClusterRole
	// TODO: This k8s-generic, well-known constant should be fetchable from another source, not be in this package
	KubeProxyClusterRoleName = "system:node-proxier"

	// KubeProxyServiceAccountName describes the name of the ServiceAccount for the kube-proxy addon
	KubeProxyServiceAccountName = "kube-proxy"
)

// GetProxyEnvVars builds a list of environment variables in order to use the right proxy
func GetProxyEnvVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{}
	for _, env := range os.Environ() {
		pos := strings.Index(env, "=")
		if pos == -1 {
			// malformed environment variable, skip it.
			continue
		}
		name := env[:pos]
		value := env[pos+1:]
		if strings.HasSuffix(strings.ToLower(name), "_proxy") && value != "" {
			envVar := corev1.EnvVar{Name: name, Value: value}
			envs = append(envs, envVar)
		}
	}
	return envs
}

func getKubeProxyConfiguration(c *common.Cluster) *kubeproxyv1alpha1.KubeProxyConfiguration {
	kubeProxyMode := "iptables"
	if c.Spec.Features.IPVS != nil && *c.Spec.Features.IPVS {
		kubeProxyMode = "ipvs"
	}

	return &kubeproxyv1alpha1.KubeProxyConfiguration{
		BindAddress: "0.0.0.0",
		Mode:        kubeproxyv1alpha1.ProxyMode(kubeProxyMode),
		ClientConnection: componentbaseconfigv1alpha1.ClientConnectionConfiguration{
			Kubeconfig: "/var/lib/kube-proxy/kubeconfig.conf",
		},
	}
}

func kubeproxyMarshal(cfg *kubeproxyv1alpha1.KubeProxyConfiguration) ([]byte, error) {
	gvks, _, err := apis.GetScheme().ObjectKinds(cfg)
	if err != nil {
		klog.Errorf("kubeproxy config get gvks err: %v", err)
		return nil, err
	}

	yamlData, err := apis.MarshalToYAML(cfg, gvks[0].GroupVersion())
	if err != nil {
		klog.Errorf("kubeproxy config Marshal err: %v", err)
		return nil, err
	}

	return yamlData, nil
}

func BuildKubeproxyAddon(cfg *config.Config, c *common.Cluster) ([]runtime.Object, error) {
	objs := make([]runtime.Object, 0)

	kubeproxyBytes, err := kubeproxyMarshal(getKubeProxyConfiguration(c))
	if err != nil {
		return nil, errors.Wrap(err, "error when kubeproxyMarshal")
	}
	apiserver := certs.BuildApiserverEndpoint(c.Cluster.Spec.PublicAlternativeNames[0], kubemisc.GetBindPort(c.Cluster))
	proxyConfigMapBytes, err := template.ParseString(KubeProxyConfigMap19,
		struct {
			ControlPlaneEndpoint string
			ProxyConfig          string
			ProxyConfigMap       string
			ProxyConfigMapKey    string
		}{
			ControlPlaneEndpoint: apiserver,
			ProxyConfigMap:       constants.KubeProxyConfigMap,
		})
	if err != nil {
		return nil, errors.Wrap(err, "error when parsing kube-proxy configmap template")
	}
	kubeproxyConfigMap := &corev1.ConfigMap{}
	if err := runtime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), proxyConfigMapBytes, kubeproxyConfigMap); err != nil {
		return nil, errors.Wrap(err, "unable to decode kube-proxy configmap")
	}
	if kubeproxyConfigMap.Data == nil {
		kubeproxyConfigMap.Data = map[string]string{}
	}
	kubeproxyConfigMap.Data[constants.KubeProxyConfigMapKey] = string(kubeproxyBytes)
	objs = append(objs, kubeproxyConfigMap)

	proxyDaemonSetBytes, err := template.ParseString(KubeProxyDaemonSet19, struct{ Image, ProxyConfigMap, ProxyConfigMapKey string }{
		Image:             cfg.KubeProxyImagesName(c.Cluster.Spec.Version),
		ProxyConfigMap:    constants.KubeProxyConfigMap,
		ProxyConfigMapKey: constants.KubeProxyConfigMapKey,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error when parsing kube-proxy daemonset template")
	}

	kubeproxyDaemonSet := &appsv1.DaemonSet{}
	if err := runtime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), proxyDaemonSetBytes, kubeproxyDaemonSet); err != nil {
		return nil, errors.Wrap(err, "unable to decode kube-proxy daemonset")
	}

	kubeproxyDaemonSet.Spec.Template.Spec.HostAliases = []corev1.HostAlias{
		{
			IP:        c.Cluster.Spec.Features.HA.ThirdPartyHA.VIP,
			Hostnames: []string{c.Cluster.Spec.PublicAlternativeNames[0]},
		},
	}

	objs = append(objs, kubeproxyDaemonSet)

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KubeProxyServiceAccountName,
			Namespace: metav1.NamespaceSystem,
		},
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeadm:node-proxier",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     KubeProxyClusterRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      KubeProxyServiceAccountName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KubeProxyConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get"},
				APIGroups:     []string{""},
				Resources:     []string{"configmaps"},
				ResourceNames: []string{constants.KubeProxyConfigMap},
			},
		},
	}

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KubeProxyConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     constants.KubeProxyConfigMap,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     rbacv1.GroupKind,
				APIGroup: rbacv1.GroupName,
				Name:     constants.NodeBootstrapTokenAuthGroup,
			},
		},
	}
	objs = append(objs, sa)
	objs = append(objs, crb)
	objs = append(objs, role)
	objs = append(objs, rb)
	return objs, nil
}
