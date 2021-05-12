package addons

import (
	"context"

	"github.com/go-logr/logr"
	workloadv1 "github.com/wtxue/kok-operator/pkg/apis/workload/v1"
	"github.com/wtxue/kok-operator/pkg/gmanager"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	controllerName = "addons"
)

// addonsReconciler reconciles a addons object
type addonsReconciler struct {
	client.Client
	Log    logr.Logger
	Mgr    manager.Manager
	Scheme *runtime.Scheme
}

type addonsContext struct {
	Ctx    context.Context
	Req    reconcile.Request
	Addons *workloadv1.Addons
	logr.Logger
}

func Add(mgr manager.Manager, pMgr *gmanager.GManager) error {
	r := &addonsReconciler{
		Client: mgr.GetClient(),
		Mgr:    mgr,
		Log:    logf.Log.WithName(controllerName),
		Scheme: mgr.GetScheme(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&workloadv1.Addons{}).
		Complete(r)
}

func (r *addonsReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Log.WithValues(controllerName, req.NamespacedName.String())

	addons := &workloadv1.Addons{}
	err := r.Client.Get(ctx, req.NamespacedName, addons)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "not find addons")
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get addons")
		return reconcile.Result{}, err
	}

	return r.reconcile(&addonsContext{Ctx: ctx, Req: req, Addons: addons, Logger: logger})
}

func (r *addonsReconciler) reconcile(ctx *addonsContext) (reconcile.Result, error) {

	return reconcile.Result{}, nil
}
