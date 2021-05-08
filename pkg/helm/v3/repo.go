package v3

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	defaultInterval = 60 * 5
)

// ErrRepoNotFound describe an error if helm repository not found
var ErrRepoNotFound = errors.New("helm repository not found!")

// HelmIndexSyncer sync helm repo index repeatedly
type HelmIndexSyncer struct {
	HelmEnv *HelmEnv

	// interval is the interval of the sync process
	Interval int
}

type HelmEnv struct {
	KubeCache string

	Cli     *cli.EnvSettings
	KubeCli kubernetes.Interface
	genericclioptions.RESTClientGetter
}

// Options struct holding directives for driving helm operations (similar to command line flags)
// extend this as required eventually build a more sophisticated solution for it
type Options struct {
	Namespace    string  `json:"namespace,omitempty"`
	DryRun       bool    `json:"dryRun,omitempty"`
	GenerateName bool    `json:"generateName,omitempty"`
	Wait         bool    `json:"wait,omitempty"`
	Timeout      int64   `json:"timeout,omitempty"`
	ReuseValues  bool    `json:"reuseValues,omitempty"`
	Install      bool    `json:"install,omitempty"`
	Filter       *string `json:"filter,omitempty"`
	SkipCRDs     bool    `json:"skipCRDs,omitempty"`
}

func GetRepoFilePath(env *cli.EnvSettings) string {
	return env.RepositoryConfig
}

func NewChartRepositoryWarp(cfg *repo.Entry, settings *cli.EnvSettings) (*repo.ChartRepository, error) {
	r, err := repo.NewChartRepository(cfg, getter.All(settings))
	if err != nil {
		return nil, err
	}

	// override the wired repository cache
	r.CachePath = settings.RepositoryCache
	return r, nil
}

// InitHelmRepoEnv Generate helm path based on orgName
func InitHelmRepoEnv(organizationName string, repoMap map[string]string) (*HelmEnv, error) {
	settings := &cli.EnvSettings{}

	helmRepoHome := fmt.Sprintf("%s/%s", "./helm3", organizationName)
	settings.RepositoryConfig = path.Join(helmRepoHome, "repositories.yaml")
	settings.RepositoryCache = path.Join(helmRepoHome, "cache")

	_, err := os.Stat(helmRepoHome)
	if os.IsNotExist(err) {
		klog.Infof("Helm directories [%s] not exists", helmRepoHome)
		err := InstallLocalHelm(settings, repoMap)
		if err != nil {
			klog.Errorf("InstallLocalHelm err: %+v", err)
			return nil, err
		}
	} else {
		entries, err := ReposGet(settings)
		if err != nil {
			klog.Errorf("get all repo err: %+v", err)
			return nil, err
		}

		for _, e := range entries {
			err := ReposUpdate(settings, e.Name)
			if err != nil {
				klog.Errorf("update repo: %s err: %+v", e.Name, err)
			}
		}
	}

	return &HelmEnv{
		KubeCache: path.Join(helmRepoHome, "kube"),
		Cli:       settings,
	}, nil
}

// InstallLocalHelm install helm into the given path
func InstallLocalHelm(env *cli.EnvSettings, repoMap map[string]string) error {
	if err := InstallHelmClient(env); err != nil {
		klog.Errorf("install client err: %v", err)
		return err
	}

	if err := ensureDefaultRepos(env, repoMap); err != nil {
		return errors.Wrap(err, "Setting up default repos failed!")
	}
	return nil
}

// DownloadChartFromRepo download a given chart
func DownloadChartFromRepo(name, version string, env *cli.EnvSettings) (string, error) {
	dl := downloader.ChartDownloader{
		RepositoryCache:  env.RepositoryCache,
		RepositoryConfig: env.RepositoryConfig,
		Getters:          getter.All(env),
	}
	if _, err := os.Stat(env.RepositoryConfig); os.IsNotExist(err) {
		klog.Infof("Creating '%s' directory.", env.RepositoryConfig)
		_ = os.MkdirAll(env.RepositoryConfig, 0744)
	}

	klog.Infof("Downloading helm chart %q, version %q to %q", name, version, env.RepositoryConfig)
	filename, _, err := dl.DownloadTo(name, version, env.RepositoryConfig)
	if err == nil {
		lname, err := filepath.Abs(filename)
		if err != nil {
			return filename, errors.Wrapf(err, "Could not create absolute path from %s", filename)
		}
		klog.Infof("Fetched helm chart %q, version %q to %q", name, version, filename)
		return lname, nil
	}

	return filename, errors.Wrapf(err, "Failed to download chart %q, version %q", name, version)
}

