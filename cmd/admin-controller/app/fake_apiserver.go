package app

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wtxue/kube-on-kube-operator/cmd/admin-controller/app/app_option"
	"github.com/wtxue/kube-on-kube-operator/pkg/apiserver"
	"github.com/wtxue/kube-on-kube-operator/pkg/option"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// CreateKubeConfigFile creates a kubeconfig file.
func CreateKubeConfigFile(outDir string, kubeConfigFileName string, cfg *rest.Config) error {
	klog.Infof("creating kubeconfig file for %s", kubeConfigFileName)

	clusterName := "fake-cluster"
	userName := "devops"

	contextName := fmt.Sprintf("%s@%s", userName, clusterName)
	apiConfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: {
				Server: cfg.Host,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  clusterName,
				AuthInfo: userName,
			},
		},
		AuthInfos:      map[string]*clientcmdapi.AuthInfo{},
		CurrentContext: contextName,
	}

	apiConfig.AuthInfos[userName] = &clientcmdapi.AuthInfo{}

	kubeConfigFilePath := filepath.Join(outDir, kubeConfigFileName)
	err := clientcmd.WriteToFile(*apiConfig, kubeConfigFilePath)
	if err != nil {
		return err
	}

	klog.Infof("kubeconfig file for [%s@%s] is write to path: %s", clusterName, userName, kubeConfigFilePath)
	return nil
}

func tryRun(opt *option.ApiServerOption, stopCh <-chan struct{}) error {
	if opt.IsLocalKube {
		svc := apiserver.New(opt)
		cfg, err := svc.Start(stopCh)
		if err != nil {
			return err
		}

		err = CreateKubeConfigFile("k8s/cfg", "fake-kubeconfig.yaml", cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewFakeApiserverCmd(opt *app_option.Options) *cobra.Command {
	var isStart bool
	apiServerOpt := option.DefaultApiServerOption()

	cmd := &cobra.Command{
		Use:   "fake",
		Short: "Manage with a fake apiserver",
		Run: func(cmd *cobra.Command, args []string) {
			stopCh := signals.SetupSignalHandler()
			err := tryRun(apiServerOpt, stopCh)
			if err != nil {
				klog.Errorf("err: %+v", err)
				return
			}

			<-stopCh
			klog.Infof("stop fake api server")
		},
	}

	cmd.Flags().BoolVarP(&isStart, "start", "s", true, "Enables start fake apiserver")
	apiServerOpt.AddFlags(cmd.Flags())
	return cmd
}
