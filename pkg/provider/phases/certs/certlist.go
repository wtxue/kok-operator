package certs

import (
	"crypto"
	"crypto/x509"

	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/apis"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/util/pkiutil"
	utilcert "k8s.io/client-go/util/cert"
)

type configMutatorsFunc func(*apis.WarpperConfiguration, *pkiutil.CertConfig) error

// KubeadmCert represents a certificate that Kubeadm will create to function properly.
type KubeadmCert struct {
	Name     string
	LongName string
	BaseName string
	CAName   string
	// Some attributes will depend on the InitConfiguration, only known at runtime.
	// These functions will be run in series, passed both the InitConfiguration and a cert Config.
	configMutators []configMutatorsFunc
	config         pkiutil.CertConfig
}

// GetConfig returns the definition for the given cert given the provided InitConfiguration
func (k *KubeadmCert) GetConfig(ic *apis.WarpperConfiguration) (*pkiutil.CertConfig, error) {
	for _, f := range k.configMutators {
		if err := f(ic, &k.config); err != nil {
			return nil, err
		}
	}

	k.config.PublicKeyAlgorithm = x509.RSA
	return &k.config, nil
}

// CreateFromCA makes and writes a certificate using the given CA cert and key.
func (k *KubeadmCert) CreateFromCA(ic *apis.WarpperConfiguration, caCert *x509.Certificate, caKey crypto.Signer) error {
	cfg, err := k.GetConfig(ic)
	if err != nil {
		return errors.Wrapf(err, "couldn't create %q certificate", k.Name)
	}
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, cfg)
	if err != nil {
		return err
	}
	err = pkiutil.WriteCertificateFilesIfNotExist(
		ic.CertificatesDir,
		k.BaseName,
		caCert,
		cert,
		key,
		cfg,
	)

	if err != nil {
		return errors.Wrapf(err, "failed to write or validate certificate %q", k.Name)
	}

	return nil
}

// CreateAsCA creates a certificate authority, writing the files to disk and also returning the created CA so it can be used to sign child certs.
func (k *KubeadmCert) CreateAsCA(ic *apis.WarpperConfiguration) (*x509.Certificate, crypto.Signer, error) {
	cfg, err := k.GetConfig(ic)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't get configuration for %q CA certificate", k.Name)
	}
	caCert, caKey, err := pkiutil.NewCertificateAuthority(cfg)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't generate %q CA certificate", k.Name)
	}

	err = pkiutil.WriteCertificateAuthorityFilesIfNotExist(
		ic.CertificatesDir,
		k.BaseName,
		caCert,
		caKey,
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't write out %q CA certificate", k.Name)
	}

	return caCert, caKey, nil
}

// CertificateTree is represents a one-level-deep tree, mapping a CA to the certs that depend on it.
type CertificateTree map[*KubeadmCert]Certificates

// CertificateMap is a flat map of certificates, keyed by Name.
type CertificateMap map[string]*KubeadmCert

// CertTree returns a one-level-deep tree, mapping a CA cert to an array of certificates that should be signed by it.
func (m CertificateMap) CertTree() (CertificateTree, error) {
	caMap := make(CertificateTree)

	for _, cert := range m {
		if cert.CAName == "" {
			if _, ok := caMap[cert]; !ok {
				caMap[cert] = []*KubeadmCert{}
			}
		} else {
			ca, ok := m[cert.CAName]
			if !ok {
				return nil, errors.Errorf("certificate %q references unknown CA %q", cert.Name, cert.CAName)
			}
			caMap[ca] = append(caMap[ca], cert)
		}
	}

	return caMap, nil
}

// Certificates is a list of Certificates that Kubeadm should create.
type Certificates []*KubeadmCert

// AsMap returns the list of certificates as a map, keyed by name.
func (c Certificates) AsMap() CertificateMap {
	certMap := make(map[string]*KubeadmCert)
	for _, cert := range c {
		certMap[cert.Name] = cert
	}

	return certMap
}

