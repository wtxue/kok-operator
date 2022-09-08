package cri

import (
	"bytes"
	"fmt"
	"strings"

	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	"github.com/wtxue/kok-operator/pkg/util/ssh"
	"github.com/wtxue/kok-operator/pkg/util/template"
)

const ContainerdConfigTemplate = `disabled_plugins = []
imports = []
oom_score = 0
plugin_dir = ""
required_plugins = []
root = "/var/lib/containerd"
state = "/run/containerd"
temp = ""
version = 2

[cgroup]
  path = ""

[debug]
  address = ""
  format = ""
  gid = 0
  level = ""
  uid = 0

[grpc]
  address = "/run/containerd/containerd.sock"
  gid = 0
  max_recv_message_size = 16777216
  max_send_message_size = 16777216
  tcp_address = ""
  tcp_tls_ca = ""
  tcp_tls_cert = ""
  tcp_tls_key = ""
  uid = 0

[metrics]
  address = ""
  grpc_histogram = false

[plugins]

  [plugins."io.containerd.gc.v1.scheduler"]
    deletion_threshold = 0
    mutation_threshold = 100
    pause_threshold = 0.02
    schedule_delay = "0s"
    startup_delay = "100ms"

  [plugins."io.containerd.grpc.v1.cri"]
    device_ownership_from_security_context = false
    disable_apparmor = false
    disable_cgroup = false
    disable_hugetlb_controller = true
    disable_proc_mount = false
    disable_tcp_service = true
    enable_selinux = false
    enable_tls_streaming = false
    enable_unprivileged_icmp = false
    enable_unprivileged_ports = false
    ignore_image_defined_volumes = false
    max_concurrent_downloads = 3
    max_container_log_line_size = 16384
    netns_mounts_under_state_dir = false
    restrict_oom_score_adj = false
    sandbox_image = "{{ default "docker.io/wtxue/pause:3.7" .PauseImage }}"
    selinux_category_range = 1024
    stats_collect_period = 10
    stream_idle_timeout = "4h0m0s"
    stream_server_address = "127.0.0.1"
    stream_server_port = "0"
    systemd_cgroup = false
    tolerate_missing_hugetlb_controller = true
    unset_seccomp_profile = ""

    [plugins."io.containerd.grpc.v1.cri".cni]
      bin_dir = "/opt/cni/bin"
      conf_dir = "/etc/cni/net.d"
      conf_template = ""
      ip_pref = ""
      max_conf_num = 1

    [plugins."io.containerd.grpc.v1.cri".containerd]
      default_runtime_name = "runc"
      disable_snapshot_annotations = true
      discard_unpacked_layers = false
      ignore_rdt_not_enabled_errors = false
      no_pivot = false
      snapshotter = "overlayfs"

      [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime]
        base_runtime_spec = ""
        cni_conf_dir = ""
        cni_max_conf_num = 0
        container_annotations = []
        pod_annotations = []
        privileged_without_host_devices = false
        runtime_engine = ""
        runtime_path = ""
        runtime_root = ""
        runtime_type = ""

        [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime.options]

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]

        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
          base_runtime_spec = ""
          cni_conf_dir = ""
          cni_max_conf_num = 0
          container_annotations = []
          pod_annotations = []
          privileged_without_host_devices = false
          runtime_engine = ""
          runtime_path = ""
          runtime_root = ""
          runtime_type = "io.containerd.runc.v2"

          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
            BinaryName = ""
            CriuImagePath = ""
            CriuPath = ""
            CriuWorkPath = ""
            IoGid = 0
            IoUid = 0
            NoNewKeyring = false
            NoPivotRoot = false
            Root = ""
            ShimCgroup = ""
            SystemdCgroup = true

      [plugins."io.containerd.grpc.v1.cri".containerd.untrusted_workload_runtime]
        base_runtime_spec = ""
        cni_conf_dir = ""
        cni_max_conf_num = 0
        container_annotations = []
        pod_annotations = []
        privileged_without_host_devices = false
        runtime_engine = ""
        runtime_path = ""
        runtime_root = ""
        runtime_type = ""

        [plugins."io.containerd.grpc.v1.cri".containerd.untrusted_workload_runtime.options]

    [plugins."io.containerd.grpc.v1.cri".image_decryption]
      key_model = "node"

    [plugins."io.containerd.grpc.v1.cri".registry]
      config_path = ""

      [plugins."io.containerd.grpc.v1.cri".registry.auths]

      [plugins."io.containerd.grpc.v1.cri".registry.configs]

      [plugins."io.containerd.grpc.v1.cri".registry.headers]

      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]

{{- if .PrivateRegistryConfig }}
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
{{- range $k, $v := .PrivateRegistryConfig.Mirrors }}
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."{{$k}}"]
          endpoint = [{{range $i, $j := $v.Endpoints}}{{if $i}}, {{end}}{{printf "%q" .}}{{end}}]
{{- end}}
{{- end }}

    [plugins."io.containerd.grpc.v1.cri".x509_key_pair_streaming]
      tls_cert_file = ""
      tls_key_file = ""

  [plugins."io.containerd.internal.v1.opt"]
    path = "/opt/containerd"

  [plugins."io.containerd.internal.v1.restart"]
    interval = "10s"

  [plugins."io.containerd.internal.v1.tracing"]
    sampling_ratio = 1.0
    service_name = "containerd"

  [plugins."io.containerd.metadata.v1.bolt"]
    content_sharing_policy = "shared"

  [plugins."io.containerd.monitor.v1.cgroups"]
    no_prometheus = false

  [plugins."io.containerd.runtime.v1.linux"]
    no_shim = false
    runtime = "runc"
    runtime_root = ""
    shim = "containerd-shim"
    shim_debug = false

  [plugins."io.containerd.runtime.v2.task"]
    platforms = ["linux/amd64"]
    sched_core = false

  [plugins."io.containerd.service.v1.diff-service"]
    default = ["walking"]

  [plugins."io.containerd.service.v1.tasks-service"]
    rdt_config_file = ""

  [plugins."io.containerd.snapshotter.v1.aufs"]
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.btrfs"]
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.devmapper"]
    async_remove = false
    base_image_size = ""
    discard_blocks = false
    fs_options = ""
    fs_type = ""
    pool_name = ""
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.native"]
    root_path = ""

  [plugins."io.containerd.snapshotter.v1.overlayfs"]
    root_path = ""
    upperdir_label = false

  [plugins."io.containerd.snapshotter.v1.zfs"]
    root_path = ""

  [plugins."io.containerd.tracing.processor.v1.otlp"]
    endpoint = ""
    insecure = false
    protocol = ""

[proxy_plugins]

[stream_processors]

  [stream_processors."io.containerd.ocicrypt.decoder.v1.tar"]
    accepts = ["application/vnd.oci.image.layer.v1.tar+encrypted"]
    args = ["--decryption-keys-path", "/etc/containerd/ocicrypt/keys"]
    env = ["OCICRYPT_KEYPROVIDER_CONFIG=/etc/containerd/ocicrypt/ocicrypt_keyprovider.conf"]
    path = "ctd-decoder"
    returns = "application/vnd.oci.image.layer.v1.tar"

  [stream_processors."io.containerd.ocicrypt.decoder.v1.tar.gzip"]
    accepts = ["application/vnd.oci.image.layer.v1.tar+gzip+encrypted"]
    args = ["--decryption-keys-path", "/etc/containerd/ocicrypt/keys"]
    env = ["OCICRYPT_KEYPROVIDER_CONFIG=/etc/containerd/ocicrypt/ocicrypt_keyprovider.conf"]
    path = "ctd-decoder"
    returns = "application/vnd.oci.image.layer.v1.tar+gzip"

[timeouts]
  "io.containerd.timeout.bolt.open" = "0s"
  "io.containerd.timeout.shim.cleanup" = "5s"
  "io.containerd.timeout.shim.load" = "5s"
  "io.containerd.timeout.shim.shutdown" = "3s"
  "io.containerd.timeout.task.state" = "2s"

[ttrpc]
  address = ""
  gid = 0
  uid = 0
`

