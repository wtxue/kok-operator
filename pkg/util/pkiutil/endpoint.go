package pkiutil

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"strconv"

	kubeadmv1beta2 "github.com/wtxue/kok-operator/pkg/apis/kubeadm/v1beta2"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog/v2"
)

// GetControlPlaneEndpoint returns a properly formatted endpoint for the control plane built according following rules:
// - If the controlPlaneEndpoint is defined, use it.
// - if the controlPlaneEndpoint is defined but without a port number, use the controlPlaneEndpoint + localEndpoint.BindPort is used.
// - Otherwise, in case the controlPlaneEndpoint is not defined, use the localEndpoint.AdvertiseAddress + the localEndpoint.BindPort.
func GetControlPlaneEndpoint(controlPlaneEndpoint string, localEndpoint *kubeadmv1beta2.APIEndpoint) (string, error) {
	// parse the bind port
	bindPortString := strconv.Itoa(int(localEndpoint.BindPort))
	if _, err := ParsePort(bindPortString); err != nil {
		return "", errors.Wrapf(err, "invalid value %q given for api.bindPort", localEndpoint.BindPort)
	}

	// parse the AdvertiseAddress
	var ip = net.ParseIP(localEndpoint.AdvertiseAddress)
	if ip == nil {
		return "", errors.Errorf("invalid value `%s` given for api.advertiseAddress", localEndpoint.AdvertiseAddress)
	}

	// set the control-plane url using localEndpoint.AdvertiseAddress + the localEndpoint.BindPort
	controlPlaneURL := &url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(ip.String(), bindPortString),
	}

	// if the controlplane endpoint is defined
	if len(controlPlaneEndpoint) > 0 {
		// parse the controlplane endpoint
		var host, port string
		var err error
		if host, port, err = ParseHostPort(controlPlaneEndpoint); err != nil {
			return "", errors.Wrapf(err, "invalid value %q given for controlPlaneEndpoint", controlPlaneEndpoint)
		}

		// if a port is provided within the controlPlaneAddress warn the users we are using it, else use the bindport
		if port != "" {
			if port != bindPortString {
				fmt.Println("[endpoint] WARNING: port specified in controlPlaneEndpoint overrides bindPort in the controlplane address")
			}
		} else {
			port = bindPortString
		}

		// overrides the control-plane url using the controlPlaneAddress (and eventually the bindport)
		controlPlaneURL = &url.URL{
			Scheme: "https",
			Host:   net.JoinHostPort(host, port),
		}
	}

	return controlPlaneURL.String(), nil
}

// ParseHostPort parses a network address of the form "host:port", "ipv4:port", "[ipv6]:port" into host and port;
// ":port" can be eventually omitted.
// If the string is not a valid representation of network address, ParseHostPort returns an error.
func ParseHostPort(hostport string) (string, string, error) {
	var host, port string
	var err error

	// try to split host and port
	if host, port, err = net.SplitHostPort(hostport); err != nil {
		// if SplitHostPort returns an error, the entire hostport is considered as host
		host = hostport
	}

	// if port is defined, parse and validate it
	if port != "" {
		if _, err := ParsePort(port); err != nil {
			return "", "", errors.Errorf("hostport %s: port %s must be a valid number between 1 and 65535, inclusive", hostport, port)
		}
	}

	// if host is a valid IP, returns it
	if ip := net.ParseIP(host); ip != nil {
		return host, port, nil
	}

	// if host is a validate RFC-1123 subdomain, returns it
	if errs := validation.IsDNS1123Subdomain(host); len(errs) == 0 {
		return host, port, nil
	}

	return "", "", errors.Errorf("hostport %s: host '%s' must be a valid IP address or a valid RFC-1123 DNS subdomain", hostport, host)
}