// GetDefaultCertList returns  all of the certificates kubeadm requires to function.
func GetDefaultCertList() Certificates {
	return Certificates{
		&KubeadmCertRootCA,
		&KubeadmCertAPIServer,
		&KubeadmCertKubeletClient,
		// Front Proxy certs
		&KubeadmCertFrontProxyCA,
		&KubeadmCertFrontProxyClient,
		// etcd certs
		&KubeadmCertEtcdCA,
		&KubeadmCertEtcdServer,
		&KubeadmCertEtcdPeer,
		&KubeadmCertEtcdHealthcheck,
		&KubeadmCertEtcdAPIClient,
	}
}

// GetCertsWithoutEtcd returns all of the certificates kubeadm needs when etcd is hosted externally.
func GetCertsWithoutEtcd() Certificates {
	return Certificates{
		&KubeadmCertRootCA,
		&KubeadmCertAPIServer,
		&KubeadmCertKubeletClient,
		// Front Proxy certs
		&KubeadmCertFrontProxyCA,
		&KubeadmCertFrontProxyClient,
	}
}

var (
	// KubeadmCertRootCA is the definition of the Kubernetes Root CA for the API Server and kubelet.
	KubeadmCertRootCA = KubeadmCert{
		Name:     "ca",
		LongName: "self-signed Kubernetes CA to provision identities for other Kubernetes components",
		BaseName: constants.CACertAndKeyBaseName,
		config: pkiutil.CertConfig{
			Config: utilcert.Config{
				CommonName: "kubernetes",
			},
		},
	}
	// KubeadmCertAPIServer is the definition of the cert used to serve the Kubernetes API.
	KubeadmCertAPIServer = KubeadmCert{
		Name:     "apiserver",
		LongName: "certificate for serving the Kubernetes API",
		BaseName: pkiutil.APIServerCertAndKeyBaseName,
		CAName:   "ca",
		config: pkiutil.CertConfig{
			Config: utilcert.Config{
				CommonName: pkiutil.APIServerCertCommonName,
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			},
		},
		configMutators: []configMutatorsFunc{
			makeAltNamesMutator(pkiutil.GetAPIServerAltNames),
		},
	}
	// KubeadmCertKubeletClient is the definition of the cert used by the API server to access the kubelet.
	KubeadmCertKubeletClient = KubeadmCert{
		Name:     "apiserver-kubelet-client",
		LongName: "certificate for the API server to connect to kubelet",
		BaseName: pkiutil.APIServerKubeletClientCertAndKeyBaseName,
		CAName:   "ca",
		config: pkiutil.CertConfig{
			Config: utilcert.Config{
				CommonName:   pkiutil.APIServerKubeletClientCertCommonName,
				Organization: []string{pkiutil.SystemPrivilegedGroup},
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}

	// KubeadmCertFrontProxyCA is the definition of the CA used for the front end proxy.
	KubeadmCertFrontProxyCA = KubeadmCert{
		Name:     "front-proxy-ca",
		LongName: "self-signed CA to provision identities for front proxy",
		BaseName: pkiutil.FrontProxyCACertAndKeyBaseName,
		config: pkiutil.CertConfig{
			Config: utilcert.Config{
				CommonName: "front-proxy-ca",
			},
		},
	}

	// KubeadmCertFrontProxyClient is the definition of the cert used by the API server to access the front proxy.
	KubeadmCertFrontProxyClient = KubeadmCert{
		Name:     "front-proxy-client",
		BaseName: pkiutil.FrontProxyClientCertAndKeyBaseName,
		LongName: "certificate for the front proxy client",
		CAName:   "front-proxy-ca",
		config: pkiutil.CertConfig{
			Config: utilcert.Config{
				CommonName: pkiutil.FrontProxyClientCertCommonName,
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}

	// KubeadmCertEtcdCA is the definition of the root CA used by the hosted etcd server.
	KubeadmCertEtcdCA = KubeadmCert{
		Name:     "etcd-ca",
		LongName: "self-signed CA to provision identities for etcd",
		BaseName: pkiutil.EtcdCACertAndKeyBaseName,
		config: pkiutil.CertConfig{
			Config: utilcert.Config{
				CommonName: "etcd-ca",
			},
		},
	}
	// KubeadmCertEtcdServer is the definition of the cert used to serve etcd to clients.
	KubeadmCertEtcdServer = KubeadmCert{
		Name:     "etcd-server",
		LongName: "certificate for serving etcd",
		BaseName: pkiutil.EtcdServerCertAndKeyBaseName,
		CAName:   "etcd-ca",
		config: pkiutil.CertConfig{
			Config: utilcert.Config{
				// TODO: etcd 3.2 introduced an undocumented requirement for ClientAuth usage on the
				// server cert: https://github.com/coreos/etcd/issues/9785#issuecomment-396715692
				// Once the upstream issue is resolved, this should be returned to only allowing
				// ServerAuth usage.
				Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			},
		},
		configMutators: []configMutatorsFunc{
			makeAltNamesMutator(pkiutil.GetEtcdAltNames),
			setCommonNameToNodeName(),
		},
	}
	// KubeadmCertEtcdPeer is the definition of the cert used by etcd peers to access each other.
	KubeadmCertEtcdPeer = KubeadmCert{
		Name:     "etcd-peer",
		LongName: "certificate for etcd nodes to communicate with each other",
		BaseName: pkiutil.EtcdPeerCertAndKeyBaseName,
		CAName:   "etcd-ca",
		config: pkiutil.CertConfig{
			Config: utilcert.Config{
				Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			},
		},
		configMutators: []configMutatorsFunc{
			makeAltNamesMutator(pkiutil.GetEtcdPeerAltNames),
			setCommonNameToNodeName(),
		},
	}
	// KubeadmCertEtcdHealthcheck is the definition of the cert used by Kubernetes to check the health of the etcd server.
	KubeadmCertEtcdHealthcheck = KubeadmCert{
		Name:     "etcd-healthcheck-client",
		LongName: "certificate for liveness probes to healthcheck etcd",
		BaseName: pkiutil.EtcdHealthcheckClientCertAndKeyBaseName,
		CAName:   "etcd-ca",
		config: pkiutil.CertConfig{
			Config: utilcert.Config{
				CommonName:   pkiutil.EtcdHealthcheckClientCertCommonName,
				Organization: []string{pkiutil.SystemPrivilegedGroup},
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}
	// KubeadmCertEtcdAPIClient is the definition of the cert used by the API server to access etcd.
	KubeadmCertEtcdAPIClient = KubeadmCert{
		Name:     "apiserver-etcd-client",
		LongName: "certificate the apiserver uses to access etcd",
		BaseName: pkiutil.APIServerEtcdClientCertAndKeyBaseName,
		CAName:   "etcd-ca",
		config: pkiutil.CertConfig{
			Config: utilcert.Config{
				CommonName:   pkiutil.APIServerEtcdClientCertCommonName,
				Organization: []string{pkiutil.SystemPrivilegedGroup},
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}
)

func makeAltNamesMutator(f func(*apis.WarpperConfiguration) (*utilcert.AltNames, error)) configMutatorsFunc {
	return func(mc *apis.WarpperConfiguration, cc *pkiutil.CertConfig) error {
		altNames, err := f(mc)
		if err != nil {
			return err
		}
		cc.AltNames = *altNames
		return nil
	}
}

func setCommonNameToNodeName() configMutatorsFunc {
	return func(mc *apis.WarpperConfiguration, cc *pkiutil.CertConfig) error {
		cc.CommonName = mc.NodeRegistration.Name
		return nil
	}
}
