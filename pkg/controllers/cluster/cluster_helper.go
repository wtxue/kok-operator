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
	"time"

	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/provider/cluster"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
)

const (
	clusterClientRetryCount    = 5
	clusterClientRetryInterval = 5 * time.Second

	reasonFailedInit   = "FailedInit"
	reasonFailedUpdate = "FailedUpdate"
)

func (r *clusterReconciler) applyStatus(ctx *common.ClusterContext) error {
	credential := &devopsv1.ClusterCredential{}
	err := r.Client.Get(ctx.Ctx, ctx.Key, credential)
	if err != nil {
		if apierrors.IsNotFound(err) {
			ctx.Error(err, "not find cluster credential")
			return nil
		}

		ctx.Error(err, "failed to get cluster credential")
		return err
	}

	if !equality.Semantic.DeepEqual(credential.CredentialInfo, ctx.Credential.CredentialInfo) {
		metaAccessor := meta.NewAccessor()
		currentResourceVersion, err := metaAccessor.ResourceVersion(credential)
		if err != nil {
			ctx.Error(err, "failed to metaAccessor")
			return err
		}
		metaAccessor.SetResourceVersion(ctx.Credential, currentResourceVersion)
		err = r.Client.Update(ctx.Ctx, ctx.Credential)
		if err != nil {
			ctx.Error(err, "failed to update cluster credential")
			return err
		}
		ctx.V(4).Info("update cluster credential success")
	}

	c := &devopsv1.Cluster{}
	err = r.Client.Get(ctx.Ctx, ctx.Key, c)
	if err != nil {
		if apierrors.IsNotFound(err) {
			ctx.Error(err, "not find cluster")
			return nil
		}

		ctx.Error(err, "failed to get cluster")
		return err
	}

	if !equality.Semantic.DeepEqual(c.Status, ctx.Cluster.Status) {
		metaAccessor := meta.NewAccessor()
		currentResourceVersion, err := metaAccessor.ResourceVersion(c)
		if err != nil {
			ctx.Error(err, "failed to metaAccessor")
			return err
		}

		metaAccessor.SetResourceVersion(ctx.Cluster, currentResourceVersion)
		err = r.Client.Status().Update(ctx.Ctx, ctx.Cluster)
		if err != nil {
			ctx.Error(err, "failed to update cluster status")
			return err
		}

		ctx.Info("update cluster status success")
	}

	return nil
}

func (r *clusterReconciler) onCreate(ctx *common.ClusterContext, p cluster.Provider) error {
	err := p.OnCreate(ctx)
	if err != nil {
		ctx.Cluster.Status.Message = err.Error()
		ctx.Cluster.Status.Reason = reasonFailedInit
	} else {
		condition := ctx.Cluster.Status.Conditions[len(ctx.Cluster.Status.Conditions)-1]
		if condition.Status == devopsv1.ConditionFalse { // means current condition run into error
			ctx.Cluster.Status.Message = condition.Message
			ctx.Cluster.Status.Reason = condition.Reason
		} else {
			ctx.Cluster.Status.Message = ""
			ctx.Cluster.Status.Reason = ""
		}
	}

	return nil
}

func (r *clusterReconciler) onUpdate(ctx *common.ClusterContext, p cluster.Provider) error {
	err := p.OnUpdate(ctx)
	if err != nil {
		ctx.Cluster.Status.Message = err.Error()
		ctx.Cluster.Status.Reason = reasonFailedUpdate
	} else {
		ctx.Cluster.Status.Message = ""
		ctx.Cluster.Status.Reason = ""
	}

	return nil
}
