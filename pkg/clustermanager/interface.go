package clustermanager

import (
	"context"
	"errors"
	"sync"

	helmv3 "github.com/wtxue/kok-operator/pkg/helm/v3"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Interface is used to shield the complexity of the underlying multi-cluster or single cluster.
type Interface interface {
	GetRawCli(clusterNames ...string) (kubernetes.Interface, error)
	GetPod(opts types.NamespacedName, clusterNames ...string) (*corev1.Pod, error)
	GetPods(opts *client.ListOptions, clusterNames ...string) ([]*corev1.Pod, error)
	GetNodes(opts *client.ListOptions, clusterNames ...string) ([]*corev1.Node, error)
	GetDeployment(opts *client.ListOptions, clusterNames ...string) ([]*appv1.Deployment, error)
	GetStatefulsets(opts *client.ListOptions, clusternames ...string) ([]*appv1.StatefulSet, error)
	GetService(opts *client.ListOptions, clusterNames ...string) ([]*corev1.Service, error)
	GetEndpoints(opts *client.ListOptions, clusterNames ...string) ([]*corev1.Endpoints, error)
	GetEvent(opts *client.ListOptions, clusterNames ...string) ([]*corev1.Event, error)
	DeletePods(opts *client.ListOptions, clusterNames ...string) error
	DeletePod(opts types.NamespacedName, clusterNames ...string) error
}

// CustomInterface as the extension of the basic cluster,
// it includes the extra methods for handling your own business.
type CustomInterface interface {
	Interface
	GetHelmRelease(opts map[string]string, clusterNames ...string) ([]*helmv3.Release, error)
}

// GetRawCli return the client of the master cluster as default if len(clusterNames) == 0,
// otherwise it will return a client with the special name.
func (m *ClusterManager) GetRawCli(clusterNames ...string) (kubernetes.Interface, error) {
	if len(clusterNames) > 0 {
		if len(clusterNames) > 1 {
			return nil, errors.New("too many clusterNames")
		}
		cluster, err := m.Get(clusterNames[0])
		if err != nil {
			return nil, err
		}
		return cluster.KubeCli, nil
	}
	return m.KubeCli, nil
}

// GetPod ...
func (m *ClusterManager) GetPod(opts types.NamespacedName, clusterNames ...string) (*corev1.Pod, error) {
	clusters := m.GetAll(clusterNames...)
	ctx := context.Background()
	pod := &corev1.Pod{}

	for _, cluster := range clusters {
		err := cluster.Client.Get(ctx, opts, pod)
		if err != nil {
			return nil, err
		}
	}
	return pod, nil
}

// GetPods ...
func (m *ClusterManager) GetPods(opts *client.ListOptions, clusterNames ...string) ([]*corev1.Pod, error) {
	clusters := m.GetAll(clusterNames...)
	ctx := context.Background()
	result := make([]*corev1.Pod, 0)

	for _, cluster := range clusters {
		podList := &corev1.PodList{}
		err := cluster.Client.List(ctx, podList, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		for i := range podList.Items {
			pod := &podList.Items[i]
			result = append(result, pod)
		}
	}
	return result, nil
}

// GetNodes ...
func (m *ClusterManager) GetNodes(opts *client.ListOptions, clusterNames ...string) ([]*corev1.Node, error) {
	clusters := m.GetAll(clusterNames...)
	ctx := context.Background()
	result := make([]*corev1.Node, 0)

	for _, cluster := range clusters {
		nodeList := &corev1.NodeList{}
		err := cluster.Client.List(ctx, nodeList, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		for i := range nodeList.Items {
			node := &nodeList.Items[i]
			result = append(result, node)
		}
	}
	return result, nil
}

// GetDeployment ...
func (m *ClusterManager) GetDeployment(opts *client.ListOptions, clusterNames ...string) ([]*appv1.Deployment, error) {
	clusters := m.GetAll(clusterNames...)
	ctx := context.Background()
	result := make([]*appv1.Deployment, 0)

	for _, cluster := range clusters {
		deployList := &appv1.DeploymentList{}
		err := cluster.Client.List(ctx, deployList, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		for i := range deployList.Items {
			deploy := &deployList.Items[i]
			result = append(result, deploy)
		}
	}
	return result, nil
}

// GetStatefulsets ...
func (m *ClusterManager) GetStatefulsets(opts *client.ListOptions, clusterNames ...string) ([]*appv1.StatefulSet, error) {
	clusters := m.GetAll(clusterNames...)
	ctx := context.Background()
	result := make([]*appv1.StatefulSet, 0)

	for _, cluster := range clusters {
		staList := &appv1.StatefulSetList{}
		err := cluster.Client.List(ctx, staList, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		for i := range staList.Items {
			sta := &staList.Items[i]
			result = append(result, sta)
		}
	}
	return result, nil
}

// GetService ...
func (m *ClusterManager) GetService(opts *client.ListOptions, clusterNames ...string) ([]*corev1.Service, error) {
	clusters := m.GetAll(clusterNames...)
	ctx := context.Background()
	result := make([]*corev1.Service, 0)

	for _, cluster := range clusters {
		serviceList := &corev1.ServiceList{}
		err := cluster.Client.List(ctx, serviceList, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		for i := range serviceList.Items {
			service := &serviceList.Items[i]
			result = append(result, service)
		}
	}
	return result, nil
}

// GetEndpoints ...
func (m *ClusterManager) GetEndpoints(opts *client.ListOptions, clusterNames ...string) ([]*corev1.Endpoints, error) {
	clusters := m.GetAll(clusterNames...)
	ctx := context.Background()
	result := make([]*corev1.Endpoints, 0)

	for _, cluster := range clusters {
		endpointsList := &corev1.EndpointsList{}
		err := cluster.Client.List(ctx, endpointsList, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		for i := range endpointsList.Items {
			endpoints := &endpointsList.Items[i]
			result = append(result, endpoints)
		}
	}
	return result, nil
}

// GetEvent ...
func (m *ClusterManager) GetEvent(opts *client.ListOptions, clusterNames ...string) ([]*corev1.Event, error) {
	clusters := m.GetAll(clusterNames...)
	ctx := context.Background()
	result := make([]*corev1.Event, 0)

	for _, cluster := range clusters {
		eventList := &corev1.EventList{}
		err := cluster.Client.List(ctx, eventList, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		for i := range eventList.Items {
			event := &eventList.Items[i]
			result = append(result, event)
		}
	}
	return result, nil
}

// DeletePods ...
func (m *ClusterManager) DeletePods(opts *client.ListOptions, clusterNames ...string) error {
	clusters := m.GetAll(clusterNames...)
	ctx := context.Background()

	for _, cluster := range clusters {
		podList := &corev1.PodList{}
		err := cluster.Client.List(ctx, podList, opts)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return err
		}
		for i := range podList.Items {
			err = cluster.Client.Delete(ctx, &podList.Items[i])
			if err != nil {
				klog.Errorf("delete pod error: %v", err)
			}
		}
	}
	return nil
}

// DeletePod ...
func (m *ClusterManager) DeletePod(opts types.NamespacedName, clusterNames ...string) error {
	clusters := m.GetAll(clusterNames...)
	ctx := context.Background()
	pod := &corev1.Pod{}

	for _, cluster := range clusters {
		err := cluster.Client.Get(ctx, opts, pod)
		if err != nil {
			return err
		}
		err = cluster.Client.Delete(ctx, pod)
		if err != nil {
			klog.Errorf("delete pod error: %v", err)
		}
	}
	return nil
}

// GetHelmRelease ...
func (m *ClusterManager) GetHelmRelease(opts map[string]string, clusterNames ...string) ([]*helmv3.Release, error) {
	clusters := m.GetAll(clusterNames...)
	result := make([]*helmv3.Release, 0)
	wg := sync.WaitGroup{}
	for _, cluster := range clusters {
		wg.Add(1)
		go func(wg *sync.WaitGroup, cluster *Cluster, result []*helmv3.Release) {

			wg.Done()
		}(&wg, cluster, result)
	}
	wg.Wait()
	return result, nil
}
