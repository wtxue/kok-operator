module github.com/wtxue/kok-operator

go 1.16

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/banzaicloud/k8s-objectmatcher v1.5.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.4.0
	github.com/goph/emperror v0.17.2
	github.com/imdario/mergo v0.3.11
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.13.0
	github.com/segmentio/ksuid v1.0.3
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/thoas/go-funk v0.8.0
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.5.4
	k8s.io/api v0.20.6
	k8s.io/apiextensions-apiserver v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/apiserver v0.20.6
	k8s.io/cli-runtime v0.20.6
	k8s.io/client-go v0.20.6
	k8s.io/cluster-bootstrap v0.20.6
	k8s.io/component-base v0.20.6
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.8.0
	k8s.io/kube-aggregator v0.20.6
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.3
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
