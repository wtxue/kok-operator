package cluster

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/segmentio/ksuid"
	"github.com/thoas/go-funk"

	bootstraputil "k8s.io/cluster-bootstrap/token/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/phases/kubeadm"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/phases/kubeconfig"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/phases/system"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/preflight"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/apiclient"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/hosts"

	"bytes"

	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/phases/addons/flannel"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/k8sutil"
	"k8s.io/klog"
)

func (p *Provider) EnsureCopyFiles(ctx context.Context, c *provider.Cluster) error {
	for _, file := range c.Spec.Features.Files {
		for _, machine := range c.Spec.Machines {
			machineSSH, err := machine.SSH()
			if err != nil {
				return err
			}

			err = machineSSH.CopyFile(file.Src, file.Dst)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Provider) EnsurePreflight(ctx context.Context, c *provider.Cluster) error {
	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		klog.Infof("start check node: %s ... ", machine.IP)
		err = preflight.RunMasterChecks(machineSSH)
		if err != nil {
			klog.Errorf("node:%s check err: %+v", machine.IP, err)
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureClusterComplete(ctx context.Context, cluster *provider.Cluster) error {
	funcs := []func(cluster *provider.Cluster) error{
		completeK8sVersion,
		completeNetworking,
		completeDNS,
		completeAddresses,
		completeCredential,
	}
	for _, f := range funcs {
		if err := f(cluster); err != nil {
			return err
		}
	}
	return nil
}

func completeK8sVersion(cluster *provider.Cluster) error {
	cluster.Status.Version = cluster.Spec.Version
	return nil
}

func completeNetworking(cluster *provider.Cluster) error {
	var (
		serviceCIDR      string
		nodeCIDRMaskSize int32
		err              error
	)

	if cluster.Spec.ServiceCIDR != nil {
		serviceCIDR = *cluster.Spec.ServiceCIDR
		nodeCIDRMaskSize, err = GetNodeCIDRMaskSize(cluster.Spec.ClusterCIDR, *cluster.Spec.Properties.MaxNodePodNum)
		if err != nil {
			return errors.Wrap(err, "GetNodeCIDRMaskSize error")
		}
	} else {
		serviceCIDR, nodeCIDRMaskSize, err = GetServiceCIDRAndNodeCIDRMaskSize(cluster.Spec.ClusterCIDR, *cluster.Spec.Properties.MaxClusterServiceNum, *cluster.Spec.Properties.MaxNodePodNum)
		if err != nil {
			return errors.Wrap(err, "GetServiceCIDRAndNodeCIDRMaskSize error")
		}
	}
	cluster.Status.ServiceCIDR = serviceCIDR
	cluster.Status.NodeCIDRMaskSize = nodeCIDRMaskSize

	return nil
}

func completeDNS(cluster *provider.Cluster) error {
	ip, err := GetIndexedIP(cluster.Status.ServiceCIDR, constants.DNSIPIndex)
	if err != nil {
		return errors.Wrap(err, "get DNS IP error")
	}
	cluster.Status.DNSIP = ip.String()

	return nil
}

func completeAddresses(cluster *provider.Cluster) error {
	for _, m := range cluster.Spec.Machines {
		cluster.AddAddress(devopsv1.AddressReal, m.IP, 6443)
	}

	if cluster.Spec.Features.HA != nil {
		if cluster.Spec.Features.HA.DKEHA != nil {
			cluster.AddAddress(devopsv1.AddressAdvertise, cluster.Spec.Features.HA.DKEHA.VIP, 6443)
		}
		if cluster.Spec.Features.HA.ThirdPartyHA != nil {
			cluster.AddAddress(devopsv1.AddressAdvertise, cluster.Spec.Features.HA.ThirdPartyHA.VIP, cluster.Spec.Features.HA.ThirdPartyHA.VPort)
		}
	}

	return nil
}

func completeCredential(cluster *provider.Cluster) error {
	token := ksuid.New().String()
	cluster.ClusterCredential.Token = &token

	bootstrapToken, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return err
	}
	cluster.ClusterCredential.BootstrapToken = &bootstrapToken

	certBytes := make([]byte, 32)
	if _, err := rand.Read(certBytes); err != nil {
		return err
	}
	certificateKey := hex.EncodeToString(certBytes)
	cluster.ClusterCredential.CertificateKey = &certificateKey

	return nil
}

func (p *Provider) EnsureKubeconfig(ctx context.Context, c *provider.Cluster) error {
	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		option := &kubeconfig.Option{
			MasterEndpoint: "https://127.0.0.1:6443",
			ClusterName:    c.Name,
			CACert:         c.ClusterCredential.CACert,
			Token:          *c.ClusterCredential.Token,
		}
		err = kubeconfig.Install(machineSSH, option)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsurePrepareForControlplane(ctx context.Context, c *provider.Cluster) error {
	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		c.Status.Version = c.Spec.Version
		klog.Infof("start write toke file to machine: %s", machine.IP)
		tokenData := fmt.Sprintf(tokenFileTemplate, *c.ClusterCredential.Token)
		err = machineSSH.WriteFile(strings.NewReader(tokenData), constants.TokenFile)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}
	return nil
}

func (p *Provider) EnsureKubeadmInitKubeletStartPhase(ctx context.Context, c *provider.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, p.getKubeadmConfig(c),
		fmt.Sprintf("kubelet-start --node-name=%s", c.Spec.Machines[0].IP))
}

func (p *Provider) EnsureKubeadmInitCertsPhase(ctx context.Context, c *provider.Cluster) error {
	cfg := p.getKubeadmConfig(c)
	if p.config.CustomeCert {
		err := kubeadm.InitCustomCerts(cfg, c)
		if err != nil {
			return err
		}
	} else {
		machineSSH, err := c.Spec.Machines[0].SSH()
		if err != nil {
			return err
		}
		klog.Infof("node: %s start init certs by kubeadm", c.Spec.Machines[0].IP)
		return kubeadm.Init(machineSSH, cfg, "certs all")
	}

	return nil
}

func (p *Provider) EnsureKubeadmInitKubeConfigPhase(ctx context.Context, c *provider.Cluster) error {
	cfg := p.getKubeadmConfig(c)
	if p.config.CustomeCert {
		machine := c.Spec.Machines[0]
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}
		err = kubeadm.InitCustomKubeconfig(cfg, machineSSH, c)
		if err != nil {
			return err
		}
	} else {
		machineSSH, err := c.Spec.Machines[0].SSH()
		if err != nil {
			return err
		}
		return kubeadm.Init(machineSSH, cfg, "kubeconfig all")
	}
	return nil
}

func (p *Provider) EnsureKubeadmInitControlPlanePhase(ctx context.Context, c *provider.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}

	return kubeadm.Init(machineSSH, p.getKubeadmConfig(c), "control-plane all")
}

