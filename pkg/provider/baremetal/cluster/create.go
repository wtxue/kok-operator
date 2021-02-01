package cluster

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/thoas/go-funk"
	"github.com/wtxue/kok-operator/pkg/addons/flannel"
	"github.com/wtxue/kok-operator/pkg/addons/metricsserver"
	"github.com/wtxue/kok-operator/pkg/addons/rawcni"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	"github.com/wtxue/kok-operator/pkg/provider/phases/certs"
	"github.com/wtxue/kok-operator/pkg/provider/phases/cri"
	"github.com/wtxue/kok-operator/pkg/provider/phases/kubeadm"
	"github.com/wtxue/kok-operator/pkg/provider/phases/kubebin"
	"github.com/wtxue/kok-operator/pkg/provider/phases/kubemisc"
	"github.com/wtxue/kok-operator/pkg/provider/phases/system"
	"github.com/wtxue/kok-operator/pkg/provider/preflight"
	"github.com/wtxue/kok-operator/pkg/util/apiclient"
	"github.com/wtxue/kok-operator/pkg/util/hosts"
	"github.com/wtxue/kok-operator/pkg/util/pkiutil"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
)

func (p *Provider) EnsureCopyFiles(ctx *common.ClusterContext) error {
	for _, file := range ctx.Cluster.Spec.Features.Files {
		for _, machine := range ctx.Cluster.Spec.Machines {
			machineSSH, err := machine.SSH()
			if err != nil {
				return err
			}

			err = system.CopyFile(machineSSH, &file)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Provider) EnsurePreflight(ctx *common.ClusterContext) error {
	for _, m := range ctx.Cluster.Spec.Machines {
		sh, err := m.SSH()
		if err != nil {
			return err
		}

		ctx.Info("node preflight start ...", "node", m.IP)
		err = preflight.RunMasterChecks(ctx, sh)
		if err != nil {
			ctx.Error(err, "node preflight", "node", m.IP)
			return errors.Wrap(err, m.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureClusterComplete(ctx *common.ClusterContext) error {
	funcs := []func(ctx *common.ClusterContext) error{
		completeK8sVersion,
		completeNetworking,
		completeDNS,
		completeAddresses,
		completeCredential,
	}
	for _, f := range funcs {
		if err := f(ctx); err != nil {
			return err
		}
	}

	ctx.Cluster.Status.Version = ctx.Cluster.Spec.Version
	return nil
}

func completeK8sVersion(ctx *common.ClusterContext) error {
	ctx.Cluster.Status.Version = ctx.Cluster.Spec.Version
	return nil
}

func completeNetworking(ctx *common.ClusterContext) error {
	var (
		serviceCIDR      string
		nodeCIDRMaskSize int32
		err              error
	)

	if ctx.Cluster.Spec.ServiceCIDR != nil {
		serviceCIDR = *ctx.Cluster.Spec.ServiceCIDR
		nodeCIDRMaskSize, err = k8sutil.GetNodeCIDRMaskSize(ctx.Cluster.Spec.ClusterCIDR, *ctx.Cluster.Spec.Properties.MaxNodePodNum)
		if err != nil {
			return errors.Wrap(err, "GetNodeCIDRMaskSize error")
		}
	} else {
		serviceCIDR, nodeCIDRMaskSize, err = k8sutil.GetServiceCIDRAndNodeCIDRMaskSize(ctx.Cluster.Spec.ClusterCIDR, *ctx.Cluster.Spec.Properties.MaxClusterServiceNum, *ctx.Cluster.Spec.Properties.MaxNodePodNum)
		if err != nil {
			return errors.Wrap(err, "GetServiceCIDRAndNodeCIDRMaskSize error")
		}
	}
	ctx.Cluster.Status.ServiceCIDR = serviceCIDR
	ctx.Cluster.Status.NodeCIDRMaskSize = nodeCIDRMaskSize

	return nil
}

func completeDNS(ctx *common.ClusterContext) error {
	ip, err := k8sutil.GetIndexedIP(ctx.Cluster.Status.ServiceCIDR, constants.DNSIPIndex)
	if err != nil {
		return errors.Wrap(err, "get DNS IP error")
	}
	ctx.Cluster.Status.DNSIP = ip.String()

	return nil
}

func completeAddresses(ctx *common.ClusterContext) error {
	for _, m := range ctx.Cluster.Spec.Machines {
		ctx.Cluster.AddAddress(devopsv1.AddressReal, m.IP, 6443)
	}

	if ctx.Cluster.Spec.Features.HA != nil {
		if ctx.Cluster.Spec.Features.HA.KubeHA != nil {
			ctx.Cluster.AddAddress(devopsv1.AddressAdvertise, ctx.Cluster.Spec.Features.HA.KubeHA.VIP, 6443)
		}
		if ctx.Cluster.Spec.Features.HA.ThirdPartyHA != nil {
			ctx.Cluster.AddAddress(devopsv1.AddressAdvertise, ctx.Cluster.Spec.Features.HA.ThirdPartyHA.VIP, ctx.Cluster.Spec.Features.HA.ThirdPartyHA.VPort)
		}
	}

	return nil
}

func completeCredential(ctx *common.ClusterContext) error {
	token := ksuid.New().String()
	ctx.Credential.Token = &token

	bootstrapToken, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return err
	}

	ctx.Credential.BootstrapToken = &bootstrapToken
	certBytes := make([]byte, 32)
	if _, err := rand.Read(certBytes); err != nil {
		return err
	}

	certificateKey := hex.EncodeToString(certBytes)
	ctx.Credential.CertificateKey = &certificateKey
	return nil
}

func (p *Provider) EnsureBuildLocalKubeconfig(ctx *common.ClusterContext) error {
	for _, machine := range ctx.Cluster.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		err = kubemisc.Install(ctx, sh)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureKubeadmInitKubeletStartPhase(ctx *common.ClusterContext) error {
	sh, err := ctx.Cluster.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(ctx, sh, kubeadm.GetKubeadmConfigByMaster0(ctx, p.Cfg),
		fmt.Sprintf("kubelet-start --node-name=%s", ctx.Cluster.Spec.Machines[0].IP))
}

func (p *Provider) EnsureImagesPull(ctx *common.ClusterContext) error {
	for _, machine := range ctx.Cluster.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		err = kubeadm.ImagesPull(ctx, sh, ctx.Cluster.Spec.Version, p.Cfg.Registry.Prefix)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Provider) EnsureCerts(ctx *common.ClusterContext) error {
	cfg := kubeadm.GetKubeadmConfigByMaster0(ctx, p.Cfg)
	err := kubeadm.InitCerts(ctx, cfg, false)
	if err != nil {
		return err
	}

	for _, machine := range ctx.Cluster.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		for pathFile, va := range ctx.Credential.CertsBinaryData {
			ctx.Info("write certs binaryData", "node", sh.HostIP(), "file", pathFile)
			err = sh.WriteFile(bytes.NewReader(va), pathFile)
			if err != nil {
				ctx.Error(err, "write certs binaryData", "node", sh.HostIP(), "file", pathFile)
				return err
			}
		}
	}

	return nil
}

func (p *Provider) EnsureKubeMiscPhase(ctx *common.ClusterContext) error {
	sh, err := ctx.Cluster.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}

	apiserver := certs.BuildApiserverEndpoint(ctx.Cluster.Spec.Machines[0].IP, 6443)
	kubeMaps := make(map[string]string)
	err = kubemisc.ApplyKubeletKubeconfig(ctx, apiserver, sh.HostIP(), kubeMaps)
	if err != nil {
		return err
	}

	err = kubemisc.BuildMasterMiscConfigToMap(ctx, apiserver)
	if err != nil {
		return err
	}

	for k, v := range ctx.Credential.KubeData {
		kubeMaps[k] = v
	}

	for pathName, va := range kubeMaps {
		ctx.Info("write misc config", "node", sh.HostIP(), "file", pathName)
		err = sh.WriteFile(strings.NewReader(va), pathName)
		if err != nil {
			ctx.Error(err, "write misc config", "node", sh.HostIP(), "file", pathName)
			return err
		}
	}

	return nil
}

func (p *Provider) EnsureKubeadmInitControlPlanePhase(ctx *common.ClusterContext) error {
	sh, err := ctx.Cluster.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}

	return kubeadm.Init(ctx, sh, kubeadm.GetKubeadmConfigByMaster0(ctx, p.Cfg), "control-plane all")
}

func (p *Provider) EnsureKubeadmInitEtcdPhase(ctx *common.ClusterContext) error {
	sh, err := ctx.Cluster.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	err = kubeadm.Init(ctx, sh, kubeadm.GetKubeadmConfigByMaster0(ctx, p.Cfg), "etcd local")
	if err != nil {
		return err
	}

	if !p.Cfg.EnableCustomImages {
		return nil
	}

	err = kubeadm.RebuildMasterManifestFile(ctx, sh, p.Cfg)
	if err != nil {
		ctx.Error(err, "modify some master config", "node", sh.HostIP())
		return err
	}

	return nil
}

func (p *Provider) EnsureKubeadmInitUploadConfigPhase(ctx *common.ClusterContext) error {
	sh, err := ctx.Cluster.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(ctx, sh, kubeadm.GetKubeadmConfigByMaster0(ctx, p.Cfg), "upload-config all")
}

func (p *Provider) EnsureKubeadmInitUploadCertsPhase(ctx *common.ClusterContext) error {
	sh, err := ctx.Cluster.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(ctx, sh, kubeadm.GetKubeadmConfigByMaster0(ctx, p.Cfg), "upload-certs --upload-certs")
}

func (p *Provider) EnsureKubeadmInitBootstrapTokenPhase(ctx *common.ClusterContext) error {
	sh, err := ctx.Cluster.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(ctx, sh, kubeadm.GetKubeadmConfigByMaster0(ctx, p.Cfg), "bootstrap-token")
}

func (p *Provider) EnsureKubeadmInitAddonPhase(ctx *common.ClusterContext) error {
	sh, err := ctx.Cluster.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(ctx, sh, kubeadm.GetKubeadmConfigByMaster0(ctx, p.Cfg), "addon all")
}

func (p *Provider) EnsureJoinControlePlane(ctx *common.ClusterContext) error {
	if len(ctx.Cluster.Spec.Machines) <= 1 {
		return nil
	}

	for _, machine := range ctx.Cluster.Spec.Machines[1:] {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		clientset, err := ctx.ClientsetForBootstrap()
		if err != nil {
			ctx.Error(err, "ClientsetForBootstrap", "node", sh.HostIP())
			return err
		}

		_, err = clientset.CoreV1().Nodes().Get(context.TODO(), sh.HostIP(), metav1.GetOptions{})
		if err == nil {
			return nil
		}

		// apiserver := certs.BuildApiserverEndpoint(c.Spec.Machines[0].IP, 6443)
		// err = joinNode.JoinNodePhase(sh, p.Cfg, c, apiserver, true)
		// if err != nil {
		// 	return errors.Wrapf(err, "node: %s JoinNodePhase", sh.HostIP())
		// }

		err = kubeadm.JoinControlPlane(ctx, sh)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}

		if !p.Cfg.EnableCustomImages {
			continue
		}

		err = kubeadm.RebuildMasterManifestFile(ctx, sh, p.Cfg)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureK8sComponent(ctx *common.ClusterContext) error {
	for _, machine := range ctx.Cluster.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		err = kubebin.Install(ctx, sh)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureSystem(ctx *common.ClusterContext) error {
	wg := sync.WaitGroup{}
	quitErrors := make(chan error)
	wgDone := make(chan struct{})
	for _, mach := range ctx.Cluster.Spec.Machines {
		sh, err := mach.SSH()
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(sh ssh.Interface) {
			defer wg.Done()
			err = system.Install(ctx, sh)
			if err != nil {
				quitErrors <- errors.Wrap(err, sh.HostIP())
			}
		}(sh)
	}

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received through the channel
	select {
	case <-wgDone:
		break
	case err := <-quitErrors:
		close(quitErrors)
		ctx.Error(err, "ensure system")
		return err
	}

	ctx.Info("ensure system all master node executed successfully")
	return nil
}

func (p *Provider) EnsureCRI(ctx *common.ClusterContext) error {
	for _, machine := range ctx.Cluster.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		err = cri.InstallCRI(ctx, sh)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureKubeadmInitWaitControlPlanePhase(ctx *common.ClusterContext) error {
	start := time.Now()
	return wait.PollImmediate(5*time.Second, 30*time.Second, func() (bool, error) {
		healthStatus := 0
		clientset, err := ctx.ClientsetForBootstrap()
		if err != nil {
			ctx.Error(err, "ClientsetForBootstrap")
			return false, nil
		}

		res := clientset.Discovery().RESTClient().Get().AbsPath("/healthz").Do(ctx.Ctx)
		res.StatusCode(&healthStatus)
		if healthStatus != http.StatusOK {
			ctx.Error(res.Error(), "Discovery healthz")
			return false, nil
		}

		ctx.Info("All control plane components are healthy", "after seconds", fmt.Sprintf("%f", time.Since(start).Seconds()))
		return true, nil
	})
}

func (p *Provider) EnsureMarkControlPlane(ctx *common.ClusterContext) error {
	clientset, err := ctx.ClientsetForBootstrap()
	if err != nil {
		return err
	}

	for _, machine := range ctx.Cluster.Spec.Machines {
		if machine.Labels == nil {
			machine.Labels = make(map[string]string)
		}

		machine.Labels[constants.LabelNodeRoleMaster] = ""
		if !ctx.Cluster.Spec.Features.EnableMasterSchedule {
			taint := corev1.Taint{
				Key:    constants.LabelNodeRoleMaster,
				Effect: corev1.TaintEffectNoSchedule,
			}
			if !funk.Contains(machine.Taints, taint) {
				machine.Taints = append(machine.Taints, taint)
			}
		}
		err := apiclient.MarkNode(ctx.Ctx, clientset, machine.IP, machine.Labels, machine.Taints)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureRegistryHosts(ctx *common.ClusterContext) error {
	if !p.Cfg.NeedSetHosts() {
		return nil
	}

	domains := []string{
		p.Cfg.Registry.Domain,
		ctx.Cluster.Spec.TenantID + "." + p.Cfg.Registry.Domain,
	}
	for _, machine := range ctx.Cluster.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		for _, one := range domains {
			remoteHosts := &hosts.RemoteHosts{Host: one, SSH: sh}
			err := remoteHosts.Set(p.Cfg.Registry.IP)
			if err != nil {
				return errors.Wrap(err, machine.IP)
			}
		}
	}

	return nil
}

func (p *Provider) EnsurePreInstallHook(ctx *common.ClusterContext) error {
	if ctx.Cluster.Spec.Features.Hooks == nil {
		return nil
	}

	hook := ctx.Cluster.Spec.Features.Hooks[devopsv1.HookPreInstall]
	if hook == "" {
		return nil
	}
	cmd := strings.Split(hook, " ")[0]

	for _, machine := range ctx.Cluster.Spec.Machines {
		s, err := machine.SSH()
		if err != nil {
			return err
		}

		s.Execf("chmod +x %s", cmd)
		_, stderr, exit, err := s.Exec(hook)
		if err != nil || exit != 0 {
			return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", hook, exit, stderr, err)
		}
	}
	return nil
}

func (p *Provider) EnsurePostInstallHook(ctx *common.ClusterContext) error {
	if ctx.Cluster.Spec.Features.Hooks == nil {
		return nil
	}

	hook := ctx.Cluster.Spec.Features.Hooks[devopsv1.HookPostInstall]
	if hook == "" {
		return nil
	}
	cmd := strings.Split(hook, " ")[0]

	for _, machine := range ctx.Cluster.Spec.Machines {
		s, err := machine.SSH()
		if err != nil {
			return err
		}

		s.Execf("chmod +x %s", cmd)
		_, stderr, exit, err := s.Exec(hook)
		if err != nil || exit != 0 {
			return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", hook, exit, stderr, err)
		}
	}
	return nil
}

func (p *Provider) EnsureRebuildEtcd(ctx *common.ClusterContext) error {
	etcdPeerEndpoints := []string{}
	etcdClusterEndpoints := []string{}
	for _, machine := range ctx.Cluster.Spec.Machines {
		etcdPeerEndpoints = append(etcdPeerEndpoints, fmt.Sprintf("%s=https://%s:2380", machine.IP, machine.IP))
		etcdClusterEndpoints = append(etcdClusterEndpoints, fmt.Sprintf("https://%s:2379", machine.IP))
	}

	for _, machine := range ctx.Cluster.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		etcdByte, err := sh.ReadFile(constants.EtcdPodManifestFile)
		if err != nil {
			return fmt.Errorf("node: %s ReadFile: %s failed error: %v", machine.IP, constants.EtcdPodManifestFile, err)
		}

		etcdObj, err := k8sutil.UnmarshalFromYaml(etcdByte, corev1.SchemeGroupVersion)
		if err != nil {
			return fmt.Errorf("node: %s marshalling %s failed error: %v", machine.IP, constants.EtcdPodManifestFile, err)
		}

		if etcdPod, ok := etcdObj.(*corev1.Pod); ok {
			isFindState := false
			isFindLogger := false
			ctx.Info("rebuild etcd", "pod-name", etcdPod.Name, "cmd", etcdPod.Spec.Containers[0].Command)
			for i, arg := range etcdPod.Spec.Containers[0].Command {
				if strings.HasPrefix(arg, "--initial-cluster=") {
					etcdPod.Spec.Containers[0].Command[i] = fmt.Sprintf("--initial-cluster=%s", strings.Join(etcdPeerEndpoints, ","))
				}
				if strings.HasPrefix(arg, "--initial-cluster-state=") {
					isFindState = true
				}

				if strings.HasPrefix(arg, "--logger=") {
					isFindLogger = true
				}
			}

			if !isFindState {
				etcdPod.Spec.Containers[0].Command = append(etcdPod.Spec.Containers[0].Command, "--initial-cluster-state=existing")
			}

			if !isFindLogger {
				etcdPod.Spec.Containers[0].Command = append(etcdPod.Spec.Containers[0].Command, "--logger=zap")
			}
			serialized, err := k8sutil.MarshalToYaml(etcdPod, corev1.SchemeGroupVersion)
			if err != nil {
				return errors.Wrapf(err, "failed to marshal manifest for %q to YAML", etcdPod.Name)
			}

			sh.WriteFile(bytes.NewReader(serialized), constants.EtcdPodManifestFile)
		}

		apiServerByte, err := sh.ReadFile(constants.KubeAPIServerPodManifestFile)
		if err != nil {
			return fmt.Errorf("node: %s ReadFile: %s failed error: %v", machine.IP, constants.KubeAPIServerPodManifestFile, err)
		}

		apiServerObj, err := k8sutil.UnmarshalFromYaml(apiServerByte, corev1.SchemeGroupVersion)
		if err != nil {
			return fmt.Errorf("node: %s marshalling %s failed error: %v", machine.IP, constants.KubeAPIServerPodManifestFile, err)
		}

		var ok bool
		var apiServerPod *corev1.Pod
		if apiServerPod, ok = apiServerObj.(*corev1.Pod); !ok {
			continue
		}

		ctx.Info("rebuild apiserver", "pod-name", apiServerPod.Name, "cmd", apiServerPod.Spec.Containers[0].Command)
		for i, arg := range apiServerPod.Spec.Containers[0].Command {
			if !strings.HasPrefix(arg, "--etcd-servers=") {
				continue
			}

			apiServerPod.Spec.Containers[0].Command[i] = fmt.Sprintf("--etcd-servers=%s", strings.Join(etcdClusterEndpoints, ","))
			break
		}

		serialized, err := k8sutil.MarshalToYaml(apiServerPod, corev1.SchemeGroupVersion)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal manifest for %q to YAML", apiServerPod.Name)
		}

		sh.WriteFile(bytes.NewReader(serialized), constants.KubeAPIServerPodManifestFile)
	}

	return nil
}

func (p *Provider) EnsureRebuildControlPlane(ctx *common.ClusterContext) error {
	for _, machine := range ctx.Cluster.Spec.Machines[1:] {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}
		err = kubemisc.CovertMasterKubeConfig(sh, ctx)
		if err != nil {
			return err
		}

		_, _, _, err = sh.Execf("systemctl enable kubelet && systemctl restart kubelet")
		if err != nil {
			return err
		}
		// err = kubeadm.RestartContainerByFilter(sh, kubeadm.DockerFilterForControlPlane("kube-apiserver"))
		// if err != nil {
		// 	return err
		// }
		// err = kubeadm.RestartContainerByFilter(sh, kubeadm.DockerFilterForControlPlane("kube-controller-manager"))
		// if err != nil {
		// 	return err
		// }
		// err = kubeadm.RestartContainerByFilter(sh, kubeadm.DockerFilterForControlPlane("kube-scheduler"))
		// if err != nil {
		// 	return err
		// }
	}

	return nil
}

func (p *Provider) EnsureExtKubeconfig(ctx *common.ClusterContext) error {
	if ctx.Credential.ExtData == nil {
		ctx.Credential.ExtData = make(map[string]string)
	}

	apiserver := certs.BuildExternalApiserverEndpoint(ctx)
	cfgMaps, err := certs.CreateApiserverKubeConfigFile(ctx.Credential.CAKey, ctx.Credential.CACert, apiserver, ctx.Cluster.Name)
	if err != nil {
		ctx.Error(err, "build apiserver kubeconfg")
		return err
	}
	ctx.Info("start convert apiserver kubeconfig ...", "apiserver", apiserver)
	for _, v := range cfgMaps {
		by, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			return err
		}

		externalKubeconfig := string(by)
		ctx.Info("make externalKubeconfig", "file", externalKubeconfig)
		ctx.Credential.ExtData[pkiutil.ExternalAdminKubeConfigFileName] = externalKubeconfig
	}

	return nil
}

func (p *Provider) EnsureMetricsServer(ctx *common.ClusterContext) error {
	clusterCtx, err := ctx.ClusterManager.Get(ctx.Cluster.Name)
	if err != nil {
		return nil
	}
	objs, err := metricsserver.BuildMetricsServerAddon(ctx)
	if err != nil {
		return errors.Wrapf(err, "build metrics-server err: %v", err)
	}

	logger := ctx.WithValues("component", "metrics-server")
	logger.Info("start reconcile ...")
	for _, obj := range objs {
		err = k8sutil.Reconcile(logger, clusterCtx.GetClient(), obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "Reconcile  err: %v", err)
		}
	}

	return nil
}

func (p *Provider) EnsureEth(ctx *common.ClusterContext) error {
	var cniType string
	var ok bool

	if cniType, ok = ctx.Cluster.Spec.Features.Hooks[devopsv1.HookCniInstall]; !ok {
		return nil
	}

	if cniType != "dke-cni" {
		return nil
	}

	for _, machine := range ctx.Cluster.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		err = rawcni.ApplyEth(sh, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) EnsureDeployCni(ctx *common.ClusterContext) error {
	var cniType string
	var ok bool

	if cniType, ok = ctx.Cluster.Spec.Features.Hooks[devopsv1.HookCniInstall]; !ok {
		return nil
	}

	switch cniType {
	case "dke-cni":
		for _, machine := range ctx.Cluster.Spec.Machines {
			sh, err := machine.SSH()
			if err != nil {
				return err
			}

			err = rawcni.ApplyCniCfg(sh, ctx)
			if err != nil {
				return err
			}
		}
	case "flannel":
		clusterCtx, err := ctx.ClusterManager.Get(ctx.Cluster.Name)
		if err != nil {
			return nil
		}
		objs, err := flannel.BuildFlannelAddon(p.Cfg, ctx)
		if err != nil {
			return errors.Wrapf(err, "build flannel err: %v", err)
		}

		logger := ctx.WithValues("component", "flannel")
		logger.Info("start reconcile ...")
		for _, obj := range objs {
			err = k8sutil.Reconcile(logger, clusterCtx.GetClient(), obj, k8sutil.DesiredStatePresent)
			if err != nil {
				return errors.Wrapf(err, "Reconcile  err: %v", err)
			}
		}
	default:
		return fmt.Errorf("unknown cni type: %s", cniType)
	}

	return nil
}

func (p *Provider) EnsureMasterNode(ctx *common.ClusterContext) error {
	clusterCtx, err := ctx.ClusterManager.Get(ctx.Cluster.Name)
	if err != nil {
		return nil
	}

	node := &corev1.Node{}
	var noReadNode *devopsv1.ClusterMachine
	for _, machine := range ctx.Cluster.Spec.Machines {
		err := clusterCtx.GetClient().Get(ctx.Ctx, types.NamespacedName{Name: machine.IP}, node)
		if err != nil {
			return errors.Wrapf(err, "failed get cluster: %s node: %s", ctx.Cluster.Name, machine.IP)
		}

		isNoReady := false
		for j := range node.Status.Conditions {
			if node.Status.Conditions[j].Type != corev1.NodeReady {
				continue
			}

			if node.Status.Conditions[j].Status != corev1.ConditionTrue {
				isNoReady = true
			}
			break
		}

		if isNoReady {
			noReadNode = machine
			break
		}
	}

	if noReadNode == nil {
		return nil
	}

	ctx.Info("start reconcile master", "node", noReadNode.IP)
	sh, err := noReadNode.SSH()
	if err != nil {
		return err
	}

	for _, file := range ctx.Cluster.Spec.Features.Files {
		err = system.CopyFile(sh, &file)
		if err != nil {
			return err
		}
	}

	phases := []func(ctx *common.ClusterContext, s ssh.Interface) error{
		system.Install,
		kubebin.Install,
		preflight.RunMasterChecks,
		kubemisc.Install,
		kubeadm.JoinControlPlane,
	}

	for _, phase := range phases {
		err = phase(ctx, sh)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) EnsureNvidiaDriver(ctx *common.ClusterContext) error {

	return nil
}

func (p *Provider) EnsureNvidiaContainerRuntime(ctx *common.ClusterContext) error {
	return nil
}
