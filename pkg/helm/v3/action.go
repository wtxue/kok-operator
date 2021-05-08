package v3

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/wtxue/kok-operator/pkg/helm/object"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	logger = logf.Log.WithName("helmv3")
)

func adaptReleasePtr(rawRelease *release.Release) *Release {
	objs, err := object.ParseK8sObjectsFromYAMLManifest(rawRelease.Manifest)
	if err != nil {
		logger.Error(err, "failed to get resources", "release", rawRelease.Name)
	}

	return &Release{
		ReleaseName:    rawRelease.Name,
		ChartName:      rawRelease.Chart.Metadata.Name,
		Namespace:      rawRelease.Namespace,
		Values:         rawRelease.Chart.Values,
		Version:        rawRelease.Chart.Metadata.Version,
		ReleaseVersion: int32(rawRelease.Version),
		ReleaseInfo: &ReleaseInfo{
			FirstDeployed: rawRelease.Info.FirstDeployed.Time,
			LastDeployed:  rawRelease.Info.LastDeployed.Time,
			Deleted:       rawRelease.Info.Deleted.Time,
			Description:   rawRelease.Info.Description,
			Status:        rawRelease.Info.Status.String(),
			Notes:         base64.StdEncoding.EncodeToString([]byte(rawRelease.Info.Notes)),
			Values:        rawRelease.Config,
		},
		ReleaseResources: objs,
	}
}

func getActionConfiguration(clientGetter genericclioptions.RESTClientGetter, namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(clientGetter, namespace, "", klog.Infof); err != nil {
		logger.Error(err, "failed to initialize action config")
		return nil, errors.Wrap(err, "failed to initialize  action config")
	}

	return actionConfig, nil
}

// isChartInstallable validates if a chart can be installed
// Application chart type is only installable
func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func processOptions(listAction *action.List, options *Options) *action.List {
	a := listAction
	if options.Filter != nil {
		a.Filter = *options.Filter
	}

	if options.Namespace == "" {
		a.AllNamespaces = true
	}
	// apply other options here
	return a
}

// UninstallReleases
func UninstallReleases(helmEnv *HelmEnv, releaseName string, opt *Options) error {
	ns := "default"
	if opt.Namespace != "" {
		ns = opt.Namespace
	}

	actionConfig, err := getActionConfiguration(helmEnv.RESTClientGetter, ns)
	if err != nil {
		return errors.Wrap(err, "failed to get action configuration")
	}

	uninstallAction := action.NewUninstall(actionConfig)
	uninstallAction.Timeout = time.Minute * 5

	res, err := uninstallAction.Run(releaseName)
	if err != nil {
		return err
	}
	if res != nil && res.Info != "" {
		logger.Info(res.Info)
	}

	logger.Info("release successfully uninstalled", "releaseName", releaseName)

	return nil
}

func ListReleases(_ context.Context, helmEnv *HelmEnv, opt *Options) ([]*Release, error) {
	actionConfig, err := getActionConfiguration(helmEnv.RESTClientGetter, opt.Namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get action configuration")
	}

	listAction := action.NewList(actionConfig)
	listAction.SetStateMask()

	// applies options if any
	listAction = processOptions(listAction, opt)

	results, err := listAction.Run()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list releases")
	}

	releases := make([]*Release, 0, len(results))
	for _, result := range results {
		releases = append(releases, adaptReleasePtr(result))
	}

	return releases, nil
}

