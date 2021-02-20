package kubemisc

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/provider/phases/certs"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
)

const (
	additPolicy = `
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: Metadata
`
)

const (
	tokenFileTemplate = `
%s,admin,admin,system:masters
`
)

type Option struct {
	MasterEndpoint string
	ClusterName    string
	CACert         []byte
	Token          string
}

func GetBindPort(obj *devopsv1.Cluster) int {
	bindPort := 6443
	if obj.Spec.Features.HA != nil && obj.Spec.Features.HA.ThirdPartyHA != nil {
		bindPort = int(obj.Spec.Features.HA.ThirdPartyHA.VPort)
	}

	return bindPort
}

func install(s ssh.Interface, option *Option) error {
	config := CreateWithToken(option.MasterEndpoint, option.ClusterName, "kubernetes-admin", option.CACert, option.Token)
	data, err := runtime.Encode(clientcmdlatest.Codec, config)
	if err != nil {
		return err
	}
	err = s.WriteFile(bytes.NewReader(data), "/root/.kube/config") // fixme ssh not support $HOME or ~
	if err != nil {
		return err
	}

	return nil
}

// Install creates all the requested kubeconfig files.
func Install(ctx *common.ClusterContext, s ssh.Interface) error {
	option := &Option{
		MasterEndpoint: "https://127.0.0.1:6443",
		ClusterName:    ctx.Cluster.Name,
		CACert:         ctx.Credential.CACert,
		Token:          *ctx.Credential.Token,
	}

	return install(s, option)
}

func InstallNode(s ssh.Interface, option *Option) error {
	return install(s, option)
}

func ApplyKubeletKubeconfig(ctx *common.ClusterContext, apiserver string, kubeletNodeAddr string, kubeMaps map[string]string) error {
	if ctx.Credential.CACert == nil {
		return fmt.Errorf("ca is nil")
	}

	cfgMaps, err := certs.CreateKubeletKubeConfigFile(ctx.Credential.CAKey, ctx.Credential.CACert,
		apiserver, kubeletNodeAddr, ctx.Cluster.Name)
	if err != nil {
		return errors.Wrap(err, "create kubeconfg")
	}

	for noPathFile, v := range cfgMaps {
		by, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			return err
		}

		key := filepath.Join(constants.KubernetesDir, noPathFile)
		kubeMaps[key] = string(by)
	}

	return nil
}

func BuildMasterMiscConfigToMap(ctx *common.ClusterContext, apiserver string) error {
	if ctx.Credential.CACert == nil {
		return fmt.Errorf("ca is nil")
	}

	cfgMaps, err := certs.CreateMasterKubeConfigFile(ctx.Credential.CAKey, ctx.Credential.CACert,
		apiserver, ctx.Cluster.Name)
	if err != nil {
		return errors.Wrap(err, "create kubeconfg")
	}

	if ctx.Credential.KubeData == nil {
		ctx.Credential.KubeData = make(map[string]string)
	}

	ctx.Info("start build kubeconfig ...")
	for noPathFile, v := range cfgMaps {
		by, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			return err
		}

		key := filepath.Join(constants.KubernetesDir, noPathFile)
		ctx.Credential.KubeData[key] = string(by)
	}

	key := filepath.Join(constants.KubernetesDir, "audit-policy.yaml")
	ctx.Credential.KubeData[key] = additPolicy

	tokenData := fmt.Sprintf(tokenFileTemplate, *ctx.Credential.Token)
	ctx.Credential.KubeData[constants.TokenFile] = tokenData
	return nil
}

func hasContains(s string, ss []string) bool {
	for _, ts := range ss {
		if strings.HasSuffix(s, ts) {
			return true
		}
	}

	return false
}

func CovertMasterKubeConfig(s ssh.Interface, ctx *common.ClusterContext) error {
	fileMaps := make(map[string]string)

	apiserver := certs.BuildApiserverEndpoint(s.HostIP(), 6443)
	for pathName, va := range ctx.Credential.KubeData {
		if !hasContains(pathName, certs.GetMasterKubeConfigList()) {
			continue
		}

		kcfg := &clientcmdapi.Config{}
		err := certs.DecodeKubeConfigByte([]byte(va), kcfg)
		if err != nil {
			return err
		}

		for _, v := range kcfg.Clusters {
			v.Server = apiserver
		}

		covertByte, err := certs.BuildKubeConfigByte(kcfg)
		if err != nil {
			return err
		}

		fileMaps[pathName] = string(covertByte)
	}

	err := ApplyKubeletKubeconfig(ctx, apiserver, s.HostIP(), fileMaps)
	if err != nil {
		return err
	}

	for name, data := range ctx.Credential.KubeData {
		if strings.Contains(name, "known_tokens.csv") {
			fileMaps[name] = data
			break
		}
	}

	for pathName, va := range fileMaps {
		ctx.Info("start write ...", "node", s.HostIP(), "pathName", pathName)
		err = s.WriteFile(strings.NewReader(va), pathName)
		if err != nil {
			return errors.Wrapf(err, "node: %s failed to write for %s ", s.HostIP(), pathName)
		}
	}

	return nil
}
