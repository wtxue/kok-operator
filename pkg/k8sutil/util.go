package k8sutil

import (
	"fmt"
	"math"
	"net"

	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/util/ipallocator"

	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func StrPointer(s string) *string {
	return &s
}

func IntPointer(i int32) *int32 {
	return &i
}

func Int64Pointer(i int64) *int64 {
	return &i
}

func BoolPointer(b bool) *bool {
	return &b
}

func PointerToBool(flag *bool) bool {
	if flag == nil {
		return false
	}

	return *flag
}

func PointerToString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func PointerToInt32(i *int32) int32 {
	if i == nil {
		return 0
	}

	return *i
}

func IntstrPointer(i int) *intstr.IntOrString {
	is := intstr.FromInt(i)
	return &is
}

func MergeStringMaps(l map[string]string, l2 map[string]string) map[string]string {
	merged := make(map[string]string)
	if l == nil {
		l = make(map[string]string)
	}
	for lKey, lValue := range l {
		merged[lKey] = lValue
	}
	for lKey, lValue := range l2 {
		merged[lKey] = lValue
	}
	return merged
}

func MergeMultipleStringMaps(stringMaps ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, stringMap := range stringMaps {
		merged = MergeStringMaps(merged, stringMap)
	}
	return merged
}

func EmptyTypedStrSlice(s ...string) []interface{} {
	ret := make([]interface{}, len(s))
	for i := 0; i < len(s); i++ {
		ret[i] = s[i]
	}
	return ret
}

func EmptyTypedFloatSlice(f ...float64) []interface{} {
	ret := make([]interface{}, len(f))
	for i := 0; i < len(f); i++ {
		ret[i] = f[i]
	}
	return ret
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func ObjectMeta(name string, labels map[string]string, obj client.Object) metav1.ObjectMeta {
	ovk := obj.GetObjectKind().GroupVersionKind()

	return metav1.ObjectMeta{
		Name:      name,
		Namespace: obj.GetNamespace(),
		Labels:    labels,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion:         ovk.GroupVersion().String(),
				Kind:               ovk.Kind,
				Name:               obj.GetName(),
				UID:                obj.GetUID(),
				Controller:         BoolPointer(true),
				BlockOwnerDeletion: BoolPointer(true),
			},
		},
	}
}

func GetNodeCIDRMaskSize(clusterCIDR string, maxNodePodNum int32) (int32, error) {
	if maxNodePodNum <= 0 {
		return 0, errors.New("maxNodePodNum must more than 0")
	}
	_, svcSubnetCIDR, err := net.ParseCIDR(clusterCIDR)
	if err != nil {
		return 0, errors.Wrap(err, "ParseCIDR error")
	}

	nodeCidrOccupy := math.Ceil(math.Log2(float64(maxNodePodNum)))
	nodeCIDRMaskSize := 32 - int(nodeCidrOccupy)
	ones, _ := svcSubnetCIDR.Mask.Size()
	if ones > nodeCIDRMaskSize {
		return 0, errors.New("clusterCIDR IP size is less than maxNodePodNum")
	}

	return int32(nodeCIDRMaskSize), nil
}

func GetServiceCIDRAndNodeCIDRMaskSize(clusterCIDR string, maxClusterServiceNum int32, maxNodePodNum int32) (string, int32, error) {
	if maxClusterServiceNum <= 0 || maxNodePodNum <= 0 {
		return "", 0, errors.New("maxClusterServiceNum or maxNodePodNum must more than 0")
	}
	_, svcSubnetCIDR, err := net.ParseCIDR(clusterCIDR)
	if err != nil {
		return "", 0, errors.Wrap(err, "ParseCIDR error")
	}

	size := ipallocator.RangeSize(svcSubnetCIDR)
	if int32(size) < maxClusterServiceNum {
		return "", 0, errors.New("clusterCIDR IP size is less than maxClusterServiceNum")
	}
	lastIP, err := ipallocator.GetIndexedIP(svcSubnetCIDR, int(size-1))
	if err != nil {
		return "", 0, errors.Wrap(err, "get last IP error")
	}

	maskSize := int(math.Ceil(math.Log2(float64(maxClusterServiceNum))))
	_, serviceCidr, _ := net.ParseCIDR(fmt.Sprintf("%s/%d", lastIP.String(), 32-maskSize))

	nodeCidrOccupy := math.Ceil(math.Log2(float64(maxNodePodNum)))
	nodeCIDRMaskSize := 32 - int(nodeCidrOccupy)
	ones, _ := svcSubnetCIDR.Mask.Size()
	if ones > nodeCIDRMaskSize {
		return "", 0, errors.New("clusterCIDR IP size is less than maxNodePodNum")
	}

	return serviceCidr.String(), int32(nodeCIDRMaskSize), nil
}

func GetIndexedIP(subnet string, index int) (net.IP, error) {
	_, svcSubnetCIDR, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't parse service subnet CIDR %q", subnet)
	}

	dnsIP, err := ipallocator.GetIndexedIP(svcSubnetCIDR, index)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get %dth IP address from service subnet CIDR %s", index, svcSubnetCIDR.String())
	}

	return dnsIP, nil
}

// GetAPIServerCertSANs returns extra APIServer's certSANs need to pass kubeadm
func GetAPIServerCertSANs(cls *devopsv1.Cluster) []string {
	certSANs := sets.NewString("127.0.0.1", "localhost")

	if cls.Spec.Features.HA != nil {
		if cls.Spec.Features.HA.KubeHA != nil {
			certSANs.Insert(cls.Spec.Features.HA.KubeHA.VIP)
		}
		if cls.Spec.Features.HA.ThirdPartyHA != nil {
			certSANs.Insert(cls.Spec.Features.HA.ThirdPartyHA.VIP)
		}
	}

	for _, address := range cls.Status.Addresses {
		certSANs.Insert(address.Host)
	}

	svcName := constants.GenComponentName(cls.GetName(), constants.KubeApiServer)
	certSANs.Insert(svcName)

	svcNameWithNs := fmt.Sprintf("%s.%s", svcName, cls.GetNamespace())
	certSANs.Insert(svcNameWithNs)

	certSANs = certSANs.Insert(cls.Spec.PublicAlternativeNames...)
	return certSANs.List()
}
