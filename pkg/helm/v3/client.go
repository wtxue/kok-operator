package v3

import (
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/helm/object"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ReleaseInfo copy of the struct form the helm library
type ReleaseInfo struct {
	// FirstDeployed is when the release was first deployed.
	FirstDeployed time.Time `json:"first_deployed,omitempty"`
	// LastDeployed is when the release was last deployed.
	LastDeployed time.Time `json:"last_deployed,omitempty"`
	// Deleted tracks when this object was deleted.
	Deleted time.Time `json:"deleted"`
	// Description is human-friendly "log entry" about this release.
	Description string `json:"description,omitempty"`
	// Status is the current state of the release
	Status string
	// Contains the rendered templates/NOTES.txt if available
	Notes string
	// Contains override values provided to the release
	Values map[string]interface{}
}

//  Release represents information related to a helm chart release
type Release struct {
	// ReleaseInput struct encapsulating information about the release to be created
	ReleaseName      string
	ChartName        string
	Namespace        string
	Values           map[string]interface{} //json representation
	Version          string
	ReleaseInfo      *ReleaseInfo
	ReleaseVersion   int32
	ReleaseResources []*object.K8sObject
	ChartPackage     []byte
}

func (ri *Release) NameAndChartSlice() []string {
	if ri.ReleaseName == "" {
		return []string{ri.ChartName}
	}
	return []string{ri.ReleaseName, ri.ChartName}
}

type customGetter struct {
	kubeConfigBytes []byte
	apiconfig       *clientcmdapi.Config
	logger          logr.Logger
	namespace       string
	cacheDir        string
}

func NewCustomGetter(namespace string, kubeconfig []byte, cacheDir string, logger logr.Logger) (genericclioptions.RESTClientGetter, error) {
	apiconfig, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load kubernetes API config")
	}

	return customGetter{
		kubeConfigBytes: kubeconfig,
		apiconfig:       apiconfig,
		logger:          logger,
		namespace:       namespace,
		cacheDir:        cacheDir,
	}, nil
}

func (c customGetter) ToRESTConfig() (*rest.Config, error) {
	return c.ToRawKubeConfigLoader().ClientConfig()
}

func (c customGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	// The more groups you have, the more discovery requests you need to make.
	// given 25 groups (our groups + a few custom resources) with one-ish version each, discovery needs to make 50 requests
	// double it just so we don't end up here again for a while.  This config is only used for discovery.
	config.Burst = 100

	return disk.NewCachedDiscoveryClientForConfig(
		config,
		filepath.Join(c.cacheDir, "discovery-cache", fmt.Sprintf("%x", sha1.Sum([]byte(config.Host)))),
		filepath.Join(c.cacheDir, "http-cache"),
		time.Minute*10,
	)
}

func (c customGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

func (c customGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return clientcmd.NewDefaultClientConfig(*c.apiconfig, &clientcmd.ConfigOverrides{Context: clientcmdapi.Context{
		Namespace: c.namespace,
	}})
}
