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
