package system

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	"github.com/wtxue/kok-operator/pkg/util/template"
	"k8s.io/klog"
)

type Option struct {
	InsecureRegistries string
	RegistryDomain     string
	Options            string
	K8sVersion         string
	DockerVersion      string
	Cgroupdriver       string
	HostIP             string
	KernelRepo         string
	ResolvConf         string
	CentosVersion      string
	ExtraArgs          map[string]string
}

func Install(s ssh.Interface, c *common.Cluster) error {
	dockerVersion := "19.03.9"
	if v, ok := c.Spec.DockerExtraArgs["version"]; ok {
		dockerVersion = v
	}
	option := &Option{
		K8sVersion:    c.Spec.Version,
		DockerVersion: dockerVersion,
		Cgroupdriver:  "systemd", // cgroupfs or systemd
		ExtraArgs:     c.Spec.KubeletExtraArgs,
		HostIP:        s.HostIP(),
	}

	initData, err := template.ParseString(initShellTemplate, option)
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(initData), constants.SystemInitFile)
	if err != nil {
		return err
	}

	klog.Infof("node: %s start exec init system ... ", option.HostIP)
	cmd := fmt.Sprintf("chmod a+x %s && %s", constants.SystemInitFile, constants.SystemInitFile)
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		klog.Errorf("%q %+v", exit, err)
		return errors.Wrapf(err, "node: %s exec init", option.HostIP)
	}

	klog.Infof("node: %s exec init system success", option.HostIP)
	result, err := s.CombinedOutput("uname -r")
	if err != nil {
		klog.Errorf("err: %+v", err)
		return err
	}
	versionStr := strings.TrimSpace(string(result))
	versions := strings.Split(strings.TrimSpace(string(result)), ".")
	if len(versions) < 2 {
		return errors.Errorf("parse version error:%s", versionStr)
	}
	kernelVersion, err := strconv.Atoi(versions[0])
	if err != nil {
		return errors.Wrapf(err, "parse kernelVersion")
	}

	if kernelVersion >= 4 {
		return nil
	}

	klog.Infof("node: %s now kernel: %s,  start reboot ... ", option.HostIP, string(result))
	_, _ = s.CombinedOutput("reboot")
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
