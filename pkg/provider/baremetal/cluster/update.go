/*
Copyright 2020 wtxue.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/common"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/phases/certs"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/phases/kubeadm"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider/phases/kubemisc"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/k8sutil"
	certutil "k8s.io/client-go/util/cert"
)

func (p *Provider) EnsureRenewCerts(ctx context.Context, c *common.Cluster) error {
	for _, machine := range c.Spec.Machines {
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

func (p *Provider) EnsureAPIServerCert(ctx context.Context, c *common.Cluster) error {
	apiserver := certs.BuildApiserverEndpoint(c.Cluster.Spec.PublicAlternativeNames[0], kubemisc.GetBindPort(c.Cluster))

	kubeadmConfig := kubeadm.GetKubeadmConfig(c, p.Cfg, apiserver)
	exptectCertSANs := k8sutil.GetAPIServerCertSANs(c.Cluster)

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
