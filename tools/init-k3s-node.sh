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
    yum makecache fast
    yum -y --nogpgcheck install nfs-utils curl yum-utils device-mapper-persistent-data lvm2 \
           net-tools conntrack-tools wget vim  ntpdate libseccomp libtool-ltdl telnet \
           ipvsadm tc ipset bridge-utils tree telnet wget net-tools  \
           tcpdump bash-completion sysstat chrony jq psmisc socat \
           cri-o sysstat conntrack  iproute dstat lsof perl bind-utils
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
    yum makecache fast
    yum install -y docker-ce-${DockerVersion} docker-ce-cli-${DockerVersion}

    echo -e "\033[32;32m 开始写 docker daemon.json\033[0m \n"
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json <<EOF
{
  "exec-opts": [
    "native.cgroupdriver=systemd"
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

echo -e "\033[32;32m 开始初始化 k3s 结点 \033[0m \n"
Firewalld_process && \
Install_depend_software && \
Install_ipvs && \
Install_depend_environment && \
Install_docker
