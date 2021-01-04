package k8sutil

import (
	"context"
	"reflect"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = logf.Log.WithName("crd")

// nolint
// ReconcileCRDs crds apply must before manager cache start, so crds apply need use raw kubernetes client
func ReconcileCRDs(cfg *rest.Config, crds []*apiextensionsv1.CustomResourceDefinition) error {
	ctx := context.TODO()
	cli := apiextensionsclient.NewForConfigOrDie(cfg)

	for _, crd := range crds {
		existing, err := cli.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				_, err = cli.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, metav1.CreateOptions{})
				if err != nil {
					return err
				}
				logger.Info("create successfully", "name", crd.Name, "kind", crd.Spec.Names.Kind)
			} else {
				return err
			}
		} else {
			// skip Conversion
			existing.Spec.Conversion = nil
			if same := reflect.DeepEqual(crd.Spec, existing.Spec); !same {
				crd.SetResourceVersion(existing.GetResourceVersion())
				_, err = cli.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, crd, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
				logger.Info("update successfully", "name", crd.Name, "kind", crd.Spec.Names.Kind)
			} else {
				logger.Info("update unchanges, ignored", "name", crd.Name, "kind", crd.Spec.Names.Kind)
			}
		}
	}

	backoff := wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    5,
	}

	for _, crd := range crds {
		err := wait.ExponentialBackoff(backoff, func() (bool, error) {
			existing, err := cli.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			for i := range existing.Status.Conditions {
				cond := &existing.Status.Conditions[i]
				if cond.Type == apiextensionsv1.Established &&
					cond.Status == apiextensionsv1.ConditionTrue {
					return true, nil
				}
			}

			logger.Info("wait established ... ", "name", crd.Name, "kind", crd.Spec.Names.Kind)
			return false, nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}
