package cluster

import (
	"sync"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
)

const conditionTypeHealthCheck = "HealthCheck"
const conditionTypeSyncVersion = "SyncVersion"
const reasonHealthCheckFail = "HealthCheckFail"

type clusterHealth struct {
	mu         sync.Mutex
	clusterMap map[string]*devopsv1.Cluster
}

func (s *clusterHealth) Exist(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.clusterMap[key]
	return ok
}

func (s *clusterHealth) Set(key string, cluster *devopsv1.Cluster) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clusterMap[key] = cluster
}

func (s *clusterHealth) Del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clusterMap, key)
}
