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

package clustermanager

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	logger = logf.Log.WithName("controller")
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
		klog.Errorf("clusters list is empty, nothing to delete")
		return nil
	}

	index, ok := m.GetClusterIndex(name)
	if !ok {
		klog.Warningf("cluster:%s  is not found in the registries list, nothing to delete", name)
		return nil
	}

	m.clusters[index].Stop()
	clusters := m.clusters
	clusters = append(clusters[:index], clusters[index+1:]...)
	m.clusters = clusters
	klog.Infof("the cluster %s has been deleted.", name)
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
	klog.V(5).Info("cluster health check.")
	for _, c := range m.clusters {
		if !c.healthCheck() {
			klog.Warningf("cluster: %s healthCheck fail", c.Name)
		}
	}
}

func (m *ClusterManager) AddNewClusters(name string, kubeconfig string) (*Cluster, error) {
	if c, _ := m.Get(name); c != nil {
		return c, nil
	}

	nc, err := NewCluster(name, []byte(kubeconfig), logger)
	if err != nil {
		klog.Errorf("cluster: %s new client err: %v", name, err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	nc.StartCache(ctx.Done())
	err = m.Add(nc)
	if err != nil {
		klog.Errorf("cluster: %s add err: %+v", name, err)
		return nil, err
	}
	return nc, nil
}

// Start timer check cluster health
func (m *ClusterManager) Start(stopCh <-chan struct{}) error {
	klog.V(4).Info("multi cluster manager start check loop ... ")
	wait.Until(m.cluterCheck, time.Minute, stopCh)

	klog.V(4).Info("multi cluster manager stoped ... ")
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
