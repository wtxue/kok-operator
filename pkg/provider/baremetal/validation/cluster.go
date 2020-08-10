package validation

import (
	"fmt"
	"net"

	"k8s.io/apimachinery/pkg/util/validation/field"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/common"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/ipallocator"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/validation"
	utilvalidation "github.com/wtxue/kube-on-kube-operator/pkg/util/validation"
)

var (
	nodePodNumAvails        = []int32{16, 32, 64, 128, 256}
	clusterServiceNumAvails = []int32{32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768}
)

// ValidateCluster validates a given Cluster.
func ValidateCluster(obj *common.Cluster) field.ErrorList {
	allErrs := ValidatClusterSpec(&obj.Spec, field.NewPath("spec"), obj.Cluster.Status.Phase)

	return allErrs
}

// ValidatClusterSpec validates a given ClusterSpec.
func ValidatClusterSpec(spec *devopsv1.ClusterSpec, fldPath *field.Path, phase devopsv1.ClusterPhase) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateClusterSpecVersion(spec.Version, fldPath.Child("version"), phase)...)
	allErrs = append(allErrs, ValidateCIDRs(spec, fldPath)...)
	allErrs = append(allErrs, ValidateClusterProperty(spec, fldPath.Child("properties"))...)
	// allErrs = append(allErrs, ValidateClusterMachines(spec.Machines, fldPath.Child("machines"))...)
	// allErrs = append(allErrs, ValidateClusterFeature(&spec.Features, fldPath.Child("features"))...)

	return allErrs
}

// ValidateClusterSpecVersion validates a given version.
func ValidateClusterSpecVersion(version string, fldPath *field.Path, phase devopsv1.ClusterPhase) field.ErrorList {
	allErrs := field.ErrorList{}
	if phase == devopsv1.ClusterInitializing {
		allErrs = utilvalidation.ValidateEnum(version, fldPath, constants.K8sVersions)
	}

	return allErrs
}

// ValidateCIDRs validates clusterCIDR and serviceCIDR.
func ValidateCIDRs(spec *devopsv1.ClusterSpec, specPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	fldPath := specPath.Child("clusterCIDR")
	cidr := spec.ClusterCIDR
	var clusterCIDR *net.IPNet
	if len(cidr) == 0 {
		allErrs = append(allErrs, field.Required(fldPath, ""))
	} else {
		var err error
		_, clusterCIDR, err = net.ParseCIDR(cidr)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, cidr, err.Error()))
		}
	}

	fldPath = specPath.Child("serviceCIDR")
	if spec.ServiceCIDR != nil {
		cidr := *spec.ServiceCIDR
		_, serviceCIDR, err := net.ParseCIDR(cidr)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, cidr, err.Error()))
		} else {
			if err := validation.IsSubNetOverlapped(clusterCIDR, serviceCIDR); err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath, cidr, err.Error()))
			}
			if _, err := ipallocator.GetIndexedIP(serviceCIDR, 10); err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath, cidr,
					"must contains at least 10 ips, because kubeadm need the 10th ip"))
			}
		}
	}

	return allErrs
}

// ValidateClusterProperty validates a given ClusterProperty.
func ValidateClusterProperty(spec *devopsv1.ClusterSpec, propPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	properties := spec.Properties

	fldPath := propPath.Child("maxNodePodNum")
	if properties.MaxNodePodNum == nil {
		allErrs = append(allErrs, field.Required(fldPath, fmt.Sprintf("validate values are %v", nodePodNumAvails)))
	} else {
		allErrs = utilvalidation.ValidateEnum(*properties.MaxNodePodNum, fldPath, nodePodNumAvails)
	}

	fldPath = propPath.Child("maxClusterServiceNum")
	if properties.MaxClusterServiceNum == nil {
		if spec.ServiceCIDR == nil { // not set serviceCIDR, need set maxClusterServiceNum
			allErrs = append(allErrs, field.Required(fldPath, fmt.Sprintf("validate values are %v", clusterServiceNumAvails)))
		}
	} else {
		if spec.ServiceCIDR != nil { // spec.serviceCIDR and properties.maxClusterServiceNum can't be used together
			allErrs = append(allErrs, field.Forbidden(fldPath, "can't be used together with spec.serviceCIDR"))
		} else {
			allErrs = utilvalidation.ValidateEnum(*properties.MaxClusterServiceNum, fldPath, clusterServiceNumAvails)
			if *properties.MaxClusterServiceNum < 10 {
				allErrs = append(allErrs, field.Invalid(fldPath, *properties.MaxClusterServiceNum,
					"must be greater than or equal to 10 because kubeadm need the 10th ip"))
			}
		}
	}

	return allErrs
}