func (p *Provider) EnsureKubeadmInitEtcdPhase(ctx context.Context, c *provider.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, p.getKubeadmConfig(c), "etcd local")
}

func (p *Provider) EnsureKubeadmInitUploadConfigPhase(ctx context.Context, c *provider.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, p.getKubeadmConfig(c), "upload-config all ")
}

func (p *Provider) EnsureKubeadmInitUploadCertsPhase(ctx context.Context, c *provider.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, p.getKubeadmConfig(c), "upload-certs --upload-certs")
}

func (p *Provider) EnsureKubeadmInitBootstrapTokenPhase(ctx context.Context, c *provider.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, p.getKubeadmConfig(c), "bootstrap-token")
}

func (p *Provider) EnsureKubeadmInitAddonPhase(ctx context.Context, c *provider.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, p.getKubeadmConfig(c), "addon all")
}

func (p *Provider) EnsureJoinControlePlane(ctx context.Context, c *provider.Cluster) error {

	option := &kubeadm.JoinControlPlaneOption{
		BootstrapToken:       *c.ClusterCredential.BootstrapToken,
		CertificateKey:       *c.ClusterCredential.CertificateKey,
		ControlPlaneEndpoint: fmt.Sprintf("%s:6443", c.Spec.Machines[0].IP),
	}
	for _, machine := range c.Spec.Machines[1:] {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		option.NodeName = machine.IP
		err = kubeadm.JoinControlPlane(machineSSH, option)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureStoreCredential(ctx context.Context, c *provider.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}

	data, err := machineSSH.ReadFile(constants.CACertName)
	if err != nil {
		return err
	}
	c.ClusterCredential.CACert = data

	data, err = machineSSH.ReadFile(constants.CAKeyName)
	if err != nil {
		return err
	}
	c.ClusterCredential.CAKey = data

	data, err = machineSSH.ReadFile(constants.EtcdCACertName)
	if err != nil {
		return err
	}
	c.ClusterCredential.ETCDCACert = data

	data, err = machineSSH.ReadFile(constants.EtcdCAKeyName)
	if err != nil {
		return err
	}
	c.ClusterCredential.ETCDCAKey = data

	data, err = machineSSH.ReadFile(constants.APIServerEtcdClientCertName)
	if err != nil {
		return err
	}
	c.ClusterCredential.ETCDAPIClientCert = data

	data, err = machineSSH.ReadFile(constants.APIServerEtcdClientKeyName)
	if err != nil {
		return err
	}
	c.ClusterCredential.ETCDAPIClientKey = data

	return nil
}

func (p *Provider) EnsureSystem(ctx context.Context, c *provider.Cluster) error {
	dockerVersion := "19.03.8"
	if v, ok := c.Spec.DockerExtraArgs["version"]; ok {
		dockerVersion = v
	}
	option := &system.Option{
		K8sVersion:    c.Spec.Version,
		DockerVersion: dockerVersion,
		Cgroupdriver:  "systemd", // cgroupfs or systemd
		ExtraArgs:     c.Spec.KubeletExtraArgs,
	}

	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		option.HostIP = machine.IP
		err = system.Install(machineSSH, option)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureKubeadmInitWaitControlPlanePhase(ctx context.Context, c *provider.Cluster) error {
	start := time.Now()

	return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		healthStatus := 0
		clientset, err := c.ClientsetForBootstrap()
		if err != nil {
			log.Warn(err.Error())
			return false, nil
		}

		res := clientset.Discovery().RESTClient().Get().AbsPath("/healthz").Do(ctx)
		res.StatusCode(&healthStatus)
		if healthStatus != http.StatusOK {
			klog.Errorf("Discovery healthz err: %+v", res.Error())
			return false, nil
		}

		log.Infof("All control plane components are healthy after %f seconds\n", time.Since(start).Seconds())
		return true, nil
	})
}

