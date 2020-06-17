package validation

import (
	"context"
	"fmt"
	"reflect"

	"github.com/thoas/go-funk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ValidateEnum validates a given enum.
// nil or nil pointer is valid.
// zero value is invalid.
func ValidateEnum(value interface{}, fldPath *field.Path, values interface{}) field.ErrorList {
	allErrs := field.ErrorList{}

	if value == nil {
		return allErrs
	}

	validValuesString := funk.Map(values, func(i interface{}) string {
		return fmt.Sprintf("%v", i)
	}).([]string)

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if reflect.ValueOf(value).IsNil() {
			return allErrs
		}
	} else {
		if v.IsZero() {
			allErrs = append(allErrs, field.Required(fldPath, fmt.Sprintf("valid values: %v", validValuesString)))
			return allErrs
		}
	}

	if !funk.Contains(values, value) {
		allErrs = append(allErrs, field.NotSupported(fldPath, value, validValuesString))
	}

	return allErrs
}

// ValidateRESTConfig validates a given rest.Config.
func ValidateRESTConfig(ctx context.Context, config *rest.Config) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	_, err = clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	return nil
}
