package cluster

import (
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	"github.com/wtxue/kok-operator/pkg/provider/phases/certs"
	"github.com/wtxue/kok-operator/pkg/provider/phases/kubeadm"
	"github.com/wtxue/kok-operator/pkg/provider/phases/kubemisc"
	certutil "k8s.io/client-go/util/cert"
)

func (p *Provider) EnsureRenewCerts(ctx *common.ClusterContext) error {
	for _, machine := range ctx.Cluster.Spec.Machines {
		s, err := machine.SSH()
		if err != nil {
			return err
		}

		data, err := s.ReadFile(constants.APIServerCertName)
		if err != nil {
			return err
		}
		cts, err := certutil.ParseCertsPEM(data)
		if err != nil {
			return err
		}
		expirationDuration := time.Until(cts[0].NotAfter)
		if expirationDuration > constants.RenewCertsTimeThreshold {
			ctx.Info("skip EnsureRenewCerts because expiration duration > threshold",
				"duration", expirationDuration, "threshold", constants.RenewCertsTimeThreshold)
			return nil
		}

		ctx.Info("EnsureRenewCerts", "node", s.Host)
		err = kubeadm.RenewCerts(s)
		if err != nil {
			return errors.Wrapf(err, "renew certs node: %s", machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureAPIServerCert(ctx *common.ClusterContext) error {
	apiserver := certs.BuildApiserverEndpoint(ctx.Cluster.Spec.PublicAlternativeNames[0], kubemisc.GetBindPort(ctx.Cluster))

	kubeadmConfig := kubeadm.GetKubeadmConfig(ctx, p.Cfg, apiserver)
	exptectCertSANs := k8sutil.GetAPIServerCertSANs(ctx.Cluster)

	needUpload := false
	for _, machine := range ctx.Cluster.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		data, err := sh.ReadFile(constants.APIServerCertName)
		if err != nil {
			return err
		}
		certList, err := certutil.ParseCertsPEM(data)
		if err != nil {
			return err
		}
		actualCertSANs := certList[0].DNSNames
		for _, ip := range certList[0].IPAddresses {
			actualCertSANs = append(actualCertSANs, ip.String())
		}

		if reflect.DeepEqual(funk.IntersectString(actualCertSANs, exptectCertSANs), exptectCertSANs) {
			return nil
		}

		ctx.Info("EnsureAPIServerCert", "node", sh.Host)
		for _, file := range []string{constants.APIServerCertName, constants.APIServerKeyName} {
			sh.CombinedOutput(fmt.Sprintf("rm -f %s", file))
		}

		err = kubeadm.Init(ctx, sh, kubeadmConfig, "certs apiserver")
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
		err = kubeadm.RestartContainerByFilter(sh, kubeadm.DockerFilterForControlPlane("kube-apiserver"))
		if err != nil {
			return err
		}

		needUpload = true
	}

	if needUpload {
		err := p.EnsureKubeadmInitUploadConfigPhase(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