func (p *Provider) EnsureMarkControlPlane(ctx context.Context, c *provider.Cluster) error {
	clientset, err := c.ClientsetForBootstrap()
	if err != nil {
		return err
	}

	for _, machine := range c.Spec.Machines {
		if machine.Labels == nil {
			machine.Labels = make(map[string]string)
		}

		machine.Labels[constants.LabelNodeRoleMaster] = ""
		if !c.Spec.Features.EnableMasterSchedule {
			taint := corev1.Taint{
				Key:    constants.LabelNodeRoleMaster,
				Effect: corev1.TaintEffectNoSchedule,
			}
			if !funk.Contains(machine.Taints, taint) {
				machine.Taints = append(machine.Taints, taint)
			}
		}
		err := apiclient.MarkNode(ctx, clientset, machine.IP, machine.Labels, machine.Taints)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureRegistryHosts(ctx context.Context, c *provider.Cluster) error {
	if !p.config.Registry.NeedSetHosts() {
		return nil
	}

	domains := []string{
		p.config.Registry.Domain,
		c.Spec.TenantID + "." + p.config.Registry.Domain,
	}
	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		for _, one := range domains {
			remoteHosts := &hosts.RemoteHosts{Host: one, SSH: machineSSH}
			err := remoteHosts.Set(p.config.Registry.IP)
			if err != nil {
				return errors.Wrap(err, machine.IP)
			}
		}
	}

	return nil
}

func (p *Provider) EnsurePreInstallHook(ctx context.Context, c *provider.Cluster) error {
	if c.Spec.Features.Hooks == nil {
		return nil
	}

	hook := c.Spec.Features.Hooks[devopsv1.HookPreInstall]
	if hook == "" {
		return nil
	}
	cmd := strings.Split(hook, " ")[0]

	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		machineSSH.Execf("chmod +x %s", cmd)
		_, stderr, exit, err := machineSSH.Exec(hook)
		if err != nil || exit != 0 {
			return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", hook, exit, stderr, err)
		}
	}
	return nil
}

func (p *Provider) EnsurePostInstallHook(ctx context.Context, c *provider.Cluster) error {
	if c.Spec.Features.Hooks == nil {
		return nil
	}

	hook := c.Spec.Features.Hooks[devopsv1.HookPostInstall]
	if hook == "" {
		return nil
	}
	cmd := strings.Split(hook, " ")[0]

	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		machineSSH.Execf("chmod +x %s", cmd)
		_, stderr, exit, err := machineSSH.Exec(hook)
		if err != nil || exit != 0 {
			return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", hook, exit, stderr, err)
		}
	}
	return nil
}

