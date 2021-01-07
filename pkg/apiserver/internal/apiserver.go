package internal

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// APIServerDefaultArgs allow tests to run offline, by preventing API server from attempting to
// use default route to determine its --advertise-address.
var APIServerDefaultArgs = []string{
	"--advertise-address=0.0.0.0",
	"--bind-address=0.0.0.0",
	"--etcd-servers={{ if .EtcdURL }}{{ .EtcdURL.String }}{{ end }}",
	"--cert-dir={{ .CertDir }}",
	"--insecure-port={{ if .URL }}{{ .URL.Port }}{{ end }}",
	"--insecure-bind-address={{ if .URL }}{{ .URL.Hostname }}{{ end }}",
	"--secure-port={{ if .SecurePort }}{{ .SecurePort }}{{ end }}",
	"--disable-admission-plugins=ServiceAccount",
	"--service-cluster-ip-range=10.10.0.0/24",
	"--allow-privileged=true",
}

// DoAPIServerArgDefaulting will set default values to allow tests to run offline when the args are not informed. Otherwise,
// it will return the same []string arg passed as param.
func DoAPIServerArgDefaulting(args []string) []string {
	if len(args) != 0 {
		return args
	}

	return APIServerDefaultArgs
}

// APIServer knows how to run a kubernetes apiserver.
type APIServer struct {
	// URL is the address the ApiServer should listen on for client connections.
	//
	// If this is not specified, we default to a random free port on localhost.
	URL *url.URL

	// SecurePort is the additional secure port that the APIServer should listen on.
	SecurePort int

	// Path is the path to the apiserver binary.
	//
	// If this is left as the empty string, we will attempt to locate a binary,
	// by checking for the TEST_ASSET_KUBE_APISERVER environment variable, and
	// the default test assets directory. See the "Binaries" section above (in
	// doc.go) for details.
	Path string

	// Args is a list of arguments which will passed to the APIServer binary.
	// Before they are passed on, they will be evaluated as go-template strings.
	// This means you can use fields which are defined and exported on this
	// APIServer struct (e.g. "--cert-dir={{ .Dir }}").
	// Those templates will be evaluated after the defaulting of the APIServer's
	// fields has already happened and just before the binary actually gets
	// started. Thus you have access to calculated fields like `URL` and others.
	//
	// If not specified, the minimal set of arguments to run the APIServer will
	// be used.
	Args []string

	// CertDir is a path to a directory containing whatever certificates the
	// APIServer will need.
	//
	// If left unspecified, then the Start() method will create a fresh temporary
	// directory, and the Stop() method will clean it up.
	CertDir string

	// EtcdURL is the URL of the Etcd the APIServer should use.
	//
	// If this is not specified, the Start() method will return an error.
	EtcdURL *url.URL

	// StartTimeout, StopTimeout specify the time the APIServer is allowed to
	// take when starting and stoppping before an error is emitted.
	//
	// If not specified, these default to 20 seconds.
	StartTimeout time.Duration
	StopTimeout  time.Duration

	// Out, Err specify where APIServer should write its StdOut, StdErr to.
	//
	// If not specified, the output will be discarded.
	Out io.Writer
	Err io.Writer

	processState *ProcessState
}

// Start starts the apiserver, waits for it to come up, and returns an error,
// if occurred.
func (s *APIServer) Start() error {
	if s.processState == nil {
		if err := s.setProcessState(); err != nil {
			return err
		}
	}
	return s.processState.Start(s.Out, s.Err)
}

func (s *APIServer) setProcessState() error {
	if s.EtcdURL == nil {
		return fmt.Errorf("expected EtcdURL to be configured")
	}

	var err error

	s.processState = &ProcessState{}
	s.processState.DefaultedProcessInput, err = DoDefaulting(
		"kube-apiserver",
		s.URL,
		s.CertDir,
		s.Path,
		s.StartTimeout,
		s.StopTimeout,
	)
	if err != nil {
		return err
	}

	s.processState.HealthCheckEndpoint = "/healthz"

	s.URL = &s.processState.URL
	s.CertDir = s.processState.Dir
	s.Path = s.processState.Path
	s.StartTimeout = s.processState.StartTimeout
	s.StopTimeout = s.processState.StopTimeout

	if err := s.populateAPIServerCerts(); err != nil {
		return err
	}

	s.processState.Args, err = RenderTemplates(
		DoAPIServerArgDefaulting(s.Args), s,
	)
	return err
}

func (s *APIServer) populateAPIServerCerts() error {
	_, statErr := os.Stat(filepath.Join(s.CertDir, "apiserver.crt"))
	if !os.IsNotExist(statErr) {
		return statErr
	}

	ca, err := NewTinyCA()
	if err != nil {
		return err
	}

	certs, err := ca.NewServingCert()
	if err != nil {
		return err
	}

	certData, keyData, err := certs.AsBytes()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(s.CertDir, "apiserver.crt"), certData, 0640); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(s.CertDir, "apiserver.key"), keyData, 0640); err != nil {
		return err
	}

	return nil
}

// Stop stops this process gracefully, waits for its termination, and cleans up
// the CertDir if necessary.
func (s *APIServer) Stop() error {
	if s.processState != nil {
		return s.processState.Stop()
	}
	return nil
}