// InstallHelmClient Installs helm client on a given path
func InstallHelmClient(env *cli.EnvSettings) error {
	configDirectories := []string{
		env.RepositoryCache,
	}

	klog.Info("Setting up helm directories.")
	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			klog.Infof("Creating '%s'", p)
			if err := os.MkdirAll(p, 0755); err != nil {
				return errors.Wrapf(err, "Could not create '%s'", p)
			}
		} else if !fi.IsDir() {
			return errors.Errorf("'%s' must be a directory", p)
		}
	}

	klog.Info("Initializing helm client succeeded, happy helming!")
	return nil
}

func ensureDefaultRepos(env *cli.EnvSettings, repoMap map[string]string) error {
	for repoName, repoUrl := range repoMap {
		klog.Infof("Setting up helm repo: %s, url: %s", repoName, repoUrl)
		_, err := ReposAdd(
			env,
			&repo.Entry{
				Name: repoName,
				URL:  repoUrl,
			})
		if err != nil {
			return errors.Wrapf(err, "cannot init repo: %s", repoName)
		}
	}

	return nil
}

// ReposGet returns repo
func ReposGet(env *cli.EnvSettings) ([]*repo.Entry, error) {
	repoFile := GetRepoFilePath(env)
	klog.V(5).Infof("helm repo path: %s", repoFile)

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return nil, err
	}
	if len(f.Repositories) == 0 {
		return make([]*repo.Entry, 0), nil
	}

	return f.Repositories, nil
}

// ReposAdd adds repo(s)
func ReposAdd(env *cli.EnvSettings, c *repo.Entry) (bool, error) {
	repoPath := GetRepoFilePath(env)
	var f *repo.File
	if _, err := os.Stat(repoPath); err != nil {
		klog.Infof("Creating %s", repoPath)
		f = repo.NewFile()
	} else {
		f, err = repo.LoadFile(repoPath)
		if err != nil {
			return false, errors.Wrap(err, "Cannot create a new ChartRepo")
		}
		klog.Infof("Profile file %s loaded.", repoPath)
	}

	if f.Has(c.Name) {
		return false, errors.Errorf("repository name (%s) already exists, please specify a different name", c.Name)
	}

	r, err := NewChartRepositoryWarp(c, env)
	if err != nil {
		return false, errors.Wrap(err, "Cannot create a new ChartRepo")
	}

	if _, errIdx := r.DownloadIndexFile(); errIdx != nil {
		return false, errors.Wrap(errIdx, "Repo index download failed")
	}
	f.Update(c)
	// f.Add(&c)
	if errW := f.WriteFile(repoPath, 0644); errW != nil {
		return false, errors.Wrap(errW, "Cannot write helm repo profile file")
	}

	klog.Infof("repo: %s has been added", c.Name)
	return true, nil
}

func removeRepoCache(root, name string) error {
	idx := filepath.Join(root, helmpath.CacheChartsFile(name))
	if _, err := os.Stat(idx); err == nil {
		os.Remove(idx)
	}

	idx = filepath.Join(root, helmpath.CacheIndexFile(name))
	if _, err := os.Stat(idx); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "can't remove index file %s", idx)
	}
	return os.Remove(idx)
}

// ReposDelete deletes repo(s)
func ReposDelete(env *cli.EnvSettings, repoName string) error {
	repoFile := GetRepoFilePath(env)
	klog.Infof("Repo File: %s", repoFile)

	r, err := repo.LoadFile(repoFile)
	if err != nil {
		if os.IsNotExist(errors.Cause(err)) || len(r.Repositories) == 0 {
			klog.Warningf("no repositories configured, nothing to do")
			return nil
		}
		return err
	}

	if !r.Remove(repoName) {
		return ErrRepoNotFound
	}
	if err := r.WriteFile(repoFile, 0644); err != nil {
		return err
	}

	if err := removeRepoCache(env.RepositoryCache, repoName); err != nil {
		return err
	}

	klog.Infof("repo:%s has been removed", repoName)
	return nil
}