func (p *Provider) EnsureMakeEtcd(ctx context.Context, c *provider.Cluster) error {
	etcdPeerEndpoints := []string{}
	etcdClusterEndpoints := []string{}
	for _, machine := range c.Spec.Machines {
		etcdPeerEndpoints = append(etcdPeerEndpoints, fmt.Sprintf("%s=https://%s:2380", machine.IP, machine.IP))
		etcdClusterEndpoints = append(etcdClusterEndpoints, fmt.Sprintf("https://%s:2379", machine.IP))
	}

	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		etcdByte, err := machineSSH.ReadFile(constants.EtcdPodManifestFile)
		if err != nil {
			return fmt.Errorf("node: %s ReadFile: %s failed error: %v", machine.IP, constants.EtcdPodManifestFile, err)
		}

		etcdObj, err := k8sutil.UnmarshalFromYaml(etcdByte, corev1.SchemeGroupVersion)
		if err != nil {
			return fmt.Errorf("node: %s marshalling %s failed error: %v", machine.IP, constants.EtcdPodManifestFile, err)
		}

		if etcdPod, ok := etcdObj.(*corev1.Pod); ok {
			isFindState := false
			klog.Infof("etcd pod name: %s, cmd: %s", etcdPod.Name, etcdPod.Spec.Containers[0].Command)
			for i, arg := range etcdPod.Spec.Containers[0].Command {
				if strings.HasPrefix(arg, "--initial-cluster=") {
					etcdPod.Spec.Containers[0].Command[i] = fmt.Sprintf("--initial-cluster=%s", strings.Join(etcdPeerEndpoints, ","))
				}
				if strings.HasPrefix(arg, "--initial-cluster-state") {
					isFindState = true
				}
			}

			if isFindState != true {
				etcdPod.Spec.Containers[0].Command = append(etcdPod.Spec.Containers[0].Command, "--initial-cluster-state=existing")
			}
			serialized, err := k8sutil.MarshalToYaml(etcdPod, corev1.SchemeGroupVersion)
			if err != nil {
				return errors.Wrapf(err, "failed to marshal manifest for %q to YAML", etcdPod.Name)
			}

			machineSSH.WriteFile(bytes.NewReader(serialized), constants.EtcdPodManifestFile)
		}

		apiServerByte, err := machineSSH.ReadFile(constants.KubeAPIServerPodManifestFile)
		if err != nil {
			return fmt.Errorf("node: %s ReadFile: %s failed error: %v", machine.IP, constants.EtcdPodManifestFile, err)
		}

		apiServerObj, err := k8sutil.UnmarshalFromYaml(apiServerByte, corev1.SchemeGroupVersion)
		if err != nil {
			return fmt.Errorf("node: %s marshalling %s failed error: %v", machine.IP, constants.EtcdPodManifestFile, err)
		}

		if apiServerPod, ok := apiServerObj.(*corev1.Pod); ok {
			klog.Infof("apiServer pod name: %s, cmd: %s", apiServerPod.Name, apiServerPod.Spec.Containers[0].Command)
			for i, arg := range apiServerPod.Spec.Containers[0].Command {
				if strings.HasPrefix(arg, "--etcd-servers=") {
					apiServerPod.Spec.Containers[0].Command[i] = fmt.Sprintf("--etcd-servers=%s", strings.Join(etcdClusterEndpoints, ","))
					break
				}
			}

			serialized, err := k8sutil.MarshalToYaml(apiServerPod, corev1.SchemeGroupVersion)
			if err != nil {
				return errors.Wrapf(err, "failed to marshal manifest for %q to YAML", apiServerPod.Name)
			}

			machineSSH.WriteFile(bytes.NewReader(serialized), constants.KubeAPIServerPodManifestFile)
		}
	}

	return nil
}

func (p *Provider) EnsureMakeControlPlane(ctx context.Context, c *provider.Cluster) error {
	if !p.config.CustomeCert {
		return nil
	}

	cfg := p.getKubeadmConfig(c)
	for _, machine := range c.Spec.Machines[1:] {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}
		err = kubeadm.InitCustomKubeconfig(cfg, sh, c)
		if err != nil {
			return err
		}

		_, _, _, err = sh.Execf("systemctl restart kubelet")
		if err != nil {
			return err
		}
		err = kubeadm.RestartContainerByFilter(sh, kubeadm.DockerFilterForControlPlane("kube-apiserver"))
		if err != nil {
			return err
		}
		err = kubeadm.RestartContainerByFilter(sh, kubeadm.DockerFilterForControlPlane("kube-controller-manager"))
		if err != nil {
			return err
		}
		err = kubeadm.RestartContainerByFilter(sh, kubeadm.DockerFilterForControlPlane("kube-scheduler"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) EnsureMakeCni(ctx context.Context, c *provider.Cluster) error {
	if c.Spec.Features.Hooks == nil {
		return nil
	}

	hook := c.Spec.Features.Hooks[devopsv1.HookPostCniInstall]
	if hook == "" {
		return nil
	}

	if hook == "flannel" {
		opt := &flannel.Option{
			ClusterPodCidr: c.Spec.ClusterCIDR,
			BackendType:    "vxlan",
		}

		machine := c.Spec.Machines[0]
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		err = flannel.Install(machineSSH, opt)
		if err != nil {
			return err
		}
	}

	return nil
}
