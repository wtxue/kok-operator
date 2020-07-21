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

func (r *clusterReconciler) applyStatus(ctx context.Context, rc *clusterContext, cluster *common.Cluster) error {
	credential := &devopsv1.ClusterCredential{}
	err := r.Client.Get(ctx, rc.Key, credential)
	if err != nil {
		if apierrors.IsNotFound(err) {
			rc.Logger.Error(err, "not find cluster credential")
			return nil
		}

		rc.Logger.Error(err, "failed to get cluster credential")
		return err
	}

	if !equality.Semantic.DeepEqual(credential.CredentialInfo, cluster.ClusterCredential.CredentialInfo) {
		metaAccessor := meta.NewAccessor()
		currentResourceVersion, err := metaAccessor.ResourceVersion(credential)
		if err != nil {
			rc.Logger.Error(err, "failed to metaAccessor")
			return err
		}
		metaAccessor.SetResourceVersion(cluster.ClusterCredential, currentResourceVersion)
		err = r.Client.Update(ctx, cluster.ClusterCredential)
		if err != nil {
			rc.Logger.Error(err, "failed to update cluster credential")
			return err
		}
		rc.Logger.V(4).Info("update cluster credential success")
	}

	c := &devopsv1.Cluster{}
	err = r.Client.Get(ctx, rc.Key, c)
	if err != nil {
		if apierrors.IsNotFound(err) {
			rc.Logger.Error(err, "not find cluster")
			return nil
		}

		rc.Logger.Error(err, "failed to get cluster")
		return err
	}

	if !equality.Semantic.DeepEqual(c.Status, cluster.Cluster.Status) {
		metaAccessor := meta.NewAccessor()
		currentResourceVersion, err := metaAccessor.ResourceVersion(c)
		if err != nil {
			rc.Logger.Error(err, "failed to metaAccessor")
			return err
		}

		metaAccessor.SetResourceVersion(cluster.Cluster, currentResourceVersion)
		err = r.Client.Status().Update(ctx, cluster.Cluster)
		if err != nil {
			rc.Logger.Error(err, "failed to update cluster status")
			return err
		}

		rc.Logger.V(4).Info("update cluster status success")
	}

	return nil
}

func (r *clusterReconciler) onCreate(ctx context.Context, rc *clusterContext, p cluster.Provider, clusterWrapper *common.Cluster) error {
	err := p.OnCreate(ctx, clusterWrapper)
	if err != nil {
		clusterWrapper.Cluster.Status.Message = err.Error()
		clusterWrapper.Cluster.Status.Reason = reasonFailedInit
	} else {
		condition := clusterWrapper.Cluster.Status.Conditions[len(clusterWrapper.Cluster.Status.Conditions)-1]
		if condition.Status == devopsv1.ConditionFalse { // means current condition run into error
			clusterWrapper.Cluster.Status.Message = condition.Message
			clusterWrapper.Cluster.Status.Reason = condition.Reason
		} else {
			clusterWrapper.Cluster.Status.Message = ""
			clusterWrapper.Cluster.Status.Reason = ""
		}
	}

	return nil
}

func (r *clusterReconciler) onUpdate(ctx context.Context, rc *clusterContext, p cluster.Provider, clusterWrapper *common.Cluster) error {
	err := p.OnUpdate(ctx, clusterWrapper)
	if err != nil {
		clusterWrapper.Cluster.Status.Message = err.Error()
		clusterWrapper.Cluster.Status.Reason = reasonFailedUpdate
	} else {
		clusterWrapper.Cluster.Status.Message = ""
		clusterWrapper.Cluster.Status.Reason = ""
	}

	return nil
}