func InstallRelease(ctx context.Context, helmEnv *HelmEnv, releaseInput *Release, options *Options) (*Release, error) {
	ns := "default"
	if options.Namespace != "" {
		ns = options.Namespace
	}

	actionConfig, err := getActionConfiguration(helmEnv.RESTClientGetter, ns)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get  action configuration")
	}

	installAction := action.NewInstall(actionConfig)
	installAction.Namespace = ns
	// TODO the generate name is already coded into the options; revisit this after h2 is removed
	if releaseInput.ReleaseName == "" {
		installAction.GenerateName = true
	}

	name, chartRef, err := installAction.NameAndChart(releaseInput.NameAndChartSlice())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get  name  and chart")
	}
	installAction.ReleaseName = name
	installAction.Wait = options.Wait
	installAction.Timeout = time.Minute * 5
	installAction.Version = releaseInput.Version
	installAction.SkipCRDs = options.SkipCRDs

	cp, err := installAction.ChartPathOptions.LocateChart(chartRef, helmEnv.Cli)
	if err != nil {
		return nil, errors.Wrap(err, "failed to locate chart")
	}

	p := getter.All(helmEnv.Cli)

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load chart")
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		return nil, errors.Wrap(err, "chart is not installable")
	}

	if chartRequested.Metadata.Deprecated {
		logger.Info(" This chart is deprecated")
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if installAction.DependencyUpdate {
				man := &downloader.Manager{
					Out:              os.Stdout,
					ChartPath:        cp,
					Keyring:          installAction.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: helmEnv.Cli.RepositoryConfig,
					RepositoryCache:  helmEnv.Cli.RepositoryCache,
				}
				if err := man.Update(); err != nil {
					return nil, errors.Wrap(err, "failed to update chart dependencies")
				}
			} else {
				return nil, errors.Wrap(err, "failed to check chart dependencies")
			}
		}
	}

	namespaces, err := helmEnv.KubeCli.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list kubernetes namespaces")
	}

	foundNs := false
	for _, ns := range namespaces.Items {
		if ns.Name == installAction.Namespace {
			foundNs = true
		}
	}

	if !foundNs {
		if _, err := helmEnv.KubeCli.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: installAction.Namespace,
			},
		}, metav1.CreateOptions{}); err != nil {
			return nil, errors.Wrap(err, "failed to create release namespace")
		}
	}

	releasePtr, err := installAction.Run(chartRequested, releaseInput.Values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to install chart")
	}

	return adaptReleasePtr(releasePtr), nil
}

func UpgradeRelease(ctx context.Context, helmEnv *HelmEnv, releaseInput *Release, options *Options) (*Release, error) {
	// this is the value coming from env settings in the CLI
	ns := "default"
	if releaseInput.Namespace != "" {
		ns = releaseInput.Namespace
	}

	if options.Namespace != "" {
		ns = options.Namespace
	}

	actionConfig, err := getActionConfiguration(helmEnv.RESTClientGetter, ns)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get  action configuration")
	}

	upgradeAction := action.NewUpgrade(actionConfig)
	upgradeAction.Namespace = ns
	upgradeAction.Install = options.Install
	upgradeAction.Wait = options.Wait
	upgradeAction.Timeout = time.Minute * 5
	upgradeAction.Version = releaseInput.Version
	upgradeAction.SkipCRDs = options.SkipCRDs

	if upgradeAction.Version == "" && upgradeAction.Devel {
		logger.Info("setting version to >0.0.0-0")
		upgradeAction.Version = ">0.0.0-0"
	}

	chartPath, err := upgradeAction.ChartPathOptions.LocateChart(releaseInput.ChartName, helmEnv.Cli)
	if err != nil {
		return nil, errors.Wrap(err, "failed to locate chart")
	}

	// Check chart dependencies to make sure all are present in /charts
	ch, err := loader.Load(chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load chart")
	}
	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return nil, errors.Wrap(err, "failed to check dependencies")
		}
	}

	if ch.Metadata.Deprecated {
		logger.Info("This chart is deprecated", "chart", ch.Name())
	}

	rel, err := upgradeAction.Run(releaseInput.ReleaseName, ch, releaseInput.Values)
	if err != nil {
		return nil, errors.Wrap(err, "UPGRADE FAILED")
	}

	logger.Info("release has been upgraded. Happy Helming!", "releaseName", releaseInput.ReleaseName)

	return adaptReleasePtr(rel), nil
}

func NewHelmEnv(helmEnv *HelmEnv, kubeConfig []byte, ns string, kubecli kubernetes.Interface) (*HelmEnv, error) {
	restClientGetter, err := NewCustomGetter(ns, kubeConfig, helmEnv.KubeCache, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to new restClientGetter")
	}

	helmEnv.RESTClientGetter = restClientGetter
	return &HelmEnv{
		KubeCache:        helmEnv.KubeCache,
		Cli:              helmEnv.Cli,
		KubeCli:          kubecli,
		RESTClientGetter: restClientGetter,
	}, nil
}

