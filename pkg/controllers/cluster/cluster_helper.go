package cluster

import (
	"context"
	"time"

	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kube-on-kube-operator/pkg/provider"
	clusterprovider "github.com/wtxue/kube-on-kube-operator/pkg/provider/cluster"
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

func (r *clusterReconciler) applyStatus(ctx context.Context, rc *clusterContext, cluster *provider.Cluster) error {
	c := &devopsv1.Cluster{}
	err := r.Client.Get(ctx, rc.Key, c)
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

	credential := &devopsv1.ClusterCredential{}
	err = r.Client.Get(ctx, rc.Key, credential)
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
		currentResourceVersion, err := metaAccessor.ResourceVersion(c)
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

	return nil
}

func (r *clusterReconciler) onCreate(ctx context.Context, rc *clusterContext) error {
	p, err := clusterprovider.GetProvider(rc.Cluster.Spec.Type)
	if err != nil {
		return err
	}

	clusterWrapper, err := provider.GetCluster(ctx, r.Client, rc.Cluster)
	if err != nil {
		return err
	}
	err = p.OnCreate(ctx, clusterWrapper)
	if err != nil {
		clusterWrapper.Status.Message = err.Error()
		clusterWrapper.Status.Reason = reasonFailedInit
	} else {
		condition := clusterWrapper.Status.Conditions[len(clusterWrapper.Status.Conditions)-1]
		if condition.Status == devopsv1.ConditionFalse { // means current condition run into error
			clusterWrapper.Status.Message = condition.Message
			clusterWrapper.Status.Reason = condition.Reason
		} else {
			clusterWrapper.Status.Message = ""
			clusterWrapper.Status.Reason = ""
		}
	}

	r.applyStatus(ctx, rc, clusterWrapper)
	return nil
}

func (r *clusterReconciler) onUpdate(ctx context.Context, rc *clusterContext) error {
	p, err := clusterprovider.GetProvider(rc.Cluster.Spec.Type)
	if err != nil {
		return err
	}

	clusterWrapper, err := provider.GetCluster(ctx, r.Client, rc.Cluster)
	if err != nil {
		return err
	}

	err = p.OnUpdate(ctx, clusterWrapper)
	if err != nil {
		clusterWrapper.Status.Message = err.Error()
		clusterWrapper.Status.Reason = reasonFailedUpdate
	} else {
		clusterWrapper.Status.Message = ""
		clusterWrapper.Status.Reason = ""
	}

	r.applyStatus(ctx, rc, clusterWrapper)
	return nil
}
