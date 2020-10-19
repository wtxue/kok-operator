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

package apiserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"net"
	"net/url"
	"strconv"

	"k8s.io/client-go/rest"

	"github.com/wtxue/kok-operator/pkg/apiserver/internal"
	"github.com/wtxue/kok-operator/pkg/option"
	"k8s.io/klog"
)

const (
	envUseExistingCluster = "USE_EXISTING_CLUSTER"
	envStartTimeout       = "KUBEBUILDER_CONTROLPLANE_START_TIMEOUT"
	envStopTimeout        = "KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT"

	defaultKubebuilderControlPlaneStartTimeout = 20 * time.Second
	defaultKubebuilderControlPlaneStopTimeout  = 20 * time.Second
)

// Environment creates a Kubernetes test environment that will start / stop the Kubernetes control plane and
// install extension APIs
type ServerWarpper struct {
	// ControlPlane is the ControlPlane including the apiserver and etcd
	ControlPlane internal.ControlPlane

	// Config can be used to talk to the apiserver.  It's automatically
	// populated if not set using the standard controller-runtime config
	// loading.
	Config *rest.Config

	// UseExisting indicates that this environments should use an
	// existing kubeconfig, instead of trying to stand up a new control plane.
	// This is useful in cases that need aggregated API servers and the like.
	UseExistingCluster bool

	UseExistingEtcd bool

	// ControlPlaneStartTimeout is the maximum duration each controlplane component
	// may take to start. It defaults to the KUBEBUILDER_CONTROLPLANE_START_TIMEOUT
	// environment variable or 20 seconds if unspecified
	ControlPlaneStartTimeout time.Duration

	// ControlPlaneStopTimeout is the maximum duration each controlplane component
	// may take to stop. It defaults to the KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT
	// environment variable or 20 seconds if unspecified
	ControlPlaneStopTimeout time.Duration

	// KubeAPIServerFlags is the set of flags passed while starting the api server.
	KubeAPIServerFlags []string

	// AttachControlPlaneOutput indicates if control plane output will be attached to os.Stdout and os.Stderr.
	// Enable this to get more visibility of the testing control plane.
	// It respect KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT environment variable.
	AttachControlPlaneOutput bool

	BasePath string
	RootDir  string
	*option.ApiServerOption
}

func New(o *option.ApiServerOption) *ServerWarpper {
	return &ServerWarpper{
		UseExistingCluster:       !o.IsLocalKube,
		AttachControlPlaneOutput: true,
		BasePath:                 o.BaseBinDir,
		ApiServerOption:          o,
		RootDir:                  o.RootDir,
	}
}

// fillAssetPath
func (te *ServerWarpper) fillAssetPath(binary string) string {
	return filepath.Join(te.BasePath, binary)
}

// Stop stops a running server.
// Previously stop ControlPlane.
func (te *ServerWarpper) Stop() error {
	if te.UseExistingCluster {
		return nil
	}

	return te.ControlPlane.Stop()
}

// getAPIServerFlags returns flags to be used with the Kubernetes API server.
// it returns empty slice for api server defined defaults to be applied if no args specified
func (te ServerWarpper) getAPIServerFlags() []string {
	// Set default API server flags if not set.
	if len(te.KubeAPIServerFlags) == 0 {
		return []string{}
	}
	// Check KubeAPIServerFlags contains service-cluster-ip-range, if not, set default value to service-cluster-ip-range
	containServiceClusterIPRange := false
	for _, flag := range te.KubeAPIServerFlags {
		if strings.Contains(flag, "service-cluster-ip-range") {
			containServiceClusterIPRange = true
			break
		}
	}
	if !containServiceClusterIPRange {
		te.KubeAPIServerFlags = append(te.KubeAPIServerFlags, "--service-cluster-ip-range=10.0.0.0/24")
	}
	return te.KubeAPIServerFlags
}

