package commands

import (
	"time"

	"context"
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sync"
)

// GenRandTime ...
func GenRandTime(periodBase, randInterval int64) time.Duration {
	return time.Duration(rand.Int63()%randInterval+periodBase) * time.Second
}

// NewToolsCmd ...
func NewToolsCmd(opt *GlobalManagerOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "tools for component",
		Run: func(cmd *cobra.Command, _ []string) {
			opt.SetupLogger()
			cfg, err := opt.GetK8sConfig()
			if err != nil {
				klog.Fatalf("unable to get cfg err: %v", err)
			}

			cfg.QPS = float32(500)
			cfg.Burst = 1000
			cfg.Timeout = 30 * time.Second
			cfg.UserAgent = rest.DefaultKubernetesUserAgent() + "/" + "tools"

			logger := logf.Log.WithName("tools")
			// kubeCli, err := kubernetes.NewForConfig(cfg)
			// if err != nil {
			// 	klog.Fatalf("unable to new kubecli err: %v", err)
			// }

			// ciliumCli, err := clientset.NewForConfig(cfg)
			// if err != nil {
			// 	klog.Fatalf("unable to new ciliumCli err: %v", err)
			// }

			mapper, err := apiutil.NewDynamicRESTMapper(cfg)
			if err != nil {
				logger.Error(err, "failed to new rest mapper")
				klog.Fatalf("failed to new rest mapper err: %v", err)
			}

			sch := GetScheme()
			_ = ciliumv2.AddToScheme(sch)
			kcli, err := client.New(cfg, client.Options{
				Scheme: sch,
				Mapper: mapper,
			})
			if err != nil {
				klog.Fatalf("failed apiReader err: %v", err)
			}

			// nodeList, err := kubeCli.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{ResourceVersion: "0"})
			// if err != nil {
			// 	klog.Fatalf("unable to new nodeList err: %v", err)
			// }

			stopCh := signals.SetupSignalHandler()

			var wg sync.WaitGroup

			// for i := 0; i < 2; i++ {
			// 	wg.Add(1)
			// 	logger.Info("start go routines list", "index", i)
			// 	go wait.Until(func() {
			// 		cep := &ciliumv2.CiliumEndpointList{}
			// 		cnode := &ciliumv2.CiliumNodeList{}
			//
			// 		err := kcli.List(context.Background(), cep)
			// 		if err != nil {
			// 			logger.Error(err, "list CiliumEndpoint")
			// 		}
			//
			// 		time.Sleep(time.Second * 10)
			// 		nerr := kcli.List(context.Background(), cnode)
			// 		if nerr != nil {
			// 			logger.Error(nerr, "list CiliumNodeList")
			// 		}
			// 	}, time.Second*30, stopCh.Done())
			// }

			time.Sleep(GenRandTime(10, 20))
			for i := 0; i < 1; i++ {
				wg.Add(1)
				logger.Info("start go routines list", "index", i)
				go wait.Until(func() {
					cnode := &corev1.NodeList{}
					cep := &corev1.PodList{}

					err := kcli.List(context.Background(), cep)
					if err != nil {
						logger.Error(err, "list pod")
					}

					time.Sleep(GenRandTime(10, 20))
					nerr := kcli.List(context.Background(), cnode)
					if nerr != nil {
						logger.Error(nerr, "list node")
					}
				}, GenRandTime(30, 30), stopCh.Done())
			}

			// for i := range nodeList.Items {
			// 	node := &nodeList.Items[i]
			// 	wg.Add(1)
			// 	go func(nodeName string) {
			// 		nlog := log.WithValues("node", nodeName)
			// 		nlog.Info("start test ......", "cnt", cnt)
			// 		for i := 0; i < cnt; i++ {
			// 			f := func() {
			// 				// ctxPod, podfunc := context.WithTimeout(stopCh, 5*time.Second)
			// 				// defer podfunc()
			//
			// 				options := metav1.ListOptions{
			// 					ResourceVersion: "0",
			// 				}
			//
			// 				// options.FieldSelector = fields.OneTermEqualSelector("spec.nodeName", node.Name).String()
			// 				podlist, err := kubeCli.CoreV1().Pods(corev1.NamespaceAll).List(context.Background(), options)
			// 				if err != nil {
			// 					nlog.Error(err, "failed list pods ")
			// 					return
			// 				}
			//
			// 				nlog.V(4).Info("podlist", "len", len(podlist.Items))
			// 			}
			//
			// 			f()
			// 			time.Sleep(5 * time.Second)
			// 		}
			//
			// 		wg.Done()
			// 	}(node.Name)
			// }

			wg.Wait()
			logger.Info("All go routines finished executing")
			select {
			case <-stopCh.Done():
			}
		},
	}

	return cmd
}
