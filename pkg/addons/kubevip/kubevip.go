package kubevip

import (
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/util/template"
)

const (
	kubeVipTemplate = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: null
  name: kube-vip
  namespace: kube-system
spec:
  containers:
    - args:
        - start
      env:
        - name: vip_arp
          value: "true"
        - name: vip_interface
          value: "{{ default "eth0" .EthInterface }}"
        - name: vip_leaderelection
          value: "true"
        - name: vip_leaseduration
          value: "5"
        - name: vip_renewdeadline
          value: "3"
        - name: vip_retryperiod
          value: "1"
        - name: vip_address
          value: "{{ default "172.16.18.243" .Vip }}"
      image: "{{ default "plndr/kube-vip:0.2.3" .ImageName }}"
      imagePullPolicy: Always
      name: kube-vip
      resources: {}
      volumeMounts:
        - mountPath: /etc/kubernetes/
          name: kubeconfig
          readOnly: true
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
            - SYS_TIME
  hostNetwork: true
  dnsPolicy: ClusterFirstWithHostNet
  volumes:
    - hostPath:
        path: /etc/kubernetes/
        type: DirectoryOrCreate
      name: kubeconfig
status: {}
`
)

type Option struct {
	EthInterface string
	Vip          string
	ImageName    string
}

func BuildKubeVipStaticPod(ctx *common.ClusterContext) map[string]string {
	vip := ""
	if ctx.Cluster.Spec.Features.HA != nil {
		if ctx.Cluster.Spec.Features.HA.KubeHA != nil {
			vip = ctx.Cluster.Spec.Features.HA.KubeHA.VIP
		}
		if ctx.Cluster.Spec.Features.HA.ThirdPartyHA != nil {
			vip = ctx.Cluster.Spec.Features.HA.ThirdPartyHA.VIP
		}
	}

	opt := &Option{
		EthInterface: ctx.Cluster.Spec.NetworkDevice,
		Vip:          vip,
		ImageName:    "",
	}

	data, err := template.ParseString(kubeVipTemplate, opt)
	if err != nil {
		return nil
	}

	staticPodMap := map[string]string{
		"kube-vip.yaml": string(data),
	}

	return staticPodMap
}
