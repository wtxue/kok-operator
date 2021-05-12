package config

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/pflag"
	"github.com/wtxue/kok-operator/pkg/constants"
)

type Config struct {
	Registry           Registry
	Audit              Audit
	Feature            Feature
	Kubelet            Kubelet           `yaml:"kubelet"`
	APIServer          APIServer         `yaml:"apiServer"`
	ControllerManager  ControllerManager `yaml:"controllerManager"`
	Scheduler          Scheduler         `yaml:"scheduler"`
	SupportK8sVersion  []string
	CustomRegistry     string
	EnableCustomCert   bool
	EnableCustomImages bool
	EnableOnKube       bool
	EnableHostNetwork  bool
}

type Registry struct {
	Prefix    string
	IP        string
	Domain    string
	Namespace string
}

type Audit struct {
	Address string
}

type Feature struct {
	SkipConditions []string
}

type Kubelet struct {
	ExtraArgs map[string]string `yaml:"extraArgs"`
}

type APIServer struct {
	ExtraArgs map[string]string `yaml:"extraArgs"`
}

type ControllerManager struct {
	ExtraArgs map[string]string `yaml:"extraArgs"`
}

type Scheduler struct {
	ExtraArgs map[string]string `yaml:"extraArgs"`
}

func NewDefaultConfig() *Config {
	return &Config{
		CustomRegistry:     "registry.aliyuncs.com/google_containers", // "docker.io/wtxue"
		SupportK8sVersion:  constants.K8sVersions,
		EnableOnKube:       true,
		EnableCustomImages: false,
		EnableHostNetwork:  false,
	}
}

func (r *Config) NeedSetHosts() bool {
	return r.Registry.IP != ""
}

func (r *Config) ImageFullName(name, tag string) string {
	b := new(bytes.Buffer)
	b.WriteString(name)
	if tag != "" {
		if !strings.Contains(tag, "v") {
			b.WriteString(":" + "v" + tag)
		} else {
			b.WriteString(":" + tag)
		}
	}

	s := strings.Split(r.CustomRegistry, "/")
	r.Registry.Domain = s[0]
	r.Registry.Namespace = s[1]
	return path.Join(r.Registry.Domain, r.Registry.Namespace, b.String())
}

func (r *Config) KubeAllImageFullName(name, tag string) string {
	if !strings.Contains(tag, "v") {
		tag = "v" + tag
	}

	return fmt.Sprintf("%s/%s:%s", r.CustomRegistry, name, tag)
}

func (r *Config) KubeProxyImagesName(tag string) string {
	if !strings.Contains(tag, "v") {
		tag = "v" + tag
	}

	return fmt.Sprintf("%s/%s:%s", r.CustomRegistry, "kube-proxy", tag)
}

func (r *Config) IsK8sSupport(version string) bool {
	for _, v := range r.SupportK8sVersion {
		if v == version {
			return true
		}
	}

	return false
}

func (r *Config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&r.CustomRegistry, "images-prefix", r.CustomRegistry, "the images prefix")
	fs.BoolVar(&r.EnableOnKube, "enable-onkube", r.EnableOnKube, "if true, the cluster manager will use on kube apiserver")
	fs.BoolVar(&r.EnableHostNetwork, "enable-host-network", r.EnableHostNetwork, "if true, the kube-apiserver pod use hostNetwork")
	fs.BoolVar(&r.EnableCustomImages, "enable-custom-images", r.EnableCustomImages, "enable custom images")
	fs.StringArrayVar(&r.SupportK8sVersion, "support-k8s-version", r.SupportK8sVersion, "the support k8s version")
}
