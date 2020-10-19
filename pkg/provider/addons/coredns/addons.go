package coredns

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/provider/config"
	"github.com/wtxue/kok-operator/pkg/util/template"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	// CoreDNSDeployment is the CoreDNS Deployment manifest
	CoreDNSDeployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .DeploymentName }}
  namespace: kube-system
  labels:
    k8s-app: kube-dns
spec:
  replicas: {{ .Replicas }}
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  selector:
    matchLabels:
      k8s-app: kube-dns
  template:
    metadata:
      labels:
        k8s-app: kube-dns
    spec:
      priorityClassName: system-cluster-critical
      serviceAccountName: coredns
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      - key: {{ .ControlPlaneTaintKey }}
        effect: NoSchedule
      nodeSelector:
        kubernetes.io/os: linux
      containers:
      - name: coredns
        image: {{ .Image }}
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            memory: 170Mi
          requests:
            cpu: 100m
            memory: 70Mi
        args: [ "-conf", "/etc/coredns/Corefile" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
          readOnly: true
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        - containerPort: 9153
          name: metrics
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        readinessProbe:
          httpGet:
            path: /ready
            port: 8181
            scheme: HTTP
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - all
          readOnlyRootFilesystem: true
      dnsPolicy: Default
      volumes:
        - name: config-volume
          configMap:
            name: coredns
            items:
            - key: Corefile
              path: Corefile
`

	// CoreDNSConfigMap is the CoreDNS ConfigMap manifest
	CoreDNSConfigMap = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health {
           lameduck 5s
        }
        ready
        kubernetes {{ .DNSDomain }} in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }{{ .Federation }}
        prometheus :9153
        forward . {{ .UpstreamNameserver }}
        cache 30
        loop
        reload
        loadbalance
    }{{ .StubDomain }}
`
	// CoreDNSClusterRole is the CoreDNS ClusterRole manifest
	CoreDNSClusterRole = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:coredns
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - services
  - pods
  - namespaces
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
`
	// CoreDNSClusterRoleBinding is the CoreDNS Clusterrolebinding manifest
	CoreDNSClusterRoleBinding = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:coredns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:coredns
subjects:
- kind: ServiceAccount
  name: coredns
  namespace: kube-system
`
	// CoreDNSServiceAccount is the CoreDNS ServiceAccount manifest
	CoreDNSServiceAccount = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: coredns
  namespace: kube-system
`

	// KubeDNSService is the kube-dns Service manifest
	KubeDNSService = `
apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "KubeDNS"
  name: kube-dns
  namespace: kube-system
  annotations:
    prometheus.io/port: "9153"
    prometheus.io/scrape: "true"
spec:
  clusterIP: {{ .DNSIP }}
  ports:
  - name: dns
    port: 53
    protocol: UDP
    targetPort: 53
  - name: dns-tcp
    port: 53
    protocol: TCP
    targetPort: 53
  - name: metrics
    port: 9153
    protocol: TCP
    targetPort: 9153
  selector:
    k8s-app: kube-dns
`
)

const (
	unableToDecodeCoreDNS = "unable to decode CoreDNS"
	coreDNSReplicas       = 2
)

func BuildCoreDNSAddon(cfg *config.Config, c *common.Cluster) ([]runtime.Object, error) {
	objs := make([]runtime.Object, 0)

	coreDNSDeploymentBytes, err := template.ParseString(CoreDNSDeployment, struct {
		DeploymentName, Image, ControlPlaneTaintKey string
		Replicas                                    string
	}{
		DeploymentName:       constants.CoreDNSDeploymentName,
		Image:                constants.GetGenericImage(cfg.Registry.Prefix, constants.CoreDNSImageName, constants.CoreDNSVersion),
		ControlPlaneTaintKey: constants.LabelNodeRoleMaster,
		Replicas:             fmt.Sprintf("%d", coreDNSReplicas),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error when parsing CoreDNS deployment template")
	}

	coreDNSDeployment := &appsv1.Deployment{}
	if err := runtime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), coreDNSDeploymentBytes, coreDNSDeployment); err != nil {
		return nil, errors.Wrapf(err, "%s Deployment", unableToDecodeCoreDNS)
	}

	objs = append(objs, coreDNSDeployment)

	// Get the config file for CoreDNS
	coreDNSConfigMapBytes, err := template.ParseString(CoreDNSConfigMap, struct{ DNSDomain, UpstreamNameserver, Federation, StubDomain string }{
		DNSDomain:          c.Cluster.Spec.DNSDomain,
		UpstreamNameserver: "/etc/resolv.conf",
		Federation:         "",
		StubDomain:         "",
	})
	if err != nil {
		return nil, errors.Wrap(err, "error when parsing CoreDNS configMap template")
	}

	coreDNSConfigMap := &corev1.ConfigMap{}
	if err := runtime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), coreDNSConfigMapBytes, coreDNSConfigMap); err != nil {
		return nil, errors.Wrapf(err, "%s ConfigMap", unableToDecodeCoreDNS)
	}

	objs = append(objs, coreDNSConfigMap)
	coreDNSServiceBytes, err := template.ParseString(KubeDNSService, struct{ DNSIP string }{
		DNSIP: c.Cluster.Status.DNSIP,
	})

	if err != nil {
		return nil, errors.Wrap(err, "error when parsing CoreDNS service template")
	}

	dnsService := &corev1.Service{}
	if err := runtime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), coreDNSServiceBytes, dnsService); err != nil {
		return nil, errors.Wrap(err, "unable to decode the DNS service")
	}

	objs = append(objs, dnsService)

	coreDNSClusterRoles := &rbacv1.ClusterRole{}
	if err := runtime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), []byte(CoreDNSClusterRole), coreDNSClusterRoles); err != nil {
		return nil, errors.Wrapf(err, "%s ClusterRole", unableToDecodeCoreDNS)
	}

	objs = append(objs, coreDNSClusterRoles)
	coreDNSClusterRolesBinding := &rbacv1.ClusterRoleBinding{}
	if err := runtime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), []byte(CoreDNSClusterRoleBinding), coreDNSClusterRolesBinding); err != nil {
		return nil, errors.Wrapf(err, "%s ClusterRoleBinding", unableToDecodeCoreDNS)
	}

	objs = append(objs, coreDNSClusterRolesBinding)
	coreDNSServiceAccount := &corev1.ServiceAccount{}
	if err := runtime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), []byte(CoreDNSServiceAccount), coreDNSServiceAccount); err != nil {
		return nil, errors.Wrapf(err, "%s ServiceAccount", unableToDecodeCoreDNS)
	}

	objs = append(objs, coreDNSServiceAccount)
	return objs, nil
}
