package gpu

import (
	"context"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	clientset "k8s.io/client-go/kubernetes"
)

type NvidiaDriverOption struct {
}

func InstallNvidiaDriver(s ssh.Interface, option *NvidiaDriverOption) error {

	return nil
}

type NvidiaContainerRuntimeOption struct {
}

func InstallNvidiaContainerRuntime(s ssh.Interface, option *NvidiaContainerRuntimeOption) error {

	return nil
}

type NvidiaDevicePluginOption struct {
	Image string
}

func InstallNvidiaDevicePlugin(ctx context.Context, clientset clientset.Interface, option *NvidiaDevicePluginOption) error {

	return nil
}

func IsEnable(labels map[string]string) bool {
	return labels["nvidia-device-enable"] == "enable"
}

func MachineIsSupport(s ssh.Interface) bool {
	return true
}
