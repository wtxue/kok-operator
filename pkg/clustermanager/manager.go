package clustermanager

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	logger = logf.Log.WithName("clustermanager")
)

// key config
const (
	ClustersAll = "all"
)

// MasterClient ...
type MasterClient struct {
	KubeCli kubernetes.Interface
	manager.Manager
}

// ClusterManager ...
type ClusterManager struct {
	MasterClient
	clusters []*Cluster
	Started  bool
	sync.RWMutex
}

// NewManager ...
func NewManager(cli MasterClient) (*ClusterManager, error) {
	cMgr := &ClusterManager{
		MasterClient: cli,
		clusters:     make([]*Cluster, 0, 4),
	}

	cMgr.Started = true
	return cMgr, nil
}

// GetAll get all cluster
func (m *ClusterManager) GetAll(name ...string) []*Cluster {
	m.RLock()
	defer m.RUnlock()

	isAll := true
	var ObserveName string
	if len(name) > 0 {
		if name[0] != ClustersAll {
			isAll = false
		}
	}

	list := make([]*Cluster, 0, 4)
	for _, c := range m.clusters {
		if c.Status == ClusterOffline {
			continue
		}

		if isAll {
			list = append(list, c)
		} else {
			if ObserveName != "" && ObserveName == c.Name {
				list = append(list, c)
				break
			}
		}
	}

	return list
}

// Add ...
func (m *ClusterManager) Add(cluster *Cluster) error {
	if _, err := m.Get(cluster.Name); err == nil {
		return fmt.Errorf("cluster name: %s is already add to manager", cluster.Name)
	}

	m.Lock()
	defer m.Unlock()
	m.clusters = append(m.clusters, cluster)
	sort.Slice(m.clusters, func(i int, j int) bool {
		return m.clusters[i].Name > m.clusters[j].Name
	})

	return nil
}

// GetClusterIndex ...
func (m *ClusterManager) GetClusterIndex(name string) (int, bool) {
	for i, r := range m.clusters {
		if r.Name == name {
			return i, true
		}
	}
	return 0, false
}

// Delete ...
func (m *ClusterManager) Delete(name string) error {
	if name == "" {
		return nil
	}

	m.Lock()
	defer m.Unlock()

	if len(m.clusters) == 0 {
		logger.Info("clusters list is empty, nothing to delete")
		return nil
	}

	index, ok := m.GetClusterIndex(name)
	if !ok {
		logger.Info("not found in the registries list, nothing to delete", "cluster", name)
		return nil
	}

	m.clusters[index].Stop()
	clusters := m.clusters
	clusters = append(clusters[:index], clusters[index+1:]...)
	m.clusters = clusters
	logger.Info("has been deleted.", "cluster", name)
	return nil
}

// Get ...
func (m *ClusterManager) Get(name string) (*Cluster, error) {
	m.Lock()
	defer m.Unlock()

	if name == "" || name == "all" {
		return nil, fmt.Errorf("single query not support: %s ", name)
	}

	var findCluster *Cluster
	for _, c := range m.clusters {
		if name == c.Name {
			findCluster = c
			break
		}
	}
	if findCluster == nil {
		return nil, fmt.Errorf("cluster: %s not found", name)
	}

	if findCluster.Status == ClusterOffline {
		return nil, fmt.Errorf("cluster: %s found, but offline", name)
	}

	return findCluster, nil
}

func (m *ClusterManager) cluterCheck() {
	for _, c := range m.clusters {
		if !c.healthCheck() {
			logger.Info("healthCheck failed", "cluster", c.Name)
		}
	}
}

func (m *ClusterManager) AddNewClusters(name string, kubeconfig string) (*Cluster, error) {
	if c, _ := m.Get(name); c != nil {
		return c, nil
	}

	nc, err := NewCluster(name, []byte(kubeconfig), logger)
	if err != nil {
		logger.Error(err, "new cluster failed", "cluster", name)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	nc.StartCache(ctx)
	err = m.Add(nc)
	if err != nil {
		logger.Error(err, "add new cluster failed", "cluster", name)
		return nil, err
	}

	return nc, nil
}

// Start timer check cluster health
func (m *ClusterManager) Start(ctx context.Context) error {
	logger.Info("multi cluster manager start check loop ... ")
	wait.Until(m.cluterCheck, time.Minute, ctx.Done())

	logger.Info("multi cluster manager stoped ... ")
	m.Stop()
	return nil
}

func (m *ClusterManager) Stop() {
	m.Lock()
	defer m.Unlock()

	for _, cluster := range m.clusters {
		cluster.Stop()
	}
}
