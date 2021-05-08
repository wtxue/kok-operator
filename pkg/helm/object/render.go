package object

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/klog/v2"
)

func GetDefaultValues(fs http.FileSystem) ([]byte, error) {
	file, err := fs.Open(chartutil.ValuesfileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read default values")
	}

	return buf.Bytes(), nil
}

func getFiles(fs http.FileSystem) ([]*loader.BufferedFile, error) {
	files := []*loader.BufferedFile{
		{
			Name: chartutil.ChartfileName,
		},
	}

	// if the Helm chart templates use some resource files (like dashboards), those should be put under resources
	for _, dirName := range []string{"resources", chartutil.TemplatesDir} {
		dir, err := fs.Open(dirName)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			dirFiles, err := dir.Readdir(-1)
			if err != nil {
				return nil, err
			}

			for _, file := range dirFiles {
				filename := file.Name()
				if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") || strings.HasSuffix(filename, "tpl") || strings.HasSuffix(filename, "json") {
					files = append(files, &loader.BufferedFile{
						Name: dirName + "/" + filename,
					})
				}
			}
		}
	}

	for _, f := range files {
		data, err := readIntoBytes(fs, f.Name)
		if err != nil {
			return nil, err
		}

		f.Data = data
	}

	return files, nil
}

func readIntoBytes(fs http.FileSystem, filename string) ([]byte, error) {
	file, err := fs.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file")
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read file")
	}

	return buf.Bytes(), nil
}

func InstallObjectOrder() func(o *K8sObject) int {
	var Order = []string{
		"CustomResourceDefinition",
		"Namespace",
		"ResourceQuota",
		"LimitRange",
		"PodSecurityPolicy",
		"PodDisruptionBudget",
		"Secret",
		"ConfigMap",
		"StorageClass",
		"PersistentVolume",
		"PersistentVolumeClaim",
		"ServiceAccount",
		"ClusterRole",
		"ClusterRoleList",
		"ClusterRoleBinding",
		"ClusterRoleBindingList",
		"Role",
		"RoleList",
		"RoleBinding",
		"RoleBindingList",
		"Service",
		"DaemonSet",
		"Pod",
		"ReplicationController",
		"ReplicaSet",
		"Deployment",
		"HorizontalPodAutoscaler",
		"StatefulSet",
		"Job",
		"CronJob",
		"Ingress",
		"APIService",
	}

	order := make(map[string]int, len(Order))
	for i, kind := range Order {
		order[kind] = i
	}

	return func(o *K8sObject) int {
		if nr, ok := order[o.Kind]; ok {
			return nr
		}
		return 1000
	}
}

func UninstallObjectOrder() func(o *K8sObject) int {
	var Order = []string{
		"APIService",
		"Ingress",
		"Service",
		"CronJob",
		"Job",
		"StatefulSet",
		"HorizontalPodAutoscaler",
		"Deployment",
		"ReplicaSet",
		"ReplicationController",
		"Pod",
		"DaemonSet",
		"RoleBindingList",
		"RoleBinding",
		"RoleList",
		"Role",
		"ClusterRoleBindingList",
		"ClusterRoleBinding",
		"ClusterRoleList",
		"ClusterRole",
		"ServiceAccount",
		"PersistentVolumeClaim",
		"PersistentVolume",
		"StorageClass",
		"ConfigMap",
		"Secret",
		"PodDisruptionBudget",
		"PodSecurityPolicy",
		"LimitRange",
		"ResourceQuota",
		"Policy",
		"Gateway",
		"VirtualService",
		"DestinationRule",
		"Handler",
		"Instance",
		"Rule",
		"Namespace",
		"CustomResourceDefinition",
	}

	order := make(map[string]int, len(Order))
	for i, kind := range Order {
		order[kind] = i
	}

	return func(o *K8sObject) int {
		if nr, ok := order[o.Kind]; ok {
			return nr
		}
		return 1000
	}
}

func GetRequestedChart(chartPackage []byte) (*chart.Chart, error) {
	return loader.LoadArchive(bytes.NewReader(chartPackage))
}

// GetVersionSet retrieves a set of available k8s API versions
func GetVersionSet(client discovery.ServerResourcesInterface) (chartutil.VersionSet, error) {
	groups, resources, err := client.ServerGroupsAndResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return chartutil.DefaultVersionSet, errors.Wrap(err, "could not get apiVersions from Kubernetes")
	}

	// FIXME: The Kubernetes test fixture for cli appears to always return nil
	// for calls to Discovery().ServerGroupsAndResources(). So in this case, we
	// return the default API list. This is also a safe value to return in any
	// other odd-ball case.
	if len(groups) == 0 && len(resources) == 0 {
		return chartutil.DefaultVersionSet, nil
	}

	versionMap := make(map[string]interface{})
	versions := []string{}

	// Extract the groups
	for _, g := range groups {
		for _, gv := range g.Versions {
			versionMap[gv.GroupVersion] = struct{}{}
		}
	}

	// Extract the resources
	var id string
	var ok bool
	for _, r := range resources {
		for _, rl := range r.APIResources {

			// A Kind at a GroupVersion can show up more than once. We only want
			// it displayed once in the final output.
			id = path.Join(r.GroupVersion, rl.Kind)
			if _, ok = versionMap[id]; !ok {
				versionMap[id] = struct{}{}
			}
		}
	}

	// Convert to a form that NewVersionSet can use
	for k := range versionMap {
		versions = append(versions, k)
	}

	return chartutil.VersionSet(versions), nil
}

