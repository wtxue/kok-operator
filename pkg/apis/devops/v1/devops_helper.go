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

const (
	ClusterAnnotationAction = "k8s.io/action"
)

// ClusterMachine is the master machine definition of cluster.
type ClusterMachine struct {
	IP       string `json:"ip" protobuf:"bytes,1,opt,name=ip"`
	Port     int32  `json:"port" protobuf:"varint,2,opt,name=port"`
	Username string `json:"username" protobuf:"bytes,3,opt,name=username"`
	// +optional
	Password string `json:"password,omitempty" protobuf:"bytes,4,opt,name=password"`
	// +optional
	PrivateKey []byte `json:"privateKey,omitempty" protobuf:"bytes,5,opt,name=privateKey"`
	// +optional
	PassPhrase []byte `json:"passPhrase,omitempty" protobuf:"bytes,6,opt,name=passPhrase"`
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,7,opt,name=labels"`
	// If specified, the node's taints.
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty" protobuf:"bytes,8,opt,name=taints"`
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
