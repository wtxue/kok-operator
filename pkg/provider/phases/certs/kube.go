package certs

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	kubeconfigutil "github.com/wtxue/kok-operator/pkg/util/kubeconfig"
	"github.com/wtxue/kok-operator/pkg/util/pkiutil"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog/v2"
)

// clientCertAuth struct holds info required to build a client certificate to provide authentication info in a kubeconfig object
type clientCertAuth struct {
	CAKey         crypto.Signer
	Organizations []string
}

// tokenAuth struct holds info required to use a token to provide authentication info in a kubeconfig object
type tokenAuth struct {
	Token string
}

// kubeConfigSpec struct holds info required to build a KubeConfig object
type kubeConfigSpec struct {
	CACert         *x509.Certificate
	APIServer      string
	ClientName     string
	TokenAuth      *tokenAuth
	ClientCertAuth *clientCertAuth
}

func LoadCertAndKeyFromByte(CAKey, CACert []byte) (*x509.Certificate, crypto.Signer, error) {
	certs, err := certutil.ParseCertsPEM(CACert)
	if err != nil {
		return nil, nil, fmt.Errorf(" reading error %v", err)
	}

	// use first ca
	cert := certs[0]

	// Check so that the certificate is valid now
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return nil, nil, errors.New("the certificate is not valid yet")
	}
	if now.After(cert.NotAfter) {
		return nil, nil, errors.New("the certificate has expired")
	}

	privKey, err := keyutil.ParsePrivateKeyPEM(CAKey)
	if err != nil {
		return nil, nil, fmt.Errorf("reading private key err: %v", err)
	}

	// Allow RSA and ECDSA formats only
	var key crypto.Signer
	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		key = k
	case *ecdsa.PrivateKey:
		key = k
	default:
		return nil, nil, errors.Errorf("the ca private key is neither in RSA nor ECDSA format")
	}

	return cert, key, nil
}

// getKubeConfigSpecs returns all KubeConfigSpecs actualized to the context of the current InitConfiguration
// NB. this methods holds the information about how kubeadm creates kubeconfig files.
func getKubeConfigSpecs(CAKey, CACert []byte, apiserver string, kubeletNodeAddr string) (map[string]*kubeConfigSpec, error) {

	caCert, caKey, err := LoadCertAndKeyFromByte(CAKey, CACert)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create a kubeconfig; the CA files couldn't be loaded")
	}

	var kubeConfigSpec = map[string]*kubeConfigSpec{
		pkiutil.AdminKubeConfigFileName: {
			CACert:     caCert,
			APIServer:  apiserver,
			ClientName: "kubernetes-admin",
			ClientCertAuth: &clientCertAuth{
				CAKey:         caKey,
				Organizations: []string{pkiutil.SystemPrivilegedGroup},
			},
		},
		pkiutil.KubeletKubeConfigFileName: {
			CACert:     caCert,
			APIServer:  apiserver,
			ClientName: fmt.Sprintf("%s%s", pkiutil.NodesUserPrefix, kubeletNodeAddr),
			ClientCertAuth: &clientCertAuth{
				CAKey:         caKey,
				Organizations: []string{pkiutil.NodesGroup},
			},
		},
		pkiutil.ControllerManagerKubeConfigFileName: {
			CACert:     caCert,
			APIServer:  apiserver,
			ClientName: pkiutil.ControllerManagerUser,
			ClientCertAuth: &clientCertAuth{
				CAKey: caKey,
			},
		},
		pkiutil.SchedulerKubeConfigFileName: {
			CACert:     caCert,
			APIServer:  apiserver,
			ClientName: pkiutil.SchedulerUser,
			ClientCertAuth: &clientCertAuth{
				CAKey: caKey,
			},
		},
	}

	return kubeConfigSpec, nil
}

// buildKubeConfigFromSpec creates a kubeconfig object for the given kubeConfigSpec
func buildKubeConfigFromSpec(spec *kubeConfigSpec, clustername string) (*clientcmdapi.Config, error) {

	// If this kubeconfig should use token
	if spec.TokenAuth != nil {
		// create a kubeconfig with a token
		return kubeconfigutil.CreateWithToken(
			spec.APIServer,
			clustername,
			spec.ClientName,
			pkiutil.EncodeCertPEM(spec.CACert),
			spec.TokenAuth.Token,
		), nil
	}

	// otherwise, create a client certs
	clientCertConfig := pkiutil.CertConfig{
		Config: certutil.Config{
			CommonName:   spec.ClientName,
			Organization: spec.ClientCertAuth.Organizations,
			Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		},
	}
	clientCert, clientKey, err := pkiutil.NewCertAndKey(spec.CACert, spec.ClientCertAuth.CAKey, &clientCertConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failure while creating %s client certificate", spec.ClientName)
	}

	encodedClientKey, err := keyutil.MarshalPrivateKeyToPEM(clientKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal private key to PEM")
	}
	// create a kubeconfig with the client certs
	return kubeconfigutil.CreateWithCerts(
		spec.APIServer,
		clustername,
		spec.ClientName,
		pkiutil.EncodeCertPEM(spec.CACert),
		encodedClientKey,
		pkiutil.EncodeCertPEM(clientCert),
	), nil
}

