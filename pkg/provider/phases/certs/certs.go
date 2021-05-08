package certs

import (
	"crypto"
	"crypto/x509"

	"github.com/pkg/errors"
	kubeadmv1beta2 "github.com/wtxue/kok-operator/pkg/apis/kubeadm/v1beta2"
	"github.com/wtxue/kok-operator/pkg/util/pkiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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

	logf.Log.V(2).Info("creating a new certificate authority", "name", certSpec.Name)
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
	logf.Log.V(2).Info("creating new public/private key files for signing service account users")
	// The key does NOT exist, let's generate it now
	key, err := pkiutil.NewPrivateKey(keyType)
	if err != nil {
		return err
	}

	// Write .key and .pub files to remote
	logf.Log.V(2).Info("[certs] Generating key and public key", "name", pkiutil.ServiceAccountKeyBaseName)
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
