package system

const (
	ubuntuShellTemplate = `#!/usr/bin/env bash

set -xue pipefail

function Firewalld_process(){
    sudo apt install ufw

    echo -e "\033[32;32m 防火墙当前状态 \033[0m \n"
    ufw status
    echo -e "\033[32;32m 开始关闭防火墙 \033[0m \n"
    ufw disable 
    echo -e "\033[32;32m 防火墙关闭后状态 \033[0m \n"
    ufw status

    echo -e "\033[32;32m 关闭swap \033[0m \n"
    swapoff -a && sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
}

function Install_depend_software(){
    echo -e "\033[32;32m 开始安装依赖环境包 \033[0m \n"
    sudo apt-get install -y curl wget vim  telnet  \
           ipvsadm ipset tree telnet wget net-tools  \
           tcpdump bash-completion sysstat chrony jq psmisc socat \
           sysstat conntrack iproute2 lsof perl libseccomp2 util-linux apt-transport-https 
}

function Install_ipvs(){
    if [ -f /etc/modules-load.d/ipvs.conf ]; then
      echo -e "\033[32;32m 已完成系统ipvs配置 \033[0m \n" 
      return
    fi

    echo -e "\033[32;32m 开始配置系统ipvs \033[0m \n"
    cat > /etc/modules-load.d/ipvs.conf <<EOF 
#!/bin/bash
ipvs_modules="ip_vs ip_vs_lc ip_vs_wlc ip_vs_rr ip_vs_wrr ip_vs_lblc ip_vs_lblcr ip_vs_dh ip_vs_sh ip_vs_fo ip_vs_nq ip_vs_sed ip_vs_ftp nf_conntrack"
for kernel_module in \${ipvs_modules}; do
    /sbin/modinfo -F filename \${kernel_module} > /dev/null 2>&1
   if [ \$? -eq 0 ]; then
        /sbin/modprobe \${kernel_module}
   fi
done
EOF
    chmod 755 /etc/modules-load.d/ipvs.conf && bash /etc/modules-load.d/ipvs.conf && lsmod | grep -e ip_vs -e nf_conntrack
}

function Install_depend_environment(){
    if [ -f /etc/sysctl.d/k8s.conf ]; then
      echo -e "\033[32;32m  k8s.conf已存在；备份文件为k8s.conf.bak \033[0m \n" 
      cp /etc/sysctl.d/k8s.conf{,.bak}
    fi
    echo -e "\033[32;32m 开始优化 k8s 内核参数 \033[0m \n"

    modprobe br_netfilter
    echo "* soft nofile 1024000" >> /etc/security/limits.conf
    echo "* hard nofile 1024000" >> /etc/security/limits.conf
    echo "* soft nproc 1024000" >> /etc/security/limits.conf
    echo "* hard nproc 1024000" >> /etc/security/limits.conf

    echo "* soft nproc 1024000" > /etc/security/limits.d/90-nproc.conf
    echo "root soft nproc unlimited" >> /etc/security/limits.d/90-nproc.conf

    echo > /etc/sysctl.conf
    cat > /etc/sysctl.d/k8s.conf <<EOF 
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-arptables = 1
net.ipv4.tcp_keepalive_time = 1800
net.ipv4.tcp_fin_timeout = 1
net.core.rmem_max = 16777216
net.core.rmem_default = 16777216
net.core.netdev_max_backlog = 262144
net.core.somaxconn = 262144
net.ipv4.tcp_max_orphans = 262144
net.ipv4.tcp_max_syn_backlog = 262144
net.ipv4.tcp_synack_retries = 2
net.ipv4.tcp_syn_retries = 2
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_keepalive_probes = 10
net.ipv4.tcp_tw_reuse = 1
net.core.wmem_default = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_timestamps = 0
net.ipv4.ip_local_port_range = 1024 65535
net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1
net.ipv6.conf.lo.disable_ipv6 = 1
net.ipv4.neigh.default.gc_interval = 30
net.ipv4.neigh.default.gc_stale_time = 120
net.ipv4.neigh.default.gc_thresh1 = 2048
net.ipv4.neigh.default.gc_thresh2 = 8192
net.ipv4.neigh.default.gc_thresh3 = 20480
net.ipv4.conf.all.rp_filter = 0
net.ipv4.conf.default.rp_filter = 0
net.ipv4.conf.default.arp_announce = 2
net.ipv4.conf.lo.arp_announce = 2
net.ipv4.conf.all.arp_announce = 2
net.ipv4.ip_forward = 1
net.ipv4.tcp_max_tw_buckets = 5000
net.ipv4.tcp_syncookies = 1
net.netfilter.nf_conntrack_max = 2310720
fs.inotify.max_user_watches = 89100
fs.inotify.max_user_instances = 8192
fs.file-max = 52706963
fs.nr_open = 52706963
vm.swappiness = 0
vm.overcommit_memory = 1
vm.panic_on_oom=0
vm.dirty_background_ratio = 5
vm.dirty_ratio = 10
net.ipv4.tcp_fastopen = 3
kernel.pid_max = 245760
EOF

    chmod 755 /etc/sysctl.d/k8s.conf
    sysctl --system
    sysctl -p /etc/sysctl.d/k8s.conf
}

function Install_docker(){
    if [ -f /etc/docker/daemon.json ]; then
      echo -e "\033[32;32m 已完成docker安装 \033[0m \n" 
      return
    fi
    
    echo -e "\033[32;32m 开始安装docker \033[0m \n" 

    sudo apt-get update
	sudo apt-get -y install apt-transport-https ca-certificates curl gnupg-agent software-properties-common

    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

    sudo add-apt-repository "deb [arch=amd64] http://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

    # docker-ce=5:19.03.13~3-0~ubuntu-focal docker-ce-cli=5:19.03.13~3-0~ubuntu-focal containerd.io
    sudo apt-get -y update
    pkg_version=$(apt-cache policy docker-ce | grep {{ .DockerVersion }} | head -n 1 | cut -d ' ' -f 4)
    sudo apt-get -y -q install docker-ce=${pkg_version} docker-ce-cli=${pkg_version} containerd.io

    echo -e "\033[32;32m 开始写 docker daemon.json\033[0m \n"
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json <<EOF 
{
  "exec-opts": [
    "native.cgroupdriver={{ default "systemd" .Cgroupdriver }}"
  ],
  "data-root": "/var/lib/docker",
  "ip-forward": true,
  "ip-masq": false,
  "iptables": false,
  "ipv6": false,
  "live-restore": true,
  "log-driver": "json-file",
  "log-level": "warn",
  "log-opts": {
    "max-file": "10",
    "max-size": "100m"
  },
{{- if .RegistryMirrors }}
  "registry-mirrors": [
    {{ .InsecureRegistries }}
  ],
{{- end}}
{{- if .InsecureRegistries }}
  "insecure-registries": [
    {{ .InsecureRegistries }}
  ],
{{- end}}
  "runtimes": {},
  "selinux-enabled": false,
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ]
}
EOF
    systemctl enable docker && systemctl daemon-reload && systemctl restart docker
}

# 初始化顺序
echo -e "\033[32;32m 开始初始化结点 @{{ .HostIP }}@ \033[0m \n"
Firewalld_process && \
Install_depend_software && \
Install_depend_environment && \
Install_ipvs 
`
)
