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

	"github.com/pkg/errors"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/gmanager"
	"github.com/wtxue/kok-operator/pkg/provider/phases/clean"
	"github.com/wtxue/kok-operator/pkg/util/pkiutil"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	controllerName = "cluster"
)

// clusterReconciler reconciles a Cluster object
type clusterReconciler struct {
	client.Client
	*gmanager.GManager
	Mgr            manager.Manager
	Scheme         *runtime.Scheme
	ClusterStarted map[string]bool
}

func Add(mgr manager.Manager, pMgr *gmanager.GManager) error {
	reconciler := &clusterReconciler{
		Client:         mgr.GetClient(),
		Mgr:            mgr,
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

// +kubebuilder:rbac:groups=devops.k8s.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.k8s.io,resources=clusters/status,verbs=get;update;patch

func (r *clusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	defer func() {
		diffTime := time.Since(startTime)
		var logLevel int
		if diffTime > 1*time.Second {
			logLevel = 2
		} else if diffTime > 100*time.Millisecond {
			logLevel = 4
		} else {
			logLevel = 5
		}
		logger.V(logLevel).Info("reconcile finished", "time taken", fmt.Sprintf("%v", diffTime))
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

	//
	clusterCtx := &common.ClusterContext{
		Ctx:     ctx,
		Key:     req.NamespacedName,
		Client:  r.Client,
		Logger:  logger,
		Cluster: c,
	}

	if !c.ObjectMeta.DeletionTimestamp.IsZero() {
		err := r.cleanClusterResources(clusterCtx)
		if err != nil {
			logger.Error(err, "failed to clean cluster resources")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if !constants.ContainsString(c.ObjectMeta.Finalizers, constants.FinalizersCluster) {
		logger.V(4).Info("start set", "finalizers", constants.FinalizersCluster)
		if c.ObjectMeta.Finalizers == nil {
			c.ObjectMeta.Finalizers = []string{}
		}
		c.ObjectMeta.Finalizers = append(c.ObjectMeta.Finalizers, constants.FinalizersCluster)
		err := r.Client.Update(ctx, c)
		if err != nil {
			logger.Error(err, "failed to set finalizers")
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	if c.Spec.Pause == true {
		logger.V(4).Info("cluster is Pause")
		return reconcile.Result{}, nil
	}

	if !r.GManager.IsK8sSupport(c.Spec.Version) {
		if c.Status.Phase != devopsv1.ClusterNotSupport {
			logger.V(4).Info("not support", "version", c.Spec.Version)
			c.Status.Phase = devopsv1.ClusterNotSupport
			err = r.Client.Status().Update(ctx, c)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if (len(string(c.Status.Phase)) == 0 || len(c.Status.Conditions) == 0) && c.Status.Phase != devopsv1.ClusterInitializing {
		logger.V(4).Info("change", "status", devopsv1.ClusterInitializing)
		c.Status.Phase = devopsv1.ClusterInitializing
		err = r.Client.Status().Update(ctx, c)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	r.reconcile(clusterCtx)
	return ctrl.Result{}, nil
}

func (r *clusterReconciler) addClusterCheck(ctx *common.ClusterContext) error {
	if _, ok := r.ClusterStarted[ctx.Cluster.Name]; ok {
		return nil
	}

	if extKubeconfig, ok := ctx.Credential.ExtData[pkiutil.ExternalAdminKubeConfigFileName]; ok {
		ctx.Info("add manager extKubeconfig", "file", extKubeconfig)
		_, err := r.GManager.AddNewClusters(ctx.Cluster.Name, extKubeconfig)
		if err != nil {
			ctx.Error(err, "add new clusters manager cache")
			return nil
		}
		ctx.Info("add cluster manager successfully")
		r.ClusterStarted[ctx.Cluster.Name] = true
		return nil
	}

	ctx.Info("can't find  extKubeconfig", "file", pkiutil.ExternalAdminKubeConfigFileName)
	return nil
}

func (r *clusterReconciler) reconcile(ctx *common.ClusterContext) error {
	phaseRestore := constants.GetAnnotationKey(ctx.Cluster.Annotations, constants.ClusterPhaseRestore)
	if len(phaseRestore) > 0 {
		ctx.Info("phase restore", "cluster", ctx.Cluster.Name, "step", phaseRestore)
		conditions := make([]devopsv1.ClusterCondition, 0)
		for i := range ctx.Cluster.Status.Conditions {
			if ctx.Cluster.Status.Conditions[i].Type == phaseRestore {
				break
			} else {
				conditions = append(conditions, ctx.Cluster.Status.Conditions[i])
			}
		}
		ctx.Cluster.Status.Conditions = conditions
		ctx.Cluster.Status.Phase = devopsv1.ClusterInitializing
		err := r.Client.Status().Update(ctx.Ctx, ctx.Cluster)
		if err != nil {
			return err
		}

		objBak := &devopsv1.Cluster{}
		r.Client.Get(ctx.Ctx, ctx.Key, objBak)
		delete(objBak.Annotations, constants.ClusterPhaseRestore)
		err = r.Client.Update(ctx.Ctx, objBak)
		if err != nil {
			return err
		}
		return nil
	}

	p, err := r.CpManager.GetProvider(ctx.Cluster.Spec.ClusterType)
	if err != nil {
		return err
	}

	err = common.FillClusterContext(ctx, r.ClusterManager)
	if err != nil {
		return err
	}

	switch ctx.Cluster.Status.Phase {
	case devopsv1.ClusterInitializing:
		ctx.Info("onCreate")
		r.onCreate(ctx, p)
	case devopsv1.ClusterRunning:
		ctx.Info("onUpdate")
		r.addClusterCheck(ctx)
		r.onUpdate(ctx, p)
	default:
		ctx.Info("cluster status %q unknown", ctx.Cluster.Status.Phase)
		return fmt.Errorf("no handler for status %q", ctx.Cluster.Status.Phase)
	}

	return r.applyStatus(ctx)
}

func (r *clusterReconciler) cleanClusterResources(ctx *common.ClusterContext) error {
	ms := &devopsv1.MachineList{}
	listOptions := &client.ListOptions{Namespace: ctx.Key.Namespace}
	err := r.Client.List(ctx.Ctx, ms, listOptions)
	if err != nil {
		if apierrors.IsNotFound(err) {
			ctx.Info("not find machineList")
		} else {
			ctx.Error(err, "failed to list machine")
			return err
		}
	}

	if started, ok := r.ClusterStarted[ctx.Cluster.Name]; ok && started {
		ctx.Info("start clean with cluster manager")
		r.ClusterManager.Delete(ctx.Cluster.Name)
		delete(r.ClusterStarted, ctx.Cluster.Name)
	}

	credential := &devopsv1.ClusterCredential{}
	err = r.Client.Get(ctx.Ctx, types.NamespacedName{Name: ctx.Cluster.Name, Namespace: ctx.Cluster.Namespace}, credential)
	if err == nil {
		ctx.Info("start clean clusterCredential")
		r.Client.Delete(ctx.Ctx, credential)
	}

	cms := &corev1.ConfigMapList{}
	err = r.Client.List(ctx.Ctx, cms, listOptions)
	if err == nil {
		for i := range cms.Items {
			cm := &cms.Items[i]
			ctx.Info("start clean", "configmap", cm.Name)
			r.Client.Delete(ctx.Ctx, cm)
		}
		r.Client.Delete(ctx.Ctx, credential)
	}

	// clean worker node
	ctx.Info("start clean worker node")
	for i := range ms.Items {
		m := &ms.Items[i]
		ctx.Info("start clean", "machine", m.Name)
		r.Client.Delete(ctx.Ctx, m)
	}

	// clean master node
	ctx.Info("start clean master node")
	for i := range ctx.Cluster.Spec.Machines {
		m := ctx.Cluster.Spec.Machines[i]
		ssh, err := m.SSH()
		if err != nil {
			ctx.Error(err, "failed new ssh", "node", m.IP)
			return err
		}

		ctx.Info("start clean", "machine", m.IP)
		err = clean.CleanNode(ssh)
		if err != nil {
			ctx.Error(err, "failed clean machine node", "node", m.IP)
			return err
		}
	}

	ctx.Info("clean all manchine success, start clean cluster finalizers")
	ctx.Cluster.ObjectMeta.Finalizers = constants.RemoveString(ctx.Cluster.ObjectMeta.Finalizers, constants.FinalizersCluster)
	return r.Client.Update(ctx.Ctx, ctx.Cluster)
}