// ReposModify modifies repo(s)
func ReposModify(env *cli.EnvSettings, repoName string, newRepo *repo.Entry) error {
	repoFile := GetRepoFilePath(env)
	klog.Infof("Repo File: %s", repoFile)
	klog.Infof("New repo content: %#v", newRepo)

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return err
	}

	if !f.Has(repoName) {
		return ErrRepoNotFound
	}

	var formerRepo *repo.Entry
	repos := f.Repositories
	for _, r := range repos {
		if r.Name == repoName {
			formerRepo = r
		}
	}

	if formerRepo != nil {
		if len(newRepo.Name) == 0 {
			newRepo.Name = formerRepo.Name
			klog.Infof("new repo name field is empty, replaced with: %s", formerRepo.Name)
		}

		if len(newRepo.URL) == 0 {
			newRepo.URL = formerRepo.URL
			klog.Infof("new repo url field is empty, replaced with: %s", formerRepo.URL)
		}
	}

	f.Update(newRepo)

	if errW := f.WriteFile(repoFile, 0644); errW != nil {
		return errors.Wrap(errW, "Cannot write helm repo profile file")
	}
	return nil
}

// ReposUpdate updates a repo(s)
func ReposUpdate(env *cli.EnvSettings, repoName string) error {
	repoFile := GetRepoFilePath(env)
	klog.V(5).Infof("start update repo:%s File: %s", repoName, repoFile)

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return errors.Wrap(err, "Load ChartRepo")
	}

	for _, cfg := range f.Repositories {
		if cfg.Name == repoName {
			r, err := NewChartRepositoryWarp(cfg, env)
			if err != nil {
				return errors.Wrap(err, "Cannot create a new ChartRepo")
			}

			_, errIdx := r.DownloadIndexFile()
			if errIdx != nil {
				return errors.Wrap(errIdx, "Repo index download failed")
			}
			return nil
		}
	}

	return ErrRepoNotFound
}

func matchesFilter(filter string, value string) bool {
	if filter == "" {
		// there is no filter
		return true
	}

	matches, err := regexp.MatchString(filter, strings.ToLower(value))
	if err != nil {
		return false
	}

	return matches
}

func ChartsGet(helmEnv *cli.EnvSettings, repos, chart, Keyword, version string) (map[string][]repo.ChartVersions, error) {
	repoFile := GetRepoFilePath(helmEnv)
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return nil, err
	}

	chartVersionsSlice := make(map[string][]repo.ChartVersions)
	for _, repoEntry := range f.Repositories {
		if !matchesFilter(repos, repoEntry.Name) {
			continue
		}

		repoIndexFilePath := path.Join(helmEnv.RepositoryCache, helmpath.CacheIndexFile(repoEntry.Name))
		repoIndexFile, err := repo.LoadIndexFile(repoIndexFilePath)
		if err != nil {
			return nil, err
		}

		for chartName, chartVersions := range repoIndexFile.Entries {
			filteredChartVersions := make(repo.ChartVersions, 0, 0)

			if !matchesFilter(chart, chartName) {
				continue
			}

			filteredChartVersions = append(filteredChartVersions, chartVersions[0])
			chartVersionsSlice[repoEntry.Name] = append(chartVersionsSlice[repoEntry.Name], filteredChartVersions)
		}
	}

	return chartVersionsSlice, nil
}

func NewDefaultHelmIndexSyncer(helmEnv *HelmEnv) *HelmIndexSyncer {
	return &HelmIndexSyncer{
		HelmEnv:  helmEnv,
		Interval: defaultInterval,
	}
}

func (h *HelmIndexSyncer) Start(stop <-chan struct{}) error {
	wait.Until(func() {
		klog.V(4).Infof("update helm repo index, time: %v", time.Now())
		entrys, err := ReposGet(h.HelmEnv.Cli)
		if err != nil {
			klog.Errorf("get all repo err: %+v", err)
			return
		}

		for _, e := range entrys {
			err := ReposUpdate(h.HelmEnv.Cli, e.Name)
			if err != nil {
				klog.Errorf("update repo: %s err: %+v", e.Name, err)
				return
			}

			// chrts, err := ChartsGet(h.HelmEnv.Cli, e.Name, "", "", "")
			// if err == nil {
			// 	klog.Infof("charts len:%d", len(chrts))
			// }
		}
	}, time.Second*time.Duration(h.Interval), stop)
	return nil
}