// Start starts a local Kubernetes server and updates te.ApiserverPort with the port it is listening on
func (te *ServerWarpper) Start(stopCh <-chan struct{}) (*rest.Config, error) {
	if te.ControlPlane.APIServer == nil {
		te.ControlPlane.APIServer = &internal.APIServer{
			Path: te.fillAssetPath("kube-apiserver"),
			Args: te.getAPIServerFlags(),
			URL: &url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort("0.0.0.0", strconv.Itoa(8080)),
			},
			SecurePort: 8083,
		}
	}
	if te.ControlPlane.Etcd == nil {
		te.ControlPlane.Etcd = &internal.Etcd{
			Path:    te.fillAssetPath("etcd"),
			DataDir: te.RootDir + "/data/etcd",
			URL: &url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort("127.0.0.1", strconv.Itoa(12379)),
			},
		}
	}

	if te.ControlPlane.APIServer.Out == nil && te.AttachControlPlaneOutput {
		te.ControlPlane.APIServer.Out = os.Stdout
	}
	if te.ControlPlane.APIServer.Err == nil && te.AttachControlPlaneOutput {
		te.ControlPlane.APIServer.Err = os.Stderr
	}
	if te.ControlPlane.Etcd.Out == nil && te.AttachControlPlaneOutput {
		te.ControlPlane.Etcd.Out = os.Stdout
	}
	if te.ControlPlane.Etcd.Err == nil && te.AttachControlPlaneOutput {
		te.ControlPlane.Etcd.Err = os.Stderr
	}

	if err := te.defaultTimeouts(); err != nil {
		return nil, fmt.Errorf("failed to default controlplane timeouts: %w", err)
	}
	te.ControlPlane.Etcd.StartTimeout = te.ControlPlaneStartTimeout
	te.ControlPlane.Etcd.StopTimeout = te.ControlPlaneStopTimeout
	te.ControlPlane.APIServer.StartTimeout = te.ControlPlaneStartTimeout
	te.ControlPlane.APIServer.StopTimeout = te.ControlPlaneStopTimeout

	klog.Infof("starting control plane api server flags [%+v]", te.ControlPlane.APIServer.Args)
	if err := te.startControlPlane(); err != nil {
		return nil, err
	}

	go func(stopCh <-chan struct{}) {
		<-stopCh
		klog.Infof("stop control plane api server")
		te.Stop()
	}(stopCh)

	// Create the *rest.Config for creating new clients
	te.Config = &rest.Config{
		Host: te.ControlPlane.APIURL().Host,
		// gotta go fast during tests -- we don't really care about overwhelming our test API server
		QPS:   1000.0,
		Burst: 2000.0,
	}

	klog.Infof("start local kubernets Host: %s", te.Config.Host)
	return te.Config, nil
}

func (te *ServerWarpper) startControlPlane() error {
	numTries, maxRetries := 0, 5
	var err error
	for ; numTries < maxRetries; numTries++ {
		// Start the control plane - retry if it fails
		err = te.ControlPlane.Start()
		if err == nil {
			break
		}
		klog.Error(err, "unable to start the controlplane", "tries", numTries)
	}
	if numTries == maxRetries {
		return fmt.Errorf("failed to start the controlplane. retried %d times: %w", numTries, err)
	}
	return nil
}

func (te *ServerWarpper) defaultTimeouts() error {
	var err error
	if te.ControlPlaneStartTimeout == 0 {
		if envVal := os.Getenv(envStartTimeout); envVal != "" {
			te.ControlPlaneStartTimeout, err = time.ParseDuration(envVal)
			if err != nil {
				return err
			}
		} else {
			te.ControlPlaneStartTimeout = defaultKubebuilderControlPlaneStartTimeout
		}
	}

	if te.ControlPlaneStopTimeout == 0 {
		if envVal := os.Getenv(envStopTimeout); envVal != "" {
			te.ControlPlaneStopTimeout, err = time.ParseDuration(envVal)
			if err != nil {
				return err
			}
		} else {
			te.ControlPlaneStopTimeout = defaultKubebuilderControlPlaneStopTimeout
		}
	}
	return nil
}

// DefaultKubeAPIServerFlags exposes the default args for the APIServer so that
// you can use those to append your own additional arguments.
var DefaultKubeAPIServerFlags = internal.APIServerDefaultArgs
