package system

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	"github.com/wtxue/kok-operator/pkg/util/template"
)

type Option struct {
	// InsecureRegistries string
	// RegistryMirrors    string // "https://yqdzw3p0.mirror.aliyuncs.com"
	// RegistryDomain     string // "download.docker.com" or "mirrors.aliyun.com"
	Options    string
	K8sVersion string
	// ContainerdVersion  string
	// Cgroupdriver       string // cgroupfs or systemd
	HostIP        string
	KernelRepo    string
	ResolvConf    string
	CentosVersion string
	ExtraArgs     map[string]string
}

func shellTemplate(ctx *common.ClusterContext) string {
	switch ctx.Cluster.Spec.OSType {
	case devopsv1.DebianType:
		return debianShellTemplate
	case devopsv1.UbuntuType:
		return ubuntuShellTemplate
	default:
		return centosShellTemplate
	}
}

func Install(ctx *common.ClusterContext, s ssh.Interface) error {
	option := &Option{
		K8sVersion: ctx.Cluster.Spec.Version,
		HostIP:     s.HostIP(),
	}

	_, _, _, err := s.Execf("hostnamectl set-hostname %s", s.HostIP())
	if err != nil {
		return err
	}

	initData, err := template.ParseString(shellTemplate(ctx), option)
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(initData), constants.SystemInitFile)
	if err != nil {
		return err
	}

	ctx.Info("start exec init system ... ", "node", option.HostIP)
	cmd := fmt.Sprintf("chmod a+x %s && %s", constants.SystemInitFile, constants.SystemInitFile)
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		ctx.Error(err, "exit", exit, "node", option.HostIP)
		return errors.Wrapf(err, "node: %s exec init", option.HostIP)
	}

	ctx.Info("system init successfully", "node", option.HostIP)
	return nil
}

func CopyFile(s ssh.Interface, file *devopsv1.File) error {
	if ok, err := s.Exist(file.Dst); err == nil && ok {
		return nil
	}

	err := s.CopyFile(file.Src, file.Dst)
	if err != nil {
		return err
	}

	if strings.Contains(file.Dst, "bin") {
		_, _, _, err = s.Execf("chmod a+x %s", file.Dst)
		if err != nil {
			return err
		}
	}

	return nil
}
