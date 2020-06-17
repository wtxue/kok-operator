package k8sutil

import (
	"context"

	"github.com/goph/emperror"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

func ReconcileCrds(cfg *rest.Config, crds []*extensionsobj.CustomResourceDefinition) error {
	ctx := context.TODO()
	cli := extensionsclient.NewForConfigOrDie(cfg)
	for _, crd := range crds {
		obj, err := cli.ApiextensionsV1beta1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				_, err := cli.ApiextensionsV1beta1().CustomResourceDefinitions().Create(ctx, crd, metav1.CreateOptions{})
				if err != nil {
					return emperror.WrapWith(err, "Create CRD failed", "kind", crd.Spec.Names.Kind)
				}
			}

			return emperror.WrapWith(err, "getting CRD failed", "kind", crd.Spec.Names.Kind)
		}

		if !equality.Semantic.DeepEqual(obj.Spec, crd.Spec) {
			metaAccessor := meta.NewAccessor()
			currentResourceVersion, err := metaAccessor.ResourceVersion(obj)
			if err != nil {
				return emperror.WrapWith(err, "failed to metaAccessor")
			}
			metaAccessor.SetResourceVersion(crd, currentResourceVersion)
			_, err = cli.ApiextensionsV1beta1().CustomResourceDefinitions().Update(ctx, crd, metav1.UpdateOptions{})
			if err != nil {
				return emperror.WrapWith(err, "Create CRD failed", "kind", crd.Spec.Names.Kind)
			}
			klog.V(4).Infof("update crd: %s success", crd.Name)
		}
	}
	return nil
}
