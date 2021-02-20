package system

const (
	centosShellTemplate = `
#!/usr/bin/env bash

set -xeuo pipefail

function Update_yumrepo() {
    mkdir -p /etc/yum.repos.d/repoBakDir
    mv /etc/yum.repos.d/*.repo /etc/yum.repos.d/repoBakDir/
	rm -rvf /etc/yum.repos.d/*.repo
    curl https://mirrors.aliyun.com/repo/epel-7.repo -o /etc/yum.repos.d/epel-7.repo
    curl https://mirrors.aliyun.com/repo/Centos-{{ default "7" .CentosVersion }}.repo -o /etc/yum.repos.d/Centos-Base.repo
}

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
    yum -y --nogpgcheck install  yum-utils device-mapper-persistent-data lvm2 \
           curl wget vim telnet ipvsadm tc ipset tree telnet wget net-tools  \
           tcpdump bash-completion sysstat chrony jq psmisc socat \
           sysstat conntrack iproute dstat lsof perl 
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
    cat << EOF | tee /etc/sysctl.d/k8s.conf
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
fs.may_detach_mounts = 1
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
    chattr +i /etc/sysctl.d/k8s.conf
    sysctl --system
    sysctl -p /etc/sysctl.d/k8s.conf
    systemctl enable chronyd && systemctl start chronyd && chronyc sources
}

# 初始化顺序
echo -e "\033[32;32m 开始初始化结点 @{{ .HostIP }}@ \033[0m \n"
Update_yumrepo && \
Firewalld_process && \
Install_depend_software && \
Install_ipvs && \
Install_depend_environment && \
Update_kernel 
`
)
