package cluster

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/wtxue/kok-operator/pkg/addons/coredns"
	"github.com/wtxue/kok-operator/pkg/addons/flannel"
	"github.com/wtxue/kok-operator/pkg/addons/kubeproxy"
	"github.com/wtxue/kok-operator/pkg/addons/metricsserver"
	"github.com/wtxue/kok-operator/pkg/addons/rawcni"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	"github.com/wtxue/kok-operator/pkg/provider/phases/certs"
	"github.com/wtxue/kok-operator/pkg/provider/phases/kubeadm"
	"github.com/wtxue/kok-operator/pkg/provider/phases/kubemisc"
	"github.com/wtxue/kok-operator/pkg/util/pkiutil"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	tokenFileTemplate = `%s,admin,admin,system:masters
`
	additPolicy = `
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: Metadata
`
)

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

func (p *Provider) ping(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, "pong")
}

func (p *Provider) EnsurePreInstallHook(ctx *common.ClusterContext) error {
	ctx.Info("ingore pre install")
	return nil
}

func (p *Provider) EnsurePostInstallHook(ctx *common.ClusterContext) error {
	ctx.Info("ingore post install")
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
	return nil
}

func (p *Provider) EnsureCerts(ctx *common.ClusterContext) error {
	apiserver := certs.BuildApiserverEndpoint(ctx.GetAPIServerName(), 6443)
	err := kubeadm.InitCerts(ctx, kubeadm.GetKubeadmConfig(ctx, p.Cfg, apiserver), true)
	if err != nil {
		return err
	}

	return ApplyCertsConfigmap(ctx, ctx.Credential.CertsBinaryData)
}

func (p *Provider) EnsureKubeMisc(ctx *common.ClusterContext) error {
	apiserver := certs.BuildApiserverEndpoint(ctx.GetAPIServerName(), kubemisc.GetBindPort(ctx.Cluster))
	err := kubemisc.BuildMasterMiscConfigToMap(ctx, apiserver)
	if err != nil {
		return err
	}

	return ApplyKubeMiscConfigmap(ctx, ctx.Credential.KubeData)
}

func (p *Provider) EnsureEtcd(ctx *common.ClusterContext) error {
	ctx.Info("use exists etcd cluster")
	return nil
}

func (p *Provider) EnsureKubeMaster(ctx *common.ClusterContext) error {
	r := &Reconciler{
		Ctx:      ctx,
		Provider: p,
	}

	var objs []client.Object
	objs = append(objs, r.apiServerDeployment())
	objs = append(objs, r.apiServerSvc())
	objs = append(objs, r.controllerManagerDeployment())
	objs = append(objs, r.schedulerDeployment())

	logger := ctx.WithValues("cluster", ctx.Cluster.Name)
	for _, obj := range objs {
		err := k8sutil.Reconcile(logger, ctx.Client, obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "apply object err: %v", err)
		}
	}

	return nil
}

func (p *Provider) EnsureKubeconfig(ctx *common.ClusterContext) error {
	if ctx.Credential.ExtData == nil {
		ctx.Credential.ExtData = make(map[string]string)
	}

	onKubeApiserver := certs.BuildApiserverEndpoint(ctx.GetOnKubeAPIServerName(), kubemisc.GetBindPort(ctx.Cluster))
	cfgMaps, err := certs.CreateApiserverKubeConfigFile(ctx.Credential.CAKey, ctx.Credential.CACert, onKubeApiserver, ctx.Cluster.Name)
	if err != nil {
		return err
	}

	for _, cfg := range cfgMaps {
		ctx.Info("start build onkube kubeconfig ...", "apiserver", onKubeApiserver)
		by, err := certs.BuildKubeConfigByte(cfg)
		if err != nil {
			return err
		}

		ctx.Credential.ExtData[pkiutil.OnKubeAdminKubeConfigFileName] = string(by)

		extApiserver := certs.BuildApiserverEndpoint(ctx.Cluster.Spec.PublicAlternativeNames[0], kubemisc.GetBindPort(ctx.Cluster))
		ctx.Info("start build ext kubeconfig ...", "apiserver", onKubeApiserver)

		for _, c := range cfg.Clusters {
			c.Server = extApiserver
		}
		by, err = certs.BuildKubeConfigByte(cfg)
		if err != nil {
			return err
		}

		ctx.Credential.ExtData[pkiutil.ExternalAdminKubeConfigFileName] = string(by)
	}

	return nil
}

func (p *Provider) EnsureAddons(ctx *common.ClusterContext) error {
	clusterCtx, err := ctx.ClusterManager.Get(ctx.Cluster.Name)
	if err != nil {
		return nil
	}

	kubeproxyObjs, err := kubeproxy.BuildKubeproxyAddon(p.Cfg, ctx)
	if err != nil {
		return errors.Wrapf(err, "build kube-proxy")
	}

	logger := ctx.WithValues("cluster", ctx.Cluster.Name)
	logger.Info("start apply kube-proxy")
	for _, obj := range kubeproxyObjs {
		err = k8sutil.Reconcile(logger, clusterCtx.GetClient(), obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "reconcile")
		}
	}

	logger.Info("start apply coredns")
	corednsObjs, err := coredns.BuildCoreDNSAddon(p.Cfg, ctx)
	if err != nil {
		return errors.Wrapf(err, "build coredns")
	}
	for _, obj := range corednsObjs {
		err = k8sutil.Reconcile(logger, clusterCtx.GetClient(), obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "reconcile")
		}
	}
	return nil
}

func (p *Provider) EnsureCni(ctx *common.ClusterContext) error {
	cniType := "flannel"
	if findCNIType, ok := ctx.Cluster.Spec.Features.Hooks[devopsv1.HookCniInstall]; ok {
		cniType = findCNIType
	}

	switch cniType {
	case "raw-cni":
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
		clusterCtx, err := ctx.ClusterManager.Get(ctx.GetClusterID())
		if err != nil {
			return err
		}
		objs, err := flannel.BuildFlannelAddon(p.Cfg, ctx)
		if err != nil {
			return errors.Wrapf(err, "build flannel")
		}

		logger := ctx.WithValues("component", "flannel")
		logger.Info("start reconcile ...")
		for _, obj := range objs {
			err = k8sutil.Reconcile(logger, clusterCtx.GetClient(), obj, k8sutil.DesiredStatePresent)
			if err != nil {
				return errors.Wrapf(err, "reconcile")
			}
		}
	default:
		return fmt.Errorf("unknown cni type: %s", cniType)
	}

	return nil
}

func (p *Provider) EnsureMetricsServer(ctx *common.ClusterContext) error {
	clusterCtx, err := ctx.ClusterManager.Get(ctx.GetClusterID())
	if err != nil {
		return err
	}
	objs, err := metricsserver.BuildMetricsServerAddon(p.Cfg, ctx)
	if err != nil {
		return errors.Wrapf(err, "build metrics-server")
	}

	logger := ctx.WithValues("component", "metrics-server")
	logger.Info("start reconcile ...")
	for _, obj := range objs {
		err = k8sutil.Reconcile(logger, clusterCtx.GetClient(), obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "reconcile")
		}
	}

	return nil
}
