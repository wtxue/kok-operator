package common

import (
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func ObjectMeta(name string, labels map[string]string, config runtime.Object) metav1.ObjectMeta {
	obj := config.DeepCopyObject()
	objMeta, _ := meta.Accessor(obj)
	ovk := config.GetObjectKind().GroupVersionKind()

	return metav1.ObjectMeta{
		Name:      name,
		Namespace: objMeta.GetNamespace(),
		Labels:    labels,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion:         ovk.GroupVersion().String(),
				Kind:               ovk.Kind,
				Name:               objMeta.GetName(),
				UID:                objMeta.GetUID(),
				Controller:         k8sutil.BoolPointer(true),
				BlockOwnerDeletion: k8sutil.BoolPointer(true),
			},
		},
	}
}

func ObjectMetaWithAnnotations(name string, labels map[string]string, annotations map[string]string, config runtime.Object) metav1.ObjectMeta {
	o := ObjectMeta(name, labels, config)
	o.Annotations = annotations
	return o
}

func ObjectMetaClusterScope(name string, labels map[string]string, config runtime.Object) metav1.ObjectMeta {
	obj := config.DeepCopyObject()
	objMeta, _ := meta.Accessor(obj)
	ovk := config.GetObjectKind().GroupVersionKind()

	return metav1.ObjectMeta{
		Name:   name,
		Labels: labels,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion:         ovk.GroupVersion().String(),
				Kind:               ovk.Kind,
				Name:               objMeta.GetName(),
				UID:                objMeta.GetUID(),
				Controller:         k8sutil.BoolPointer(true),
				BlockOwnerDeletion: k8sutil.BoolPointer(true),
			},
		},
	}
}
