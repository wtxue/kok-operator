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

package machine

import (
	"context"
	"fmt"

	"time"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/common"
)

const (
	machineClientRetryCount    = 5
	machineClientRetryInterval = 5 * time.Second

	reasonFailedInit   = "FailedInit"
	reasonFailedUpdate = "FailedUpdate"
)

func (r *machineReconciler) onCreate(ctx context.Context, rc *manchineContext) error {
	p, err := r.MpManager.GetProvider(rc.Cluster.Spec.Type)
	if err != nil {
		return err
	}

	clusterWrapper := &common.Cluster{
		Cluster:           rc.Cluster,
		ClusterCredential: rc.ClusterCredential,
		Client:            r.Client,
		ClusterManager:    r.ClusterManager,
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
	p, err := r.MpManager.GetProvider(rc.Cluster.Spec.Type)
	if err != nil {
		return err
	}

	clusterWrapper := &common.Cluster{
		Cluster:           rc.Cluster,
		ClusterCredential: rc.ClusterCredential,
		Client:            r.Client,
		ClusterManager:    r.ClusterManager,
	}

	err = p.OnUpdate(ctx, rc.Machine, clusterWrapper)
	if err != nil {
		clusterWrapper.Cluster.Status.Message = err.Error()
		clusterWrapper.Cluster.Status.Reason = reasonFailedUpdate
		r.Client.Status().Update(ctx, rc.Cluster)
		return err
	}
	clusterWrapper.Cluster.Status.Message = ""
	clusterWrapper.Cluster.Status.Reason = ""
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
	default:
		err = fmt.Errorf("no handler for %q", rc.Cluster.Status.Phase)
	}

	return err
}