// capabilities builds a Capabilities from discovery information.
func GetCapabilities(getter genericclioptions.RESTClientGetter) (*chartutil.Capabilities, error) {
	dc, err := getter.ToDiscoveryClient()
	if err != nil {
		return nil, errors.Wrap(err, "could not get Kubernetes discovery client")
	}
	// force a discovery cache invalidation to always fetch the latest server version/capabilities.
	dc.Invalidate()
	kubeVersion, err := dc.ServerVersion()
	if err != nil {
		return nil, errors.Wrap(err, "could not get server version from Kubernetes")
	}
	// Issue #6361:
	// Client-Go emits an error when an API service is registered but unimplemented.
	// We trap that error here and print a warning. But since the discovery client continues
	// building the API object, it is correctly populated with all valid APIs.
	// See https://github.com/kubernetes/kubernetes/issues/72051#issuecomment-521157642
	apiVersions, err := GetVersionSet(dc)
	if err != nil {
		if discovery.IsGroupDiscoveryFailedError(err) {
			klog.Warningf("WARNING: The Kubernetes server has an orphaned API service. Server reports: %s", err)
			klog.Warningf("WARNING: To fix this, kubectl delete apiservice <service-name>")
		} else {
			return nil, errors.Wrap(err, "could not get apiVersions from Kubernetes")
		}
	}

	return &chartutil.Capabilities{
		APIVersions: apiVersions,
		KubeVersion: chartutil.KubeVersion{
			Version: kubeVersion.GitVersion,
			Major:   kubeVersion.Major,
			Minor:   kubeVersion.Minor,
		},
	}, nil
}

func RenderTemplate(chartPackage []byte, rlsName, ns string, overrideValue string) (K8sObjects, error) {
	chrt, err := GetRequestedChart(chartPackage)
	if err != nil {
		return nil, fmt.Errorf("loading chart has an error: %+v", err)
	}

	overrideValues, err := chartutil.ReadValues([]byte(overrideValue))
	if err != nil {
		return nil, fmt.Errorf("ReadValues has an error: %+v", err)
	}

	options := chartutil.ReleaseOptions{
		Name:      rlsName,
		Namespace: ns,
		Revision:  1,
		IsInstall: true,
		IsUpgrade: false,
	}

	chrtValues, err := chartutil.ToRenderValues(chrt, overrideValues, options, nil)
	if err != nil {
		return nil, err
	}

	renderedTemplates, err := engine.Render(chrt, chrtValues)
	if err != nil {
		klog.Errorf("render err: %+v", err)
		return nil, err
	}

	var objects []*K8sObject
	for name, yaml := range renderedTemplates {
		yaml = RemoveNonYAMLLines(yaml)
		if yaml == "" {
			continue
		}

		objs, err := ParseK8sObjectsFromYAMLManifest(yaml)
		if err != nil {
			return nil, errors.Wrapf(err, "name: %s Failed to parse yaml to a k8s objs", name)
		}

		objects = append(objects, objs...)
	}

	return objects, nil
}

func Render(fs http.FileSystem, rlsName, ns, chartName string, overrideValue string) (K8sObjects, error) {
	files, err := getFiles(fs)
	if err != nil {
		return nil, err
	}

	// Create chart and render templates
	chrt, err := loader.LoadFiles(files)
	if err != nil {
		return nil, err
	}

	overrideValues, err := chartutil.ReadValues([]byte(overrideValue))
	if err != nil {
		return nil, fmt.Errorf("ReadValues has an error: %+v", err)
	}

	options := chartutil.ReleaseOptions{
		Name:      rlsName,
		Namespace: ns,
		Revision:  1,
		IsInstall: true,
		IsUpgrade: false,
	}

	chrtValues, err := chartutil.ToRenderValues(chrt, overrideValues, options, nil)
	if err != nil {
		return nil, err
	}

	renderedTemplates, err := engine.Render(chrt, chrtValues)
	if err != nil {
		klog.Errorf("render err: %+v", err)
		return nil, err
	}

	// Merge templates and inject
	var buf bytes.Buffer
	for _, tmpl := range files {
		if !strings.HasSuffix(tmpl.Name, "yaml") && !strings.HasSuffix(tmpl.Name, "yml") && !strings.HasSuffix(tmpl.Name, "tpl") {
			continue
		}
		t := path.Join(chartName, tmpl.Name)
		if _, err := buf.WriteString(renderedTemplates[t]); err != nil {
			return nil, err
		}
		buf.WriteString("\n---\n")
	}

	objects, err := ParseK8sObjectsFromYAMLManifest(buf.String())
	if err != nil {
		return nil, err
	}

	return objects, nil
}
