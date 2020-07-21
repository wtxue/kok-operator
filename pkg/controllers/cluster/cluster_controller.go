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

package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pkg/errors"
	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/common"
	"github.com/wtxue/kube-on-kube-operator/pkg/gmanager"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/pkiutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// clusterReconciler reconciles a Cluster object
type clusterReconciler struct {
	client.Client
	*gmanager.GManager
	Log            logr.Logger
	Mgr            manager.Manager
	Scheme         *runtime.Scheme
	ClusterStarted map[string]bool
}

type clusterContext struct {
	Key     types.NamespacedName
	Logger  logr.Logger
	Cluster *devopsv1.Cluster
}

func Add(mgr manager.Manager, pMgr *gmanager.GManager) error {
	reconciler := &clusterReconciler{
		Client:         mgr.GetClient(),
		Mgr:            mgr,
		Log:            ctrl.Log.WithName("controllers").WithName("cluster"),
		Scheme:         mgr.GetScheme(),
		GManager:       pMgr,
		ClusterStarted: make(map[string]bool),
	}

	err := reconciler.SetupWithManager(mgr)
	if err != nil {
		return errors.Wrapf(err, "unable to create cluster controller")
	}

	return nil
}

func (r *clusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.Cluster{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=devops.k8s.io,resources=virtulclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.k8s.io,resources=virtulclusters/status,verbs=get;update;patch

func (r *clusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("cluster", req.NamespacedName.Name)

	startTime := time.Now()
	defer func() {
		diffTime := time.Since(startTime)
		var logLevel klog.Level
		if diffTime > 1*time.Second {
			logLevel = 1
		} else if diffTime > 100*time.Millisecond {
			logLevel = 2
		} else {
			logLevel = 4
		}
		klog.V(logLevel).Infof("##### [%s] reconciling is finished. time taken: %v. ", req.NamespacedName, diffTime)
	}()

	c := &devopsv1.Cluster{}
	err := r.Client.Get(ctx, req.NamespacedName, c)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(4).Info("not find cluster")
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get cluster")
		return reconcile.Result{}, err
	}

	if c.Spec.Pause == true {
		logger.V(4).Info("cluster is Pause")
		return reconcile.Result{}, nil
	}

	if !constants.IsK8sSupport(c.Spec.Version) {
		if c.Status.Phase != devopsv1.ClusterNotSupport {
			c.Status.Phase = devopsv1.ClusterNotSupport
			err = r.Client.Status().Update(ctx, c)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	klog.Infof("name: %s", c.Name)
	if len(string(c.Status.Phase)) == 0 {
		c.Status.Phase = devopsv1.ClusterInitializing
		err = r.Client.Status().Update(ctx, c)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	r.reconcile(ctx, &clusterContext{
		Key:     req.NamespacedName,
		Logger:  logger,
		Cluster: c,
	})
	return ctrl.Result{}, nil
}

func (r *clusterReconciler) addClusterCheck(ctx context.Context, c *common.Cluster) error {
	if _, ok := r.ClusterStarted[c.Cluster.Name]; ok {
		return nil
	}

	if extKubeconfig, ok := c.ClusterCredential.ExtData[pkiutil.ExternalAdminKubeConfigFileName]; ok {
		_, err := r.GManager.AddNewClusters(c.Cluster.Name, extKubeconfig)
		if err != nil {
			klog.Errorf("failed add cluster: %s manager cache", c.Cluster.Name)
			return nil
		}
		klog.Infof("#######  add cluster: %s to manager cache success", c.Cluster.Name)
		r.ClusterStarted[c.Cluster.Name] = true
		return nil
	}

	klog.Warningf("can't find %s", pkiutil.ExternalAdminKubeConfigFileName)
	return nil
}

func (r *clusterReconciler) reconcile(ctx context.Context, rc *clusterContext) error {
	phaseRestore := constants.GetAnnotationKey(rc.Cluster.Annotations, constants.ClusterPhaseRestore)
	if len(phaseRestore) > 0 {
		conditions := make([]devopsv1.ClusterCondition, 0)
		for i := range rc.Cluster.Status.Conditions {
			if rc.Cluster.Status.Conditions[i].Type == phaseRestore {
				break
			} else {
				conditions = append(conditions, rc.Cluster.Status.Conditions[i])
			}
		}
		rc.Cluster.Status.Conditions = conditions
		rc.Cluster.Status.Phase = devopsv1.ClusterInitializing
		err := r.Client.Status().Update(ctx, rc.Cluster)
		if err != nil {
			return err
		}

		objBak := &devopsv1.Cluster{}
		r.Client.Get(ctx, rc.Key, objBak)
		delete(objBak.Annotations, constants.ClusterPhaseRestore)
		err = r.Client.Update(ctx, objBak)
		if err != nil {
			return err
		}
		return nil
	}

	p, err := r.CpManager.GetProvider(rc.Cluster.Spec.Type)
	if err != nil {
		return err
	}

	clusterWrapper, err := common.GetCluster(ctx, r.Client, rc.Cluster, r.ClusterManager)
	if err != nil {
		return err
	}

	switch rc.Cluster.Status.Phase {
	case devopsv1.ClusterInitializing:
		rc.Logger.Info("onCreate")
		r.onCreate(ctx, rc, p, clusterWrapper)
	case devopsv1.ClusterRunning:
		rc.Logger.Info("onUpdate")
		r.addClusterCheck(ctx, clusterWrapper)
		r.onUpdate(ctx, rc, p, clusterWrapper)
	default:
		return fmt.Errorf("no handler for %q", rc.Cluster.Status.Phase)
	}

	return r.applyStatus(ctx, rc, clusterWrapper)
}
