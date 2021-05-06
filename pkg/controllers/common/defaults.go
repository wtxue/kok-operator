package common

import (
	"github.com/wtxue/kok-operator/pkg/k8sutil"
	appsv1 "k8s.io/api/apps/v1"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultDeployAnnotations() map[string]string {
	return map[string]string{}
}

func DefaultRollingUpdateStrategy() appsv1.DeploymentStrategy {
	return appsv1.DeploymentStrategy{
		Type: appsv1.RollingUpdateDeploymentStrategyType,
		RollingUpdate: &appsv1.RollingUpdateDeployment{
			MaxSurge:       k8sutil.IntstrPointer(1),
			MaxUnavailable: k8sutil.IntstrPointer(1),
		},
	}
}

func TargetAvgCpuUtil80() []autoscalev2beta1.MetricSpec {
	return []autoscalev2beta1.MetricSpec{
		{
			Type: autoscalev2beta1.ResourceMetricSourceType,
			Resource: &autoscalev2beta1.ResourceMetricSource{
				Name:                     corev1.ResourceCPU,
				TargetAverageUtilization: k8sutil.IntPointer(80),
			},
		},
	}
}

func ComponentAffinity(ns string, labels map[string]string) *corev1.Affinity {
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 1,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
						Namespaces:  []string{ns},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}
}

func ComponentTolerations() []corev1.Toleration {
	return []corev1.Toleration{
		{
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoSchedule,
		},
		{
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
	}
}

func ComponentEnv(ctx *ClusterContext) []corev1.EnvVar {
	envs := []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: "INSTANCE_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
	}

	return envs
}
