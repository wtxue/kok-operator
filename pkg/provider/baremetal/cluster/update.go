package cluster

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/thoas/go-funk"
	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/phases/kubeadm"
	certutil "k8s.io/client-go/util/cert"
)

func (p *Provider) EnsureRenewCerts(ctx context.Context, c *provider.Cluster) error {
	for _, machine := range c.Spec.Machines {
		s, err := machine.SSH()
		if err != nil {
			return err
		}

		data, err := s.ReadFile(constants.APIServerCertName)
		if err != nil {
			return err
		}
		certs, err := certutil.ParseCertsPEM(data)
		if err != nil {
			return err
		}
		expirationDuration := time.Until(certs[0].NotAfter)
		if expirationDuration > constants.RenewCertsTimeThreshold {
			log.Infof("skip EnsureRenewCerts because expiration duration(%s) > threshold(%s)", expirationDuration, constants.RenewCertsTimeThreshold)
			return nil
		}

		log.Infof("EnsureRenewCerts for %s", s.Host)
		err = kubeadm.RenewCerts(s)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureAPIServerCert(ctx context.Context, c *provider.Cluster) error {
	kubeadmConfig := p.getKubeadmConfig(c)
	exptectCertSANs := GetAPIServerCertSANs(c.Cluster)

	needUpload := false
	for _, machine := range c.Spec.Machines {
		s, err := machine.SSH()
		if err != nil {
			return err
		}

		data, err := s.ReadFile(constants.APIServerCertName)
		if err != nil {
			return err
		}
		certs, err := certutil.ParseCertsPEM(data)
		if err != nil {
			return err
		}
		actualCertSANs := certs[0].DNSNames
		for _, ip := range certs[0].IPAddresses {
			actualCertSANs = append(actualCertSANs, ip.String())
		}
		if reflect.DeepEqual(funk.IntersectString(actualCertSANs, exptectCertSANs), exptectCertSANs) {
			return nil
		}

		log.Infof("EnsureAPIServerCert for %s", s.Host)
		for _, file := range []string{constants.APIServerCertName, constants.APIServerKeyName} {
			s.CombinedOutput(fmt.Sprintf("rm -f %s", file))
		}

		err = kubeadm.Init(s, kubeadmConfig, "certs apiserver")
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
		err = kubeadm.RestartContainerByFilter(s, kubeadm.DockerFilterForControlPlane("kube-apiserver"))
		if err != nil {
			return err
		}

		needUpload = true
	}

	if needUpload {
		err := p.EnsureKubeadmInitUploadConfigPhase(ctx, c)
		if err != nil {
			return err
		}
	}

	return nil
}
