package k8sutil

import (
	"context"
	"encoding/json"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/wtxue/kok-operator/pkg/k8sclient"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type DesiredState string

const (
	DesiredStatePresent DesiredState = "present"
	DesiredStateAbsent  DesiredState = "absent"
)

func Reconcile(log logr.Logger, cli client.Client, desired client.Object, desiredState DesiredState) error {
	if desiredState == "" {
		desiredState = DesiredStatePresent
	}

	current := desired.DeepCopyObject().(client.Object)
	desiredCopy := desired.DeepCopyObject().(client.Object)
	key := client.ObjectKeyFromObject(current)

	gvk, _ := apiutil.GVKForObject(desired, k8sclient.GetScheme())
	desiredType := gvk.Kind
	logger := log.WithValues("kind", desiredType, "key", key.String())

	err := cli.Get(context.TODO(), key, current)
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "kind", desiredType, "key", key.String())
	}
	if apierrors.IsNotFound(err) {
		if desiredState == DesiredStatePresent {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				logger.Error(err, "Failed to set last applied annotation", "desired", desired)
			}
			if err := cli.Create(context.TODO(), desired); err != nil {
				return emperror.WrapWith(err, "creating resource failed", "kind", desiredType, "key", key.String())
			}
			logger.Info("resource created")
		}
	} else {
		if desiredState == DesiredStatePresent {
			patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired, patch.IgnoreStatusFields())
			if err != nil {
				logger.Error(err, "could not match objects", "kind", desiredType, "key", key.String())
			} else if patchResult.IsEmpty() {
				logger.Info("resource is in sync")
				return nil
			} else {
				logger.Info("resource diffs",
					"patch", string(patchResult.Patch),
					"current", string(patchResult.Current),
					"modified", string(patchResult.Modified),
					"original", string(patchResult.Original))
			}

			// Need to set this before resourceversion is set, as it would constantly change otherwise
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				logger.Error(err, "Failed to set last applied annotation", "key", key.String())
			}

			metaAccessor := meta.NewAccessor()
			currentResourceVersion, err := metaAccessor.ResourceVersion(current)
			if err != nil {
				return err
			}

			metaAccessor.SetResourceVersion(desired, currentResourceVersion)
			prepareResourceForUpdate(current, desired)

			if err := cli.Update(context.TODO(), desired); err != nil {
				if apierrors.IsConflict(err) || apierrors.IsInvalid(err) {
					logger.Error(err, "resource needs to be re-created")
					err := cli.Delete(context.TODO(), current)
					if err != nil {
						return emperror.WrapWith(err, "could not delete resource", "kind", desiredType, "key", key.String())
					}
					logger.Info("resource deleted")
					if err := cli.Create(context.TODO(), desiredCopy); err != nil {
						return emperror.WrapWith(err, "creating resource failed", "kind", desiredType, "key", key.String())
					}
					logger.Info("resource created")
					return nil
				}

				return emperror.WrapWith(err, "updating resource failed", "kind", desiredType, "key", key.String())
			}
			logger.Info("resource updated")
		} else if desiredState == DesiredStateAbsent {
			if err := cli.Delete(context.TODO(), current); err != nil {
				return emperror.WrapWith(err, "deleting resource failed", "kind", desiredType, "key", key.String())
			}
			logger.Info("resource deleted")
		}
	}
	return nil
}

func prepareResourceForUpdate(current, desired runtime.Object) {
	switch desired.(type) {
	case *corev1.Service:
		svc := desired.(*corev1.Service)
		svc.Spec.ClusterIP = current.(*corev1.Service).Spec.ClusterIP
	}
}

// IsObjectChanged checks whether there is an actual difference between the two objects
func IsObjectChanged(oldObj, newObj runtime.Object, ignoreStatusChange bool) (bool, error) {
	oldCopy := oldObj.DeepCopyObject()
	newCopy := newObj.DeepCopyObject()

	metaAccessor := meta.NewAccessor()
	currentResourceVersion, err := metaAccessor.ResourceVersion(oldCopy)
	if err == nil {
		metaAccessor.SetResourceVersion(newCopy, currentResourceVersion)
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(oldCopy, newCopy, patch.IgnoreStatusFields())
	if err != nil {
		return true, emperror.WrapWith(err, "could not match objects", "kind", oldCopy.GetObjectKind())
	} else if patchResult.IsEmpty() {
		return false, nil
	}

	if ignoreStatusChange {
		var patchMap map[string]interface{}
		json.Unmarshal(patchResult.Patch, &patchMap)
		delete(patchMap, "status")
		if len(patchMap) == 0 {
			return false, nil
		}
	}

	return true, nil
}

// ReconcileNamespaceLabelsIgnoreNotFound patches namespaces by adding/removing labels, returns without error if namespace is not found
func ReconcileNamespaceLabelsIgnoreNotFound(log logr.Logger, cli client.Client, namespace string, labels map[string]string, labelsToRemove []string) error {
	var ns = &corev1.Namespace{}
	err := cli.Get(context.TODO(), client.ObjectKey{Name: namespace}, ns)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("namespace not found, ignoring", "namespace", namespace)
			return nil
		}

		return emperror.WrapWith(err, "getting namespace failed", "namespace", namespace)
	}

	updateNeeded := false
	for dlk, dlv := range labels {
		if ns.Labels == nil {
			ns.Labels = make(map[string]string)
		}
		if clv, ok := ns.Labels[dlk]; !ok || clv != dlv {
			ns.Labels[dlk] = dlv
			updateNeeded = true
		}
	}
	for _, labelKey := range labelsToRemove {
		if _, ok := ns.Labels[labelKey]; ok {
			delete(ns.Labels, labelKey)
			updateNeeded = true
		}
	}
	if updateNeeded {
		if err := cli.Update(context.TODO(), ns); err != nil {
			return emperror.WrapWith(err, "updating namespace failed", "namespace", namespace)
		}
		log.Info("namespace labels reconciled", "namespace", namespace, "labels", labels)
	}

	return nil
}
