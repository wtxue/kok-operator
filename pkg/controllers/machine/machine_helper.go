package machine

import (
	"fmt"
	"time"

	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
)

const (
	machineClientRetryCount    = 5
	machineClientRetryInterval = 5 * time.Second

	reasonFailedInit   = "FailedInit"
	reasonFailedUpdate = "FailedUpdate"
)

func (r *machineReconciler) onCreate(ctx *common.ClusterContext, machine *devopsv1.Machine) error {
	p, err := r.MpManager.GetProvider(ctx.Cluster.Spec.ClusterType)
	if err != nil {
		return err
	}

	err = p.OnCreate(ctx, machine)
	if err != nil {
		machine.Status.Message = err.Error()
		machine.Status.Reason = reasonFailedInit
		r.Client.Status().Update(ctx.Ctx, machine)
		return err
	}

	condition := machine.Status.Conditions[len(machine.Status.Conditions)-1]
	if condition.Status == devopsv1.ConditionFalse { // means current condition run into error
		machine.Status.Message = condition.Message
		machine.Status.Reason = condition.Reason
		r.Client.Status().Update(ctx.Ctx, machine)
		return fmt.Errorf("Provider.OnCreate.%s [Failed] reason: %s message: %s",
			condition.Type, condition.Reason, condition.Message)
	}

	machine.Status.Message = ""
	machine.Status.Reason = ""
	err = r.Client.Status().Update(ctx.Ctx, machine)
	if err != nil {
		return err
	}
	return nil
}

func (r *machineReconciler) onUpdate(ctx *common.ClusterContext, machine *devopsv1.Machine) error {
	p, err := r.MpManager.GetProvider(ctx.Cluster.Spec.ClusterType)
	if err != nil {
		return err
	}

	err = p.OnUpdate(ctx, machine)
	if err != nil {
		ctx.Cluster.Status.Message = err.Error()
		ctx.Cluster.Status.Reason = reasonFailedUpdate
		r.Client.Status().Update(ctx.Ctx, ctx.Cluster)
		return err
	}
	ctx.Cluster.Status.Message = ""
	ctx.Cluster.Status.Reason = ""
	r.Client.Status().Update(ctx.Ctx, ctx.Credential)
	r.Client.Status().Update(ctx.Ctx, ctx.Cluster)
	return nil
}

func (r *machineReconciler) reconcile(rc *manchineContext) error {
	ctx := &common.ClusterContext{
		Ctx:            rc.Ctx,
		Cluster:        rc.Cluster,
		Credential:     rc.ClusterCredential,
		Client:         r.Client,
		ClusterManager: r.ClusterManager,
		Logger:         rc.Logger,
	}

	var err error
	switch rc.Machine.Status.Phase {
	case devopsv1.MachineInitializing:
		rc.Logger.Info("onCreate")
		err = r.onCreate(ctx, rc.Machine)
	case devopsv1.MachineRunning:
		rc.Logger.Info("onUpdate")
		err = r.onUpdate(ctx, rc.Machine)
	default:
		err = fmt.Errorf("no handler for %q", rc.Cluster.Status.Phase)
	}

	return err
}
