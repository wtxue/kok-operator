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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// clusterReconciler reconciles a Cluster object
type clusterReconciler struct {
	client.Client
	Log    logr.Logger
	Mgr    manager.Manager
	Scheme *runtime.Scheme
}

type clusterContext struct {
	Key     types.NamespacedName
	Logger  logr.Logger
	Cluster *devopsv1.Cluster
}

func Add(mgr manager.Manager) error {
	reconciler := &clusterReconciler{
		Client: mgr.GetClient(),
		Mgr:    mgr,
		Log:    ctrl.Log.WithName("controllers").WithName("cluster"),
		Scheme: mgr.GetScheme(),
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

func (r *clusterReconciler) reconcile(ctx context.Context, rc *clusterContext) error {
	var err error
	switch rc.Cluster.Status.Phase {
	case devopsv1.ClusterInitializing:
		rc.Logger.Info("onCreate")
		err = r.onCreate(ctx, rc)
	case devopsv1.ClusterRunning:
		rc.Logger.Info("onUpdate")
		err = r.onUpdate(ctx, rc)
		if err == nil {
			// c.ensureHealthCheck(ctx, key, cluster) // after update to avoid version conflict
		}
	default:
		err = fmt.Errorf("no handler for %q", rc.Cluster.Status.Phase)
	}

	return nil
}