// ParsePort parses a string representing a TCP port.
// If the string is not a valid representation of a TCP port, ParsePort returns an error.
func ParsePort(port string) (int, error) {
	portInt, err := strconv.Atoi(port)
	if err == nil && (1 <= portInt && portInt <= 65535) {
		return portInt, nil
	}

	return 0, errors.New("port must be a valid number between 1 and 65535, inclusive")
}

// validateCertificateWithConfig makes sure that a given certificate is valid at
// least for the SANs defined in the configuration.
func validateCertificateWithConfig(cert *x509.Certificate, baseName string, cfg *CertConfig) error {
	for _, dnsName := range cfg.AltNames.DNSNames {
		if err := cert.VerifyHostname(dnsName); err != nil {
			return errors.Wrapf(err, "certificate %s is invalid", baseName)
		}
	}
	for _, ipAddress := range cfg.AltNames.IPs {
		if err := cert.VerifyHostname(ipAddress.String()); err != nil {
			return errors.Wrapf(err, "certificate %s is invalid", baseName)
		}
	}
	return nil
}

// WriteCertificateFilesIfNotExist write a new certificate to the given path.
// If there already is a certificate file at the given path; kubeadm tries to load it and check if the values in the
// existing and the expected certificate equals. If they do; kubeadm will just skip writing the file as it's up-to-date,
// otherwise this function returns an error.
func WriteCertificateFilesIfNotExist(pkiDir string, baseName string, signingCert *x509.Certificate, cert *x509.Certificate, key crypto.Signer, cfg *CertConfig) error {
	// Checks if the signed certificate exists in the PKI directory
	if CertOrKeyExist(pkiDir, baseName) {
		// Try to load signed certificate .crt and .key from the PKI directory
		signedCert, _, err := TryLoadCertAndKeyFromDisk(pkiDir, baseName)
		if err != nil {
			return errors.Wrapf(err, "failure loading %s certificate", baseName)
		}

		// Check if the existing cert is signed by the given CA
		if err := signedCert.CheckSignatureFrom(signingCert); err != nil {
			return errors.Errorf("certificate %s is not signed by corresponding CA", baseName)
		}

		// Check if the certificate has the correct attributes
		if err := validateCertificateWithConfig(signedCert, baseName, cfg); err != nil {
			return err
		}

		klog.Infof("[certs] Using the existing %q certificate and key\n", baseName)
	} else {
		// Write .crt and .key files to disk
		klog.Infof("[certs] Generating %q certificate and key\n", baseName)

		if err := WriteCertAndKey(pkiDir, baseName, cert, key); err != nil {
			return errors.Wrapf(err, "failure while saving %s certificate and key", baseName)
		}
		if HasServerAuth(cert) {
			klog.Infof("[certs] %s serving cert is signed for DNS names %v and IPs %v\n", baseName, cert.DNSNames, cert.IPAddresses)
		}
	}

	return nil
}

// WriteCertificateAuthorityFilesIfNotExist write a new certificate Authority to the given path.
// If there already is a certificate file at the given path; kubeadm tries to load it and check if the values in the
// existing and the expected certificate equals. If they do; kubeadm will just skip writing the file as it's up-to-date,
// otherwise this function returns an error.
func WriteCertificateAuthorityFilesIfNotExist(pkiDir string, baseName string, caCert *x509.Certificate, caKey crypto.Signer) error {
	// If cert or key exists, we should try to load them
	if CertOrKeyExist(pkiDir, baseName) {

		// Try to load .crt and .key from the PKI directory
		caCert, _, err := TryLoadCertAndKeyFromDisk(pkiDir, baseName)
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
		klog.Infof("[certs] Using the existing %q certificate and key\n", baseName)
	} else {
		// Write .crt and .key files to disk
		klog.Infof("[certs] Generating %q certificate and key\n", baseName)

		if err := WriteCertAndKey(pkiDir, baseName, caCert, caKey); err != nil {
			return errors.Wrapf(err, "failure while saving %s certificate and key", baseName)
		}
	}
	return nil
}
