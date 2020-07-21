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

package v1

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/wtxue/kube-on-kube-operator/pkg/util/ssh"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionStatus defines the status of Condition.
type ConditionStatus string

// These are valid condition statuses.
// "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition.
// "ConditionUnknown" means server can't decide if a resource is in the condition
// or not.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// ClusterMachine is the master machine definition of cluster.
type ClusterMachine struct {
	IP       string `json:"ip"`
	Port     int32  `json:"port"`
	Username string `json:"username"`
	// +optional
	Password string `json:"password,omitempty"`
	// +optional
	PrivateKey []byte `json:"privateKey,omitempty"`
	// +optional
	PassPhrase []byte `json:"passPhrase,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// If specified, the node's taints.
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty"`
}

func (in *Cluster) SetCondition(newCondition ClusterCondition) {
	var conditions []ClusterCondition

	exist := false

	if newCondition.LastProbeTime.IsZero() {
		newCondition.LastProbeTime = metav1.Now()
	}
	for _, condition := range in.Status.Conditions {
		if condition.Type == newCondition.Type {
			exist = true
			if newCondition.LastTransitionTime.IsZero() {
				newCondition.LastTransitionTime = condition.LastTransitionTime
			}
			condition = newCondition
		}
		conditions = append(conditions, condition)
	}

	if !exist {
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = metav1.Now()
		}
		conditions = append(conditions, newCondition)
	}

	in.Status.Conditions = conditions
}

func (in *ClusterMachine) SSH() (*ssh.SSH, error) {
	sshConfig := &ssh.Config{
		User:        in.Username,
		Host:        in.IP,
		Port:        int(in.Port),
		Password:    in.Password,
		PrivateKey:  in.PrivateKey,
		PassPhrase:  in.PassPhrase,
		DialTimeOut: time.Second,
		Retry:       0,
	}
	return ssh.New(sshConfig)
}

func (in *Cluster) Address(addrType AddressType) *ClusterAddress {
	for _, one := range in.Status.Addresses {
		if one.Type == addrType {
			return &one
		}
	}

	return nil
}

func (in *Cluster) AddAddress(addrType AddressType, host string, port int32) {
	addr := ClusterAddress{
		Type: addrType,
		Host: host,
		Port: port,
	}
	for _, one := range in.Status.Addresses {
		if one == addr {
			return
		}
	}
	in.Status.Addresses = append(in.Status.Addresses, addr)
}

func (in *Cluster) RemoveAddress(addrType AddressType) {
	var addrs []ClusterAddress
	for _, one := range in.Status.Addresses {
		if one.Type == addrType {
			continue
		}
		addrs = append(addrs, one)
	}
	in.Status.Addresses = addrs
}

func (in *Cluster) Host() (string, error) {
	addrs := make(map[AddressType][]ClusterAddress)
	for _, one := range in.Status.Addresses {
		addrs[one.Type] = append(addrs[one.Type], one)
	}

	var address *ClusterAddress
	if len(addrs[AddressInternal]) != 0 {
		address = &addrs[AddressInternal][rand.Intn(len(addrs[AddressInternal]))]
	} else if len(addrs[AddressAdvertise]) != 0 {
		address = &addrs[AddressAdvertise][rand.Intn(len(addrs[AddressAdvertise]))]
	} else {
		if len(addrs[AddressReal]) != 0 {
			address = &addrs[AddressReal][rand.Intn(len(addrs[AddressReal]))]
		}
	}

	if address == nil {
		return "", errors.New("can't find valid address")
	}

	return fmt.Sprintf("%s:%d", address.Host, address.Port), nil
}

func (in *Machine) SetCondition(newCondition MachineCondition) {
	var conditions []MachineCondition

	exist := false

	if newCondition.LastProbeTime.IsZero() {
		newCondition.LastProbeTime = metav1.Now()
	}
	for _, condition := range in.Status.Conditions {
		if condition.Type == newCondition.Type {
			exist = true
			if newCondition.LastTransitionTime.IsZero() {
				newCondition.LastTransitionTime = condition.LastTransitionTime
			}
			condition = newCondition
		}
		conditions = append(conditions, condition)
	}

	if !exist {
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = metav1.Now()
		}
		conditions = append(conditions, newCondition)
	}

	in.Status.Conditions = conditions
}

func (in *MachineSpec) SSH() (*ssh.SSH, error) {
	sshConfig := &ssh.Config{
		User:        in.Machine.Username,
		Host:        in.Machine.IP,
		Port:        int(in.Machine.Port),
		Password:    in.Machine.Password,
		PrivateKey:  in.Machine.PrivateKey,
		PassPhrase:  in.Machine.PassPhrase,
		DialTimeOut: time.Second,
		Retry:       0,
	}
	return ssh.New(sshConfig)
}
