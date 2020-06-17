package apiclient

import (
	"fmt"

	"github.com/Masterminds/semver"

	"k8s.io/client-go/kubernetes"
)

func ClusterVersionIsBefore19(client kubernetes.Interface) bool {
	result, err := CheckClusterVersion(client, "< 1.9")
	if err != nil {
		return false
	}

	return result
}

func CheckClusterVersion(client kubernetes.Interface, versionConstraint string) (bool, error) {
	version, err := GetClusterVersion(client)
	if err != nil {
		return false, err
	}

	return CheckVersion(version, versionConstraint)
}

func GetClusterVersion(client kubernetes.Interface) (string, error) {
	version, err := client.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v.%v", version.Major, version.Minor), nil
}

func CheckVersion(version string, versionConstraint string) (bool, error) {
	c, err := semver.NewConstraint(versionConstraint)
	if err != nil {
		return false, err
	}
	v, err := semver.NewVersion(version)
	if err != nil {
		return false, err
	}

	return c.Check(v), nil
}