type ContainerdConfig struct {
	PrivateRegistryConfig *devopsv1.Registry
	PauseImage            string
}

func InstallCRI(ctx *common.ClusterContext, s ssh.Interface) error {
	// dir := "bin/linux/" // local debug config dir
	otherDir := "/k8s/bin/"
	if dir := constants.GetMapKey(ctx.Cluster.Annotations, constants.ClusterDebugLocalDir); len(dir) > 0 {
		otherDir = dir + otherDir
	}

	var CopyList = []devopsv1.File{
		{
			Src: otherDir + "crictl",
			Dst: "/usr/local/bin/crictl",
		},
		{
			Src: otherDir + "containerd.tar.gz",
			Dst: "/opt/k8s/containerd.tar.gz",
		},
		{
			Src: otherDir + "runc",
			Dst: "/usr/local/sbin/runc",
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

		if strings.Contains(ls.Dst, "crictl") {
			_, err := s.CombinedOutput("chmod a+x /usr/local/bin/crictl")
			if err != nil {
				return err
			}
		}

		if strings.Contains(ls.Dst, "runc") {
			_, err := s.CombinedOutput("chmod a+x /usr/local/sbin/runc")
			if err != nil {
				return err
			}
		}

		if strings.Contains(ls.Dst, "containerd") {
			cmd := "mkdir -p /usr/local/bin /usr/local/sbin /etc/systemd/system /opt/k8s/containerd && " +
				"tar -C /opt/k8s/containerd -xzf /opt/k8s/containerd.tar.gz && " +
				"cp -rf /opt/k8s/containerd/usr/local/bin/* /usr/local/bin/ && " +
				"cp -rf /opt/k8s/containerd/etc/crictl.yaml /etc/ && " +
				"cp -rf /opt/k8s/containerd/etc/systemd/system/containerd.service /etc/systemd/system/ && " +
				"rm -rf /opt/k8s/containerd"

			_, err := s.CombinedOutput(cmd)
			if err != nil {
				return err
			}
		}
		ctx.Info("copy successfully", "node", s.HostIP(), "path", ls.Dst)
	}

	config := &ContainerdConfig{
		PrivateRegistryConfig: ctx.Cluster.Spec.Registry,
		PauseImage:            "",
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
