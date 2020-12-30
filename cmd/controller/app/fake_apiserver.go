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

package app

import (
	"fmt"
	"path/filepath"

	"context"

	"github.com/spf13/cobra"
	"github.com/wtxue/kok-operator/cmd/controller/app/app_option"
	"github.com/wtxue/kok-operator/pkg/apiserver"
	"github.com/wtxue/kok-operator/pkg/option"
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

func tryRun(opt *option.ApiServerOption, ctx context.Context) error {
	if opt.IsLocalKube {
		svc := apiserver.New(opt)
		cfg, err := svc.Start(ctx.Done())
		if err != nil {
			return err
		}

		err = CreateKubeConfigFile(opt.RootDir+"/cfg", "fake-kubeconfig.yaml", cfg)
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
			ctx := signals.SetupSignalHandler()
			err := tryRun(apiServerOpt, ctx)
			if err != nil {
				klog.Errorf("err: %+v", err)
				return
			}

			<-ctx.Done()
			klog.Infof("stop fake api server")
		},
	}

	cmd.Flags().BoolVarP(&isStart, "start", "s", true, "Enables start fake apiserver")
	apiServerOpt.AddFlags(cmd.Flags())
	return cmd
}
