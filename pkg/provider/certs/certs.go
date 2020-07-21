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

package certs

import (
	"crypto"
	"crypto/x509"

	"github.com/pkg/errors"
	kubeadmv1beta2 "github.com/wtxue/kube-on-kube-operator/pkg/apis/kubeadm/v1beta2"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/pkiutil"
	"k8s.io/klog"
)

type CaAll struct {
	CaCert *x509.Certificate
	CaKey  crypto.Signer
	Cfg    *KubeadmCert
}

// CreateCACertAndKeyFiles generates and writes out a given certificate authority.
// The certSpec should be one of the variables from this package.
func CreateCACertAndKeyFiles(certSpec *KubeadmCert, cfg *kubeadmv1beta2.WarpperConfiguration, cfgMaps map[string][]byte) (*CaAll, error) {
	if certSpec.CAName != "" {
		return nil, errors.Errorf("this function should only be used for CAs, but cert %s has CA %s", certSpec.Name, certSpec.CAName)
	}
	klog.V(1).Infof("creating a new certificate authority for %s", certSpec.Name)

	certConfig, err := certSpec.GetConfig(cfg)
	if err != nil {
		return nil, err
	}

	caCert, caKey, err := pkiutil.NewCertificateAuthority(certConfig)
	if err != nil {
		return nil, err
	}

	keyPath, keyByte, err := pkiutil.BuildKeyByte(cfg.CertificatesDir, certSpec.BaseName, caKey)
	if err != nil {
		return nil, err
	}

	cfgMaps[keyPath] = keyByte
	certPath, certByte, err := pkiutil.BuildCertByte(cfg.CertificatesDir, certSpec.BaseName, caCert)
	if err != nil {
		return nil, err
	}
	cfgMaps[certPath] = certByte

	return &CaAll{
		CaCert: caCert,
		CaKey:  caKey,
		Cfg:    certSpec}, nil
}

func CreateCertAndKeyFilesWithCA(certSpec *KubeadmCert, ca *CaAll, cfg *kubeadmv1beta2.WarpperConfiguration, certsMaps map[string][]byte) error {
	if certSpec.CAName != ca.Cfg.Name {
		return errors.Errorf("expected CAname for %s to be %q, but was %s", certSpec.Name, certSpec.CAName, ca.Cfg.Name)
	}

	certConfig, err := certSpec.GetConfig(cfg)
	if err != nil {
		return errors.Wrapf(err, "couldn't create %q certificate", certSpec.Name)
	}

	cert, key, err := pkiutil.NewCertAndKey(ca.CaCert, ca.CaKey, certConfig)
	if err != nil {
		return err
	}

	keyPath, keyByte, err := pkiutil.BuildKeyByte(cfg.CertificatesDir, certSpec.BaseName, key)
	if err != nil {
		return err
	}

	certsMaps[keyPath] = keyByte
	certPath, certByte, err := pkiutil.BuildCertByte(cfg.CertificatesDir, certSpec.BaseName, cert)
	if err != nil {
		return err
	}
	certsMaps[certPath] = certByte
	return nil
}

// CreateServiceAccountKeyAndPublicKeyFiles creates new public/private key files for signing service account users.
// If the sa public/private key files already exist in the target folder, they are used only if evaluated equals; otherwise an error is returned.
func CreateServiceAccountKeyAndPublicKeyFiles(certsDir string, keyType x509.PublicKeyAlgorithm, certsMaps map[string][]byte) error {
	klog.V(1).Infoln("creating new public/private key files for signing service account users")

	// The key does NOT exist, let's generate it now
	key, err := pkiutil.NewPrivateKey(keyType)
	if err != nil {
		return err
	}

	// Write .key and .pub files to remote
	klog.Infof("[certs] Generating %q key and public key\n", pkiutil.ServiceAccountKeyBaseName)
	keyPath, keyByte, err := pkiutil.BuildKeyByte(certsDir, pkiutil.ServiceAccountKeyBaseName, key)
	if err != nil {
		return err
	}
	certsMaps[keyPath] = keyByte

	publicPath, publicByte, err := pkiutil.BuildPublicKeyByte(certsDir, pkiutil.ServiceAccountKeyBaseName, key.Public())
	if err != nil {
		return err
	}
	certsMaps[publicPath] = publicByte
	return nil
}