// Create or update a release, creating it if release parameter is nil, otherwise, updating it.
func ApplyRelease(helmEnv *HelmEnv, rlsName, chartUrlName, specChartVersion string, chartPackage []byte,
	namespace string, va []byte, runningRls *Release) (*Release, error) {
	var (
		appliedRls *Release
		rlsErr     error
	)

	overrideValues, err := chartutil.ReadValues(va)
	if err != nil {
		return nil, fmt.Errorf("ReadValues has an error: %v", err)
	}

	installRls := &Release{
		ReleaseName: rlsName,
		ChartName:   chartUrlName,
		Namespace:   namespace,
		Values:      overrideValues,
		Version:     specChartVersion,
	}

	opt := &Options{
		Namespace:   namespace,
		DryRun:      false,
		Wait:        false,
		Timeout:     0,
		ReuseValues: false,
		Install:     true,
		Filter:      &rlsName,
		SkipCRDs:    false,
	}

	if specChartVersion == "" && len(chartPackage) > 0 {
		c, err := loader.LoadArchive(bytes.NewReader(chartPackage))
		if err == nil {
			specChartVersion = c.Metadata.Version
		}
	}

	if runningRls == nil {
		listRep, err := ListReleases(context.TODO(), helmEnv, opt)
		if err == nil && listRep != nil && len(listRep) > 0 {
			if listRep[0].ReleaseInfo.Status != string(release.StatusDeployed) {
				err := UninstallReleases(helmEnv, rlsName, opt)
				if err != nil {
					klog.Errorf("delete not deployed rlsName: %s err:%v", rlsName, err)
					return nil, err
				} else {
					klog.Infof("====> delete not deployed rrlsName: %s successfully", rlsName)
					listRep = nil
				}
			} else {
				runningRls = listRep[0]
				klog.Infof("====> find runningRls name: %s  version: %s releaseVersion: %d",
					rlsName, runningRls.Version, runningRls.ReleaseVersion)
			}
		}
	}

	// If the release need to apply is nil, we create this release directly.
	if runningRls == nil {
		rep, err := InstallRelease(context.TODO(), helmEnv, installRls, opt)
		if err == nil && rep != nil {
			appliedRls = rep
			klog.V(4).Infof("Release[%s] has been installed successfully, current version: %d", rlsName, appliedRls.ReleaseVersion)
		}
		rlsErr = err
	} else {
		// If the release need to apply has been passed here, it is necessary to compare it with the running release.
		var isDifferent int

		if specChartVersion != "" {
			runningVersion := runningRls.Version
			if strings.Compare(specChartVersion, runningVersion) != 0 {
				klog.V(3).Infof("Release[%s] chart version will changed, runningVersion %s => specChartVersion %s", rlsName, runningVersion, specChartVersion)
				isDifferent++
			}
		}

		if isDifferent <= 0 {
			runningRaw := runningRls.ReleaseInfo.Values
			if runningRaw == nil {
				runningRaw = map[string]interface{}{}
			}
			isEquivalent := reflect.DeepEqual(installRls.Values, runningRaw)
			if isEquivalent {
				klog.V(4).Infof("Release[%s]'s running values not changed, ignore", rlsName)
				return runningRls, nil
			} else {
				isDifferent++
			}
		}

		// if the running release differ with the spec one, we update it with the spec one.
		if isDifferent > 0 {
			klog.Infof("start upgrade rlsName: %s ...", installRls.ReleaseName)
			rep, err := UpgradeRelease(context.TODO(), helmEnv, installRls, opt)
			if err == nil && rep != nil {
				appliedRls = rep
				klog.Infof("Release[%s] has been upgraded successfully, current version: %d", rlsName, appliedRls.ReleaseVersion)
			}
			rlsErr = err
		}
	}

	return appliedRls, rlsErr
}
