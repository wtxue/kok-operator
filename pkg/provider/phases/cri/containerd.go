package cri

import (
	"fmt"

	"bytes"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	"github.com/wtxue/kok-operator/pkg/util/template"
	"strings"
)

const ContainerdConfigTemplate = `
[plugins]
  [plugins."io.containerd.internal.v1.opt"]
    path = "/opt/containerd"
  [plugins."io.containerd.grpc.v1.cri"]
    sandbox_image = {{ default "k8s.gcr.io/pause:3.2" .PauseImage }}
    [plugins."io.containerd.grpc.v1.cri".containerd]
      snapshotter = {{ default "overlayfs" .Snapshotter }}
{{ if .PrivateRegistryConfig }}
    [plugins."io.containerd.grpc.v1.cri".registry]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
{{range $k, $v := .PrivateRegistryConfig.Mirrors }}
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."{{$k}}"]
          endpoint = [{{range $i, $j := $v.Endpoints}}{{if $i}}, {{end}}{{printf "%q" .}}{{end}}]
{{end}}
{{end}}
`

type ContainerdConfig struct {
	Registry    *devopsv1.Registry
	PauseImage  string
	Snapshotter string
}

func InstallCRI(ctx *common.ClusterContext, s ssh.Interface) error {
	// dir := "bin/linux/" // local debug config dir
	otherDir := "/k8s/bin/"
	if dir := constants.GetAnnotationKey(ctx.Cluster.Annotations, constants.ClusterAnnoLocalDebugDir); len(dir) > 0 {
		otherDir = dir + otherDir
	}

	var CopyList = []devopsv1.File{
		{
			Src: otherDir + "containerd.tar.gz",
			Dst: "/opt/k8s/containerd.tar.gz",
		},
	}

	for _, ls := range CopyList {
		// if ok, err := s.Exist(ls.Dst); err == nil && ok {
		// 	ctx.Info("file exist ignoring", "node", s.HostIP(), "dst", ls.Dst)
		// 	continue
		// }

		err := s.CopyFile(ls.Src, ls.Dst)
		if err != nil {
			ctx.Error(err, "CopyFile", "node", s.HostIP(), "src", ls.Src)
			return err
		}

		if strings.Contains(ls.Dst, "containerd") {
			cmd := "mkdir -p /usr/local/bin /usr/local/sbin /etc/systemd/system /opt/k8s/containerd && " +
				"tar -C /opt/k8s/containerd -xzf /opt/k8s/containerd.tar.gz && " +
				"cp -rf /opt/k8s/containerd/usr/local/sbin/* /usr/local/sbin/ && " +
				"cp -rf /opt/k8s/containerd/usr/local/bin/* /usr/local/bin/ && " +
				"cp -rf /opt/k8s/containerd/etc/crictl.yaml /etc/ && " +
				"cp -rf /opt/k8s/containerd/etc/systemd/system/containerd.service /etc/systemd/system/ && " +
				"rm -f /usr/local/bin/critest &&" +
				"rm -rf /opt/k8s/containerd"

			_, err := s.CombinedOutput(cmd)
			if err != nil {
				return err
			}
		}
		ctx.Info("copy successfully", "node", s.HostIP(), "path", ls.Dst)
	}

	if ctx.Cluster.Spec.Registry != nil {
		config := &ContainerdConfig{
			Registry:    ctx.Cluster.Spec.Registry,
			PauseImage:  "",
			Snapshotter: "",
		}

		configData, err := template.ParseString(ContainerdConfigTemplate, config)
		if err != nil {
			return err
		}

		// mkdir -p /etc/containerd && containerd config default > /etc/containerd/config.toml
		_, _, _, err = s.Execf("mkdir -p /etc/containerd")
		if err != nil {
			return err
		}

		ctx.Info("start write containerd config", "node", s.HostIP(), "config", string(configData))
		err = s.WriteFile(bytes.NewReader(configData), "/etc/containerd/config.toml")
		if err != nil {
			return err
		}
	}

	// systemctl enable containerd && systemctl daemon-reload && systemctl restart containerd
	cmd := "systemctl enable containerd && systemctl daemon-reload && systemctl restart containerd"
	if _, stderr, exit, err := s.Execf(cmd); err != nil || exit != 0 {
		cmd = "journalctl --unit containerd -n10 --no-pager"
		jStdout, _, jExit, jErr := s.Execf(cmd)
		if jErr != nil || jExit != 0 {
			return fmt.Errorf("exec %q:error %s", cmd, err)
		}
		ctx.Info("log", "cmd", cmd, "stdout", jStdout)

		return fmt.Errorf("Exec %s failed:exit %d:stderr %s:error %s:log:\n%s", cmd, exit, stderr, err, jStdout)
	}

	ctx.Info("exec successfully", "node", s.HostIP(), "cmd", cmd)
	return nil
}
