#!/usr/bin/env bash

set -xeuo pipefail

DockerVersion=${1:-"19.03.13"}

function Firewalld_process() {
    grep SELINUX=disabled /etc/selinux/config && echo -e "\033[32;32m 已关闭防火墙，退出防火墙设置 \033[0m \n" && return

    echo -e "\033[32;32m 关闭防火墙 \033[0m \n"
    systemctl stop firewalld && systemctl disable firewalld

    echo -e "\033[32;32m 关闭selinux \033[0m \n"
    setenforce 0
    sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config
    echo -e "\033[32;32m 关闭swap \033[0m \n"
    swapoff -a && sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
}

function Install_depend_software(){
    echo -e "\033[32;32m 开始安装依赖环境包 \033[0m \n"
    yum -y --nogpgcheck install curl yum-utils device-mapper-persistent-data lvm2 \
           conntrack-tools wget vim  libseccomp libtool-ltdl  \
           ipvsadm ipset tree telnet wget net-tools  \
           tcpdump bash-completion sysstat chrony jq psmisc socat \
           sysstat conntrack iproute dstat lsof
}

function Install_ipvs(){
    if [ -f /etc/sysconfig/modules/ipvs.modules ]; then
      echo -e "\033[32;32m 已完成系统ipvs配置 \033[0m \n"
      return
    fi

    echo -e "\033[32;32m 开始配置系统ipvs \033[0m \n"
    cat <<EOF |tee /etc/sysconfig/modules/ipvs.modules
#!/bin/bash
ipvs_modules="ip_vs ip_vs_lc ip_vs_wlc ip_vs_rr ip_vs_wrr ip_vs_lblc ip_vs_lblcr ip_vs_dh ip_vs_sh ip_vs_fo ip_vs_nq ip_vs_sed ip_vs_ftp nf_conntrack"
for kernel_module in \${ipvs_modules}; do
    /sbin/modinfo -F filename \${kernel_module} > /dev/null 2>&1
   if [ \$? -eq 0 ]; then
        /sbin/modprobe \${kernel_module}
   fi
done
EOF
    chmod 755 /etc/sysconfig/modules/ipvs.modules && bash /etc/sysconfig/modules/ipvs.modules && lsmod | grep -e ip_vs -e nf_conntrack
}

function Install_docker(){
    if [ -f /etc/docker/daemon.json ]; then
      echo -e "\033[32;32m 已完成docker安装 \033[0m \n"
      return
    fi

    echo -e "\033[32;32m 开始安装docker \033[0m \n"
    yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
    # centos8 need ?
    yum install -y https://download.docker.com/linux/fedora/30/x86_64/stable/Packages/containerd.io-1.2.6-3.3.fc30.x86_64.rpm

    yum install -y docker-ce-${DockerVersion} docker-ce-cli-${DockerVersion} containerd.io

    echo -e "\033[32;32m 开始写 docker daemon.json\033[0m \n"
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json <<EOF
{
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
    "max-size": "50m"
  },
  "registry-mirrors": [
    "https://mirror.ccs.tencentyun.com",
    "https://4xr1qpsp.mirror.aliyuncs.com"
  ],
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

# --- write systemd service file ---
function Install_k3s_service() {
    echo -e "\033[32;32m 开始写 /lib/systemd/system/etcd.service \033[0m \n"

    cat > /lib/systemd/system/etcd.service <<EOF
[Unit]
Description=Etcd Server
After=network.target
After=network-online.target
Wants=network-online.target
Documentation=https://github.com/coreos/etcd

[Service]
Type=notify
User=root
ExecStart=/usr/local/bin/etcd --data-dir=/var/lib/etcd/
Restart=always
LimitNOFILE=65536
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
    systemctl enable etcd && systemctl daemon-reload && systemctl restart etcd

    echo -e "\033[32;32m 开始写 /lib/systemd/system/k3s.service \033[0m \n"
    cat > /lib/systemd/system/k3s.service <<EOF
[Unit]
Description=Lightweight Kubernetes
Documentation=https://k3s.io
Wants=network-online.target
After=network-online.target etcd.service

[Service]
Type=notify
Environment="K3S_TYPE=server"
Environment="K3S_RUNTIME=--docker"
Environment="K3S_DATASTORE=--datastore-endpoint=http://localhost:2379"
KillMode=process
Delegate=yes
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
TimeoutStartSec=0
Restart=always
RestartSec=5s
ExecStartPre=-/sbin/modprobe br_netfilter
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/local/bin/k3s $K3S_TYPE $K3S_RUNTIME $K3S_DATASTORE

[Install]
WantedBy=multi-user.target
EOF
    systemctl enable k3s && systemctl daemon-reload && systemctl restart k3s
}

echo -e "\033[32;32m 开始初始化 k3s 结点 \033[0m \n"
Firewalld_process && \
Install_depend_software && \
Install_ipvs && \
Install_docker && \
Install_k3s_service
