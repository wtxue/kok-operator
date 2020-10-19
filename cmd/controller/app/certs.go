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

package app

import (
	"crypto"
	"crypto/x509"
	"net"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/wtxue/kok-operator/cmd/controller/app/app_option"
	"github.com/wtxue/kok-operator/pkg/util/pkiutil"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog"
)

func writeCertificateAuthorityFilesIfNotExist(pkiDir string, baseName string, caCert *x509.Certificate, caKey crypto.Signer) error {
	// If cert or key exists, we should try to load them
	if pkiutil.CertOrKeyExist(pkiDir, baseName) {

		// Try to load .crt and .key from the PKI directory
		caCert, _, err := pkiutil.TryLoadCertAndKeyFromDisk(pkiDir, baseName)
		if err != nil {
			return errors.Wrapf(err, "failure loading %s certificate", baseName)
		}

		// Check if the existing cert is a CA
		if !caCert.IsCA {
			return errors.Errorf("certificate %s is not a CA", baseName)
		}

		// kubeadm doesn't validate the existing certificate Authority more than this;
		// Basically, if we find a certificate file with the same path; and it is a CA
		// kubeadm thinks those files are equal and doesn't bother writing a new file
		klog.Infof("[certs] Using the existing %s certificate and key\n", baseName)
	} else {
		// Write .crt and .key files to disk
		klog.Infof("[certs] Generating %s certificate and key\n", baseName)

		if err := pkiutil.WriteCertAndKey(pkiDir, baseName, caCert, caKey); err != nil {
			return errors.Wrapf(err, "failure while saving %s certificate and key", baseName)
		}
	}
	return nil
}

func TryRun() error {
	klog.Infof("try run cert ... ")

	klog.Infof("start generating ca certificate and key ... ")
	caCsr := &pkiutil.CertConfig{
		PublicKeyAlgorithm: x509.RSA,
		Config: certutil.Config{
			CommonName:   "kubernetes",
			Organization: []string{"k8s"},
			Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		},
	}

	caCert, caKey, err := pkiutil.NewCertificateAuthority(caCsr)
	if err != nil {
		return err
	}
	klog.Infof("start generating ca certificate and key  success ")

	apiserverCsr := &pkiutil.CertConfig{
		PublicKeyAlgorithm: x509.RSA,
		Config: certutil.Config{
			CommonName:   "kube-apiserver",
			Organization: []string{"dke"},
			AltNames: certutil.AltNames{
				DNSNames: []string{
					"vip-otdyiqyb.dke.k8s.io", // dns
					"localhost",
					"kubernetes",
					"kubernetes.default",
					"kubernetes.default.svc",
					"kubernetes.default.svc.cluster",
					"kubernetes.default.svc.cluster.local",
				},
				IPs: []net.IP{
					net.ParseIP("127.0.0.1"),    //
					net.ParseIP("10.12.99.10"),  // master1
					net.ParseIP("10.12.99.10"),  // master2
					net.ParseIP("10.12.99.10"),  // master3
					net.ParseIP("10.12.99.254"), // vip
					net.ParseIP("10.96.0.1"),    // kubernetes svc ip, always first svc cidr
				},
			},
			Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		},
	}

	pkiutil.NewCertAndKey(caCert, caKey, apiserverCsr)
	writeCertificateAuthorityFilesIfNotExist("./tools/pki", "ca", caCert, caKey)
	return nil
}

func NewCertCmd(opt *app_option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert",
		Short: "Manage k8s cert demo",
		Run: func(cmd *cobra.Command, args []string) {
			TryRun()
		},
	}

	return cmd
}
