package clustermanager

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/k8sclient"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ClusterStatusType string

// These are valid status of a cluster.
const (
	// ClusterReady means the cluster is ready to accept workloads.
	ClusterReady ClusterStatusType = "Ready"
	// ClusterOffline means the cluster is temporarily down or not reachable
	ClusterOffline  ClusterStatusType = "Offline"
	ClusterMaintain ClusterStatusType = "Maintaining"
)

var (
	SyncPeriodTime = 1 * time.Hour
)

type Cluster struct {
	Name          string
	AliasName     string
	RawKubeconfig []byte
	Meta          map[string]string
	RestConfig    *rest.Config
	Client        client.Client
	KubeCli       kubernetes.Interface

	Log        logr.Logger
	Mgr        manager.Manager
	Cache      cache.Cache
	SyncPeriod time.Duration

	StopperCancel context.CancelFunc

	Status ClusterStatusType
	// Started is true if the Informers has been Started
	Started bool
}

func NewCluster(name string, kubeconfig []byte, log logr.Logger) (*Cluster, error) {
	cluster := &Cluster{
		Name:          name,
		RawKubeconfig: kubeconfig,
		Log:           log.WithValues("cluster", name),
		SyncPeriod:    SyncPeriodTime,
		Started:       false,
	}

	err := cluster.initK8SClients()
	if err != nil {
		return nil, errors.Wrapf(err, "could not re-init k8s clients name:%s", name)
	}

	return cluster, nil
}

func (c *Cluster) GetName() string {
	return c.Name
}

func (c *Cluster) initK8SClients() error {
	startTime := time.Now()
	cfg, err := k8sclient.NewClientConfig(c.RawKubeconfig)
	if err != nil {
		return errors.Wrapf(err, "could not get rest config name: %s", c.Name)
	}

	c.RestConfig = cfg
	kubecli, err := kubernetes.NewForConfig(c.RestConfig)
	if err != nil {
		return errors.Wrapf(err, "could not new kubecli name:%s", c.Name)
	}

	c.Log.Info("new kube cli", "time taken", fmt.Sprintf("%v", time.Since(startTime)))
	c.KubeCli = kubecli
	opt := manager.Options{
		Scheme:                 k8sclient.GetScheme(),
		SyncPeriod:             &c.SyncPeriod,
		MetricsBindAddress:     "0",
		HealthProbeBindAddress: "0",
		LeaderElection:         false,
	}

	mgr, err := manager.New(c.RestConfig, opt)
	if err != nil {
		return errors.Wrapf(err, "could not new manager name: %s", c.Name)
	}

	c.Mgr = mgr
	c.Client = mgr.GetClient()
	c.Cache = mgr.GetCache()
	c.Log.Info("new kube manager", "time taken", fmt.Sprintf("%v", time.Since(startTime)))
	return nil
}

func (c *Cluster) healthCheck() bool {
	body, err := c.KubeCli.Discovery().RESTClient().Get().AbsPath("/healthz").Do(context.TODO()).Raw()
	if err != nil {
		runtime.HandleError(errors.Wrapf(err, "Failed to do cluster health check for cluster %q", c.Name))
		c.Status = ClusterOffline
		return false
	}

	if !strings.EqualFold(string(body), "ok") {
		c.Status = ClusterOffline
		return false
	}
	c.Status = ClusterReady
	return true
}

func (c *Cluster) StartCache(stopCtx context.Context) {
	if c.Started {
		c.Log.Info("cache informers is already startd")
		return
	}

	c.Log.Info("start cache informers ... ")
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		err := c.Cache.Start(ctx)
		if err != nil {
			c.Log.Error(err, "cache Informers quit")
		}
	}()

	c.Cache.WaitForCacheSync(stopCtx)
	c.Started = true
	c.StopperCancel = cancelFunc
}

func (c *Cluster) Stop() {
	c.Log.Info("start stop cache informers")
	c.StopperCancel()
}
