package machine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pkg/errors"
	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/common"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/addons/cni"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/phases/certs"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/phases/component"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/phases/joinnode"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/phases/kubemisc"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/phases/system"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/preflight"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/apiclient"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/hosts"
	"k8s.io/klog"
)

func (p *Provider) EnsureCopyFiles(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	for _, file := range c.Spec.Features.Files {
		err = system.CopyFile(machineSSH, &file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) EnsurePreInstallHook(ctx context.Context, machine *devopsv1.Machine, cluster *common.Cluster) error {
	hook := cluster.Spec.Features.Hooks[devopsv1.HookPreInstall]
	if hook == "" {
		return nil
	}

	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	cmd := strings.Split(hook, " ")[0]

	machineSSH.Execf("chmod +x %s", cmd)
	_, stderr, exit, err := machineSSH.Exec(hook)
	if err != nil || exit != 0 {
		return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", hook, exit, stderr, err)
	}
	return nil
}

func (p *Provider) EnsurePostInstallHook(ctx context.Context, machine *devopsv1.Machine, cluster *common.Cluster) error {
	hook := cluster.Spec.Features.Hooks[devopsv1.HookPostInstall]
	if hook == "" {
		return nil
	}

	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	cmd := strings.Split(hook, " ")[0]

	machineSSH.Execf("chmod +x %s", cmd)
	_, stderr, exit, err := machineSSH.Exec(hook)
	if err != nil || exit != 0 {
		return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", hook, exit, stderr, err)
	}
	return nil
}

func (p *Provider) EnsureClean(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	_, err = machineSSH.CombinedOutput(fmt.Sprintf("rm -rf %s", constants.KubernetesDir))
	if err != nil {
		return err
	}

	_, err = machineSSH.CombinedOutput(fmt.Sprintf("rm -rf %s", constants.CNIConfDIr))
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) EnsurePreflight(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	err = preflight.RunNodeChecks(machineSSH)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) EnsureRegistryHosts(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	var vip string
	vipNodeKey := constants.GetAnnotationKey(machine.Annotations, constants.ClusterApiSvcVip)
	vipMasterKey := constants.GetAnnotationKey(c.Cluster.Annotations, constants.ClusterApiSvcVip)
	if vipMasterKey != "" {
		vip = vipMasterKey
	} else {
		if len(c.Cluster.Spec.Machines) == 0 {
			return fmt.Errorf("cluster: %s no vip and machines", c.Cluster.Name)
		}

		vip = c.Cluster.Spec.Machines[0].IP
	}

	if vipNodeKey != "" && vipNodeKey == vip {
		return nil
	}

	sh, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	domains := []string{
		c.Cluster.Spec.PublicAlternativeNames[0],
	}

	for _, one := range domains {
		remoteHosts := hosts.RemoteHosts{Host: one, SSH: sh}
		err := remoteHosts.Set(vip)
		if err != nil {
			return err
		}
	}

	klog.Infof("node: %s set vip: %s, domains: %+v", sh.HostIP(), vip, domains)
	if machine.Annotations == nil {
		machine.Annotations = map[string]string{}
	}

	machine.Annotations[constants.ClusterApiSvcVip] = vip
	err = c.Client.Update(context.TODO(), machine)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) EnsureSystem(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	sh, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	err = system.Install(sh, c)
	if err != nil {
		return errors.Wrap(err, sh.HostIP())
	}

	return nil
}

func (p *Provider) EnsureK8sComponent(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	sh, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	err = component.Install(sh, c)
	if err != nil {
		return errors.Wrap(err, sh.HostIP())
	}

	return nil
}

func (p *Provider) EnsureKubeconfig(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	apiserver := certs.BuildApiserverEndpoint(c.Cluster.Spec.PublicAlternativeNames[0], kubemisc.GetBindPort(c.Cluster))
	klog.Infof("join apiserver: %s", apiserver)

	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	option := &kubemisc.Option{
		MasterEndpoint: apiserver,
		ClusterName:    c.Name,
		CACert:         c.ClusterCredential.CACert,
		Token:          *c.ClusterCredential.Token,
	}
	err = kubemisc.InstallNode(machineSSH, option)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) EnsureJoinNode(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	sh, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	apiserver := certs.BuildApiserverEndpoint(c.Cluster.Spec.PublicAlternativeNames[0], kubemisc.GetBindPort(c.Cluster))
	klog.Infof("join apiserver: %s", apiserver)

	err = joinnode.JoinNodePhase(sh, p.Cfg, c, apiserver, false)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) EnsureMarkNode(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	clusterCtx, err := c.ClusterManager.Get(c.Name)
	if err != nil {
		return nil
	}

	err = apiclient.MarkNode(ctx, clusterCtx.KubeCli, machine.Spec.Machine.IP, machine.Spec.Machine.Labels, machine.Spec.Machine.Taints)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) EnsureEth(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	var cniType string
	var ok bool

	if cniType, ok = c.Cluster.Spec.Features.Hooks[devopsv1.HookCniInstall]; !ok {
		return nil
	}

	if cniType != "dke-cni" {
		return nil
	}

	sh, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	err = cni.ApplyEth(sh, c)
	if err != nil {
		klog.Errorf("node: %s apply eth err: %v", sh.HostIP(), err)
		return err
	}

	return nil
}

func (p *Provider) EnsureCni(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	var cniType string
	var ok bool

	if cniType, ok = c.Cluster.Spec.Features.Hooks[devopsv1.HookCniInstall]; !ok {
		return nil
	}

	if cniType != "dke-cni" {
		return nil
	}

	sh, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	err = cni.ApplyCniCfg(sh, c)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) EnsureNodeReady(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	clusterCtx, err := c.ClusterManager.Get(c.Name)
	if err != nil {
		return nil
	}

	return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		node, err := clusterCtx.KubeCli.CoreV1().Nodes().Get(ctx, machine.Spec.Machine.IP, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		for _, one := range node.Status.Conditions {
			if one.Type == corev1.NodeReady && one.Status == corev1.ConditionTrue {
				return true, nil
			}
		}

		return false, nil
	})
}

func GetMasterEndpoint(addresses []devopsv1.ClusterAddress) (string, error) {
	var advertise, internal []*devopsv1.ClusterAddress
	for _, one := range addresses {
		if one.Type == devopsv1.AddressAdvertise {
			advertise = append(advertise, &one)
		}
		if one.Type == devopsv1.AddressReal {
			internal = append(internal, &one)
		}
	}

	var address *devopsv1.ClusterAddress
	if advertise != nil {
		address = advertise[rand.Intn(len(advertise))]
	} else {
		if internal != nil {
			address = internal[rand.Intn(len(internal))]
		}
	}
	if address == nil {
		return "", errors.New("no advertise or internal address for the cluster")
	}

	return fmt.Sprintf("https://%s:%d", address.Host, address.Port), nil
}