func BuildKubeConfigByte(config *clientcmdapi.Config) ([]byte, error) {
	return runtime.Encode(clientcmdlatest.Codec, config)
}

func DecodeKubeConfigByte(data []byte, config *clientcmdapi.Config) error {
	return runtime.DecodeInto(clientcmdlatest.Codec, data, config)
}

// createKubeConfigFiles creates all the requested kubeconfig files.
// If kubeconfig files already exists, they are used only if evaluated equal; otherwise an error is returned.
func CreateKubeConfigFiles(CAKey, CACert []byte, apiserver string, kubeletNodeAddr string, clusterName string, kubeConfigFileNames ...string) (map[string]*clientcmdapi.Config, error) {
	cfgMaps := make(map[string]*clientcmdapi.Config)
	// gets the KubeConfigSpecs, actualized for the current InitConfiguration
	specs, err := getKubeConfigSpecs(CAKey, CACert, apiserver, kubeletNodeAddr)
	if err != nil {
		return nil, err
	}

	for _, kubeConfigFileName := range kubeConfigFileNames {
		klog.V(1).Infof("creating kubeconfig file for %s", kubeConfigFileName)
		// retrieves the KubeConfigSpec for given kubeConfigFileName
		spec, exists := specs[kubeConfigFileName]
		if !exists {
			return cfgMaps, errors.Errorf("couldn't retrieve KubeConfigSpec for %s", kubeConfigFileName)
		}

		// builds the KubeConfig object
		config, err := buildKubeConfigFromSpec(spec, clusterName)
		if err != nil {
			return cfgMaps, err
		}

		cfgMaps[kubeConfigFileName] = config
	}

	return cfgMaps, nil
}

func BuildExternalApiserverEndpoint(ctx *common.ClusterContext) string {
	var vip string
	port := "6443"
	vipMasterKey := constants.GetAnnotationKey(ctx.Cluster.Annotations, constants.ClusterApiSvcVip)
	if vipMasterKey != "" {
		vip = vipMasterKey
	}

	if ctx.Cluster.Spec.Features.HA != nil && ctx.Cluster.Spec.Features.HA.ThirdPartyHA != nil {
		port = fmt.Sprintf("%d", ctx.Cluster.Spec.Features.HA.ThirdPartyHA.VPort)
		if vip == "" {
			vip = ctx.Cluster.Spec.Features.HA.ThirdPartyHA.VIP
		}
	}

	if vip == "" && len(ctx.Cluster.Spec.Machines) > 0 {
		vip = ctx.Cluster.Spec.Machines[0].IP
	}

	controlPlaneURL := &url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(vip, port),
	}

	return controlPlaneURL.String()
}

func BuildApiserverEndpoint(ipOrDns string, bindPort int) string {
	bindPortString := strconv.Itoa(bindPort)
	controlPlaneURL := &url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(ipOrDns, bindPortString),
	}

	return controlPlaneURL.String()
}

func CreateKubeletKubeConfigFile(CAKey, CACert []byte, apiserver string, kubeletNodeAddr string, clusterName string) (map[string]*clientcmdapi.Config, error) {
	return CreateKubeConfigFiles(CAKey, CACert, apiserver, kubeletNodeAddr, clusterName, GetKubeletKubeconfigList()...)
}

func CreateMasterKubeConfigFile(CAKey, CACert []byte, apiserver string, clusterName string) (map[string]*clientcmdapi.Config, error) {
	return CreateKubeConfigFiles(CAKey, CACert, apiserver, "", clusterName, GetMasterKubeConfigList()...)
}

func CreateApiserverKubeConfigFile(CAKey, CACert []byte, apiserver string, clusterName string) (map[string]*clientcmdapi.Config, error) {
	return CreateKubeConfigFiles(CAKey, CACert, apiserver, "", clusterName, GetApiserverKubeconfigList()...)
}

func GetKubeletKubeconfigList() []string {
	return []string{
		pkiutil.KubeletKubeConfigFileName,
	}
}

func GetApiserverKubeconfigList() []string {
	return []string{
		pkiutil.AdminKubeConfigFileName,
	}
}

func GetMasterKubeConfigList() []string {
	return []string{
		pkiutil.AdminKubeConfigFileName,
		pkiutil.ControllerManagerKubeConfigFileName,
		pkiutil.SchedulerKubeConfigFileName,
	}
}
