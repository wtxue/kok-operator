package common

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/clustermanager"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultTimeout = 30 * time.Second
	defaultQPS     = 100
	defaultBurst   = 200
)

type ClusterContext struct {
	Ctx        context.Context
	Key        types.NamespacedName
	Cluster    *devopsv1.Cluster
	Credential *devopsv1.ClusterCredential
	*clustermanager.ClusterManager
	client.Client
	logr.Logger
}

func FillClusterContext(ctx *ClusterContext, multiMgr *clustermanager.ClusterManager) error {
	clusterCredential := &devopsv1.ClusterCredential{}
	err := ctx.Client.Get(ctx.Ctx, ctx.Key, clusterCredential)
	if err != nil && apierrors.IsNotFound(err) {
		ctx.Info("not find credential, start create ...", "cluster ", ctx.Cluster.Name)
		credential := &devopsv1.ClusterCredential{
			ObjectMeta: k8sutil.ObjectMeta(ctx.Cluster.Name, constants.CtrlLabels, ctx.Cluster),
			CredentialInfo: devopsv1.CredentialInfo{
				TenantID:    ctx.Cluster.Spec.TenantID,
				ClusterName: ctx.Cluster.Name,
			},
		}

		err := ctx.Client.Create(ctx.Ctx, credential)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}

		clusterCredential = credential
	}

	if err != nil {
		return err
	}

	ctx.Credential = clusterCredential
	if multiMgr != nil {
		ctx.ClusterManager = multiMgr
	}
	return nil
}

func (c *ClusterContext) GetClusterID() string {
	return c.Cluster.GetName()
}

func (c *ClusterContext) GetAPIServerName() string {
	return fmt.Sprintf("%s-%s", c.Cluster.GetName(), constants.KubeApiServer)
}

func (c *ClusterContext) GetNamespace() string {
	return c.Cluster.GetNamespace()
}

func (c *ClusterContext) Clientset() (kubernetes.Interface, error) {
	config, err := c.RESTConfig(&rest.Config{})
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func (c *ClusterContext) ClientsetForBootstrap() (kubernetes.Interface, error) {
	config, err := c.RESTConfigForBootstrap(&rest.Config{})
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func (c *ClusterContext) RESTConfigForBootstrap(config *rest.Config) (*rest.Config, error) {
	host, err := c.HostForBootstrap()
	if err != nil {
		return nil, err
	}
	config.Host = host

	return c.RESTConfig(config)
}
func (c *ClusterContext) RESTConfig(config *rest.Config) (*rest.Config, error) {
	if config.Host == "" {
		host, err := c.Host()
		if err != nil {
			return nil, err
		}
		config.Host = host
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.QPS == 0 {
		config.QPS = defaultQPS
	}
	if config.Burst == 0 {
		config.Burst = defaultBurst
	}

	if c.Credential.CACert != nil {
		config.TLSClientConfig.CAData = c.Credential.CACert
	} else {
		config.TLSClientConfig.Insecure = true
	}
	if c.Credential.ClientCert != nil && c.Credential.ClientKey != nil {
		config.TLSClientConfig.CertData = c.Credential.ClientCert
		config.TLSClientConfig.KeyData = c.Credential.ClientKey
	}

	if c.Credential.Token != nil {
		config.BearerToken = *c.Credential.Token
	}

	return config, nil
}

func (c *ClusterContext) Host() (string, error) {
	addrs := make(map[devopsv1.AddressType][]devopsv1.ClusterAddress)
	for _, one := range c.Cluster.Status.Addresses {
		addrs[one.Type] = append(addrs[one.Type], one)
	}

	var address *devopsv1.ClusterAddress
	if len(addrs[devopsv1.AddressInternal]) != 0 {
		address = &addrs[devopsv1.AddressInternal][rand.Intn(len(addrs[devopsv1.AddressInternal]))]
	} else if len(addrs[devopsv1.AddressAdvertise]) != 0 {
		address = &addrs[devopsv1.AddressAdvertise][rand.Intn(len(addrs[devopsv1.AddressAdvertise]))]
	} else {
		if len(addrs[devopsv1.AddressReal]) != 0 {
			address = &addrs[devopsv1.AddressReal][rand.Intn(len(addrs[devopsv1.AddressReal]))]
		}
	}

	if address == nil {
		return "", errors.New("can't find valid address")
	}

	return fmt.Sprintf("%s:%d", address.Host, address.Port), nil
}

func (c *ClusterContext) HostForBootstrap() (string, error) {
	for _, one := range c.Cluster.Status.Addresses {
		if one.Type == devopsv1.AddressReal {
			return fmt.Sprintf("%s:%d", one.Host, one.Port), nil
		}
	}

	return "", errors.New("can't find bootstrap address")
}

func (c *ClusterContext) IPs() []string {
	ips := []string{}
	for _, m := range c.Cluster.Spec.Machines {
		ips = append(ips, m.IP)
	}
	return ips
}
