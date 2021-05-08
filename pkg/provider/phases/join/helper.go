package join

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/apis"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	kubeletv1beta1 "k8s.io/kubelet/config/v1beta1"
	kubeadmv1beta2 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// GetGenericImage generates and returns a platform agnostic image (backed by manifest list)
func GetGenericImage(prefix, image, tag string) string {
	return fmt.Sprintf("%s/%s:%s", prefix, image, tag)
}

// GetPauseImage returns the image for the "pause" container
func GetPauseImage(imageRepository string) string {
	return GetGenericImage(imageRepository, "pause", constants.PauseVersion)
}

// BuildArgumentListFromMap takes two string-string maps, one with the base arguments and one
// with optional override arguments. In the return list override arguments will precede base arguments
func BuildArgumentListFromMap(baseArguments map[string]string, overrideArguments map[string]string) []string {
	var command []string
	var keys []string

	argsMap := make(map[string]string)

	for k, v := range baseArguments {
		argsMap[k] = v
	}

	for k, v := range overrideArguments {
		argsMap[k] = v
	}

	for k := range argsMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		command = append(command, fmt.Sprintf("--%s=%s", k, argsMap[k]))
	}

	return command
}

// GetHostname returns OS's hostname if 'hostnameOverride' is empty; otherwise, return 'hostnameOverride'
// NOTE: This function copied from pkg/util/node package to avoid external kubeadm dependency
func GetHostname(hostnameOverride string) (string, error) {
	hostName := hostnameOverride
	if len(hostName) == 0 {
		nodeName, err := os.Hostname()
		if err != nil {
			return "", errors.Wrap(err, "couldn't determine hostname")
		}
		hostName = nodeName
	}

	// Trim whitespaces first to avoid getting an empty hostname
	// For linux, the hostname is read from file /proc/sys/kernel/hostname directly
	hostName = strings.TrimSpace(hostName)
	if len(hostName) == 0 {
		return "", errors.New("empty hostname is invalid")
	}

	return strings.ToLower(hostName), nil
}

// GetNodeNameAndHostname obtains the name for this Node using the following precedence
// (from lower to higher):
// - actual hostname
// - NodeRegistrationOptions.Name (same as "--node-name" passed to "kubeadm init/join")
// - "hostname-overide" flag in NodeRegistrationOptions.KubeletExtraArgs
// It also returns the hostname or an error if getting the hostname failed.
func GetNodeNameAndHostname(cfg *kubeadmv1beta2.NodeRegistrationOptions) (string, string, error) {
	hostname, err := GetHostname("")
	nodeName := hostname
	if cfg.Name != "" {
		nodeName = cfg.Name
	}
	if name, ok := cfg.KubeletExtraArgs["hostname-override"]; ok {
		nodeName = name
	}
	return nodeName, hostname, err
}

func BuildKubeletDynamicEnvFile(imageRepository string, nodeReg *kubeadmv1beta2.NodeRegistrationOptions) string {
	kubeletFlags := map[string]string{}

	kubeletFlags["network-plugin"] = "cni"
	kubeletFlags["container-runtime"] = "remote"
	kubeletFlags["container-runtime-endpoint"] = "unix:///var/run/containerd/containerd.sock"

	// Pass the "--hostname-override" flag to the kubelet only if it's different from the hostname
	nodeName, hostname, err := GetNodeNameAndHostname(nodeReg)
	if err != nil {
		logf.Log.Error(err, "GetNodeNameAndHostname")
	}

	if nodeName != hostname {
		logf.Log.V(2).Info("setting kubelet args", "hostname-override", nodeName)
		kubeletFlags["hostname-override"] = nodeName
	}

	kubeletFlags["pod-infra-container-image"] = GetPauseImage(imageRepository)
	argList := BuildArgumentListFromMap(kubeletFlags, nodeReg.KubeletExtraArgs)
	envFileContent := fmt.Sprintf("%s=%q\n", constants.KubeletEnvFileVariableName, strings.Join(argList, " "))

	return envFileContent
}

func getKubeletConfiguration(ctx *common.ClusterContext) *kubeletv1beta1.KubeletConfiguration {
	return &kubeletv1beta1.KubeletConfiguration{
		StaticPodPath: constants.KubeletPodManifestDir,
		Authentication: kubeletv1beta1.KubeletAuthentication{
			X509: kubeletv1beta1.KubeletX509Authentication{
				ClientCAFile: constants.CACertName,
			},
			Webhook: kubeletv1beta1.KubeletWebhookAuthentication{
				Enabled: k8sutil.BoolPointer(true),
			},
			Anonymous: kubeletv1beta1.KubeletAnonymousAuthentication{
				Enabled: k8sutil.BoolPointer(false),
			},
		},
		Authorization: kubeletv1beta1.KubeletAuthorization{
			Mode:    kubeletv1beta1.KubeletAuthorizationModeWebhook,
			Webhook: kubeletv1beta1.KubeletWebhookAuthorization{},
		},
		ClusterDNS:    []string{ctx.Cluster.Status.DNSIP},
		ClusterDomain: ctx.Cluster.Spec.DNSDomain,
		KubeReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
		SystemReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
	}
}

func KubeletMarshal(cfg *kubeletv1beta1.KubeletConfiguration) ([]byte, error) {
	gvks, _, err := apis.GetScheme().ObjectKinds(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "kubelet config get gvks")
	}

	yamlData, err := apis.MarshalToYAML(cfg, gvks[0].GroupVersion())
	if err != nil {
		return nil, errors.Wrap(err, "kubelet config marshal")
	}

	return yamlData, nil
}
