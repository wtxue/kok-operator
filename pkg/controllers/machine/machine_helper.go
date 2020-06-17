package machine

import (
	"context"
	"fmt"

	"time"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider"
	machineprovider "github.com/wtxue/kube-on-kube-operator/pkg/provider/machine"
)

const (
	machineClientRetryCount    = 5
	machineClientRetryInterval = 5 * time.Second

	reasonFailedInit   = "FailedInit"
	reasonFailedUpdate = "FailedUpdate"
)

func (r *machineReconciler) onCreate(ctx context.Context, rc *manchineContext) error {
	p, err := machineprovider.GetProvider(rc.Cluster.Spec.Type)
	if err != nil {
		return err
	}

	clusterWrapper := &provider.Cluster{
		Cluster:           rc.Cluster,
		ClusterCredential: rc.ClusterCredential,
	}
	err = p.OnCreate(ctx, rc.Machine, clusterWrapper)
	if err != nil {
		rc.Machine.Status.Message = err.Error()
		rc.Machine.Status.Reason = reasonFailedInit
		r.Client.Status().Update(ctx, rc.Machine)
		return err
	}

	condition := rc.Machine.Status.Conditions[len(rc.Machine.Status.Conditions)-1]
	if condition.Status == devopsv1.ConditionFalse { // means current condition run into error
		rc.Machine.Status.Message = condition.Message
		rc.Machine.Status.Reason = condition.Reason
		r.Client.Status().Update(ctx, rc.Machine)
		return fmt.Errorf("Provider.OnCreate.%s [Failed] reason: %s message: %s",
			condition.Type, condition.Reason, condition.Message)
	}

	rc.Machine.Status.Message = ""
	rc.Machine.Status.Reason = ""
	err = r.Client.Status().Update(ctx, rc.Machine)
	if err != nil {
		return err
	}
	return nil
}

func (r *machineReconciler) onUpdate(ctx context.Context, rc *manchineContext) error {
	p, err := machineprovider.GetProvider(rc.Cluster.Spec.Type)
	if err != nil {
		return err
	}

	clusterWrapper := &provider.Cluster{
		Cluster:           rc.Cluster,
		ClusterCredential: rc.ClusterCredential,
	}

	err = p.OnUpdate(ctx, rc.Machine, clusterWrapper)
	if err != nil {
		clusterWrapper.Status.Message = err.Error()
		clusterWrapper.Status.Reason = reasonFailedUpdate
		r.Client.Status().Update(ctx, rc.Cluster)
		return err
	}
	clusterWrapper.Status.Message = ""
	clusterWrapper.Status.Reason = ""
	r.Client.Status().Update(ctx, clusterWrapper.ClusterCredential)
	r.Client.Status().Update(ctx, clusterWrapper.Cluster)
	return nil
}

func (r *machineReconciler) reconcile(ctx context.Context, rc *manchineContext) error {
	var err error
	switch rc.Machine.Status.Phase {
	case devopsv1.MachineInitializing:
		rc.Logger.Info("onCreate")
		err = r.onCreate(ctx, rc)
	case devopsv1.MachineRunning:
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
