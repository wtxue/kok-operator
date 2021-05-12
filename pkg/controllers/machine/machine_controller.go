package machine

import (
	"context"
	"fmt"
	"time"

	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/gmanager"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	controllerName = "machine"
)

// machineReconciler reconciles a machine object
type machineReconciler struct {
	client.Client
	Mgr    manager.Manager
	Scheme *runtime.Scheme
	Log    logr.Logger
	*gmanager.GManager
}

type manchineContext struct {
	Ctx context.Context
	Key types.NamespacedName
	*devopsv1.Cluster
	*devopsv1.Machine
	*devopsv1.ClusterCredential
	logr.Logger
}

func Add(mgr manager.Manager, pMgr *gmanager.GManager) error {
	reconciler := &machineReconciler{
		Client:   mgr.GetClient(),
		Mgr:      mgr,
		Scheme:   mgr.GetScheme(),
		Log:      logf.Log.WithName(controllerName),
		GManager: pMgr,
	}

	err := reconciler.SetupWithManager(mgr)
	if err != nil {
		return errors.Wrapf(err, "unable to create machine controller")
	}

	return nil
}

func (r *machineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.Machine{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=devops.fake.io,resources=virtulclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.fake.io,resources=virtulclusters/status,verbs=get;update;patch

func (r *machineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("machine", req.Name)
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

	m := &devopsv1.Machine{}
	err := r.Client.Get(ctx, req.NamespacedName, m)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "not find machine")
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get machine")
		return reconcile.Result{}, err
	}

	logger = logger.WithValues("cluster", m.Spec.ClusterName)
	if m.Spec.Pause == true {
		logger.Info("machine is Pause")
		return reconcile.Result{}, nil
	}

	if len(string(m.Status.Phase)) == 0 {
		m.Status.Phase = devopsv1.MachineInitializing
		err = r.Client.Status().Update(ctx, m)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	cluster := &devopsv1.Cluster{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: m.Spec.ClusterName, Namespace: m.Namespace}, cluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "not find cluster")
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get cluster")
		return reconcile.Result{}, err
	}

	if cluster.Status.Phase != devopsv1.ClusterRunning {
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: 30 * time.Second,
		}, nil
	}

	credential := &devopsv1.ClusterCredential{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: m.Spec.ClusterName, Namespace: m.Namespace}, credential)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("not find ClusterCredential", "cluster", m.Spec.ClusterName)
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get ClusterCredential", "cluster", m.Spec.ClusterName)
		return reconcile.Result{}, err
	}

	r.reconcile(&manchineContext{
		Ctx:               ctx,
		Key:               req.NamespacedName,
		Logger:            logger,
		Machine:           m,
		Cluster:           cluster,
		ClusterCredential: credential,
	})
	return ctrl.Result{}, nil
}
