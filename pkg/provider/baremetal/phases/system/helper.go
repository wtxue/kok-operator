package system

const (
	initShellTemplate = `
#!/usr/bin/env bash

set -xeuo pipefail

function Firewalld_process() {
    grep SELINUX=disabled /etc/selinux/config && echo -e "\033[32;32m 已关闭防火墙，退出防火墙设置 \033[0m \n" && return

    echo -e "\033[32;32m 关闭防火墙 \033[0m \n"
    systemctl stop firewalld && systemctl disable firewalld

    echo -e "\033[32;32m 关闭selinux \033[0m \n"
    setenforce 0
    sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config
    
    swapoff -a && sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
}

function Install_depend_environment(){
    if [ -f /etc/sysctl.d/k8s.conf ]; then
      echo -e "\033[32;32m 已完成依赖环境安装 \033[0m \n" 
      return
    fi

    echo -e "\033[32;32m 开始安装依赖环境包 \033[0m \n" 
    yum makecache fast
    yum install -y nfs-utils curl yum-utils device-mapper-persistent-data lvm2 \
           net-tools conntrack-tools wget vim  ntpdate libseccomp libtool-ltdl telnet \
           ipvsadm tc ipset bridge-utils tree telnet wget net-tools  \
           tcpdump bash-completion sysstat chrony 

    echo -e "\033[32;32m 开始配置 k8s sysctl \033[0m \n" 
    cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
vm.swappiness=0 # 禁止使用 swap 空间，只有当系统 OOM 时才允许使用它
vm.overcommit_memory=1 # 不检查物理内存是否够用
vm.panic_on_oom=0 # 开启 OOM
fs.inotify.max_user_instances=8192
fs.inotify.max_user_watches=1048576
fs.file-max=52706963
fs.nr_open=52706963
net.ipv6.conf.all.disable_ipv6=1
net.netfilter.nf_conntrack_max=2310720
EOF
    modprobe br_netfilter && sysctl -p /etc/sysctl.d/k8s.conf
    systemctl enable chronyd && systemctl start chronyd && chronyc sources

    echo -e "\033[32;32m 开始配置系统ipvs \033[0m \n"

    cat > /etc/sysconfig/modules/ipvs.modules <<EOF
#!/bin/bash
modprobe -- ip_vs
modprobe -- ip_vs_rr
modprobe -- ip_vs_wrr
modprobe -- ip_vs_sh
modprobe -- nf_conntrack
modprobe -- ip_tables
modprobe -- ip_set
modprobe -- xt_set
modprobe -- ipt_set
modprobe -- ipt_rpfilter
modprobe -- ipt_REJECT
modprobe -- ipip
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
    yum install -y docker-ce-{{ .DockerVersion }} docker-ce-cli-{{ .DockerVersion }}

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
  "registry-mirrors": [
    "https://mirror.ccs.tencentyun.com",
    "https://4xr1qpsp.mirror.aliyuncs.com"
  ],
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

function Install_kubernetes_component(){
    rpm -qa | grep kubelet && echo -e "\033[32;32m 已安装kubernetes组件 \033[0m \n" && return
    
    echo -e "\033[32;32m 开始安装k8s组件 \033[0m \n"
    cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64/
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF
    yum install -y --nogpgcheck kubelet-{{ .K8sVersion }} kubeadm-{{ .K8sVersion }} kubectl-{{ .K8sVersion }} kubernetes-cni
    echo "source <(kubectl completion bash)" >> ~/.bashrc
}

function Update_kernel(){
    uname -r | grep 5.7 &> /dev/null && echo -e "\033[32;32m 已完成内核升级 \033[0m \n" && return 

    echo -e "\033[32;32m 升级Centos7系统内核到5版本，解决Docker-ce版本兼容问题\033[0m \n"
    rpm --import https://www.elrepo.org/RPM-GPG-KEY-elrepo.org && \
    rpm -Uvh http://www.elrepo.org/elrepo-release-7.0-3.el7.elrepo.noarch.rpm 
    yum --disablerepo=\* --enablerepo=elrepo-kernel repolist && \
    yum --disablerepo=\* --enablerepo=elrepo-kernel install -y kernel-ml.x86_64 && \
    yum remove -y kernel-tools-libs.x86_64 kernel-tools.x86_64 && \
    yum --disablerepo=\* --enablerepo=elrepo-kernel install -y kernel-ml-tools.x86_64 && \
    grub2-set-default 0

#    wget https://cbs.centos.org/kojifiles/packages/kernel/4.9.221/37.el7/x86_64/kernel-4.9.221-37.el7.x86_64.rpm
#    rpm -ivh kernel-4.9.221-37.el7.x86_64.rpm
}

# 初始化顺序
echo -e "\033[32;32m 开始初始化结点 @{{ .HostIP }}@ \033[0m \n" 
Firewalld_process && \
Install_depend_environment && \
Install_docker && \
Install_kubernetes_component  && \
Update_kernel
`

	haproxyCfg = `
#---------------------------------------------------------------------
# Global settings
#---------------------------------------------------------------------
global
    # to have these messages end up in /var/log/haproxy.log you will
    # need to:
    #
    # 1) configure syslog to accept network log events.  This is done
    #    by adding the '-r' option to the SYSLOGD_OPTIONS in
    #    /etc/sysconfig/syslog
    #
    # 2) configure local2 events to go to the /var/log/haproxy.log
    #   file. A line like the following can be added to
    #   /etc/sysconfig/syslog
    #
    #    local2.*                       /var/log/haproxy.log
    #
    log         127.0.0.1 local2

    chroot      /var/lib/haproxy
    pidfile     /var/run/haproxy.pid
    maxconn     4000
    user        haproxy
    group       haproxy
    daemon

    # turn on stats unix socket
    stats socket /var/lib/haproxy/stats

#---------------------------------------------------------------------
# common defaults that all the 'listen' and 'backend' sections will
# use if not designated in their block
#---------------------------------------------------------------------
defaults
    mode                    http
    log                     global
    option                  httplog
    option                  dontlognull
    option http-server-close
    option                  redispatch
    retries                 3
    timeout http-request    10s
    timeout queue           1m
    timeout connect         10s
    timeout client          1m
    timeout server          1m
    timeout http-keep-alive 10s
    timeout check           10s
    maxconn                 3000

#---------------------------------------------------------------------
# kubernetes apiserver frontend which proxys to the backends
#---------------------------------------------------------------------
frontend kubernetes
    mode                 tcp
    bind                 *:{{ default "8443" .BindPort }}
    option               tcplog
    default_backend      kubernetes-apiserver

#---------------------------------------------------------------------
# round robin balancing between the various backends
#---------------------------------------------------------------------
backend kubernetes-apiserver
    mode        tcp
    balance     roundrobin
{{range $nodeName, $endpoint := .EndpointMap }}
    server  .nodeName .endpoint check
{{end}}
#    server  k8s-master01 10.211.55.3:6443 check
#    server  k8s-master02 10.211.55.5:6443 check
#    server  k8s-master03 10.211.55.6:6443 check

#---------------------------------------------------------------------
# collection haproxy statistics message
#---------------------------------------------------------------------
listen stats
    bind                 *:9999
    stats auth           admin:P@ssW0rd
    stats refresh        5s
    stats realm          HAProxy\ Statistics
    stats uri            /admin?stats
`

	keepalivedConf = `
! Configuration File for keepalived

global_defs {
   notification_email {
     acassen@firewall.loc
     failover@firewall.loc
     sysadmin@firewall.loc
   }
   notification_email_from Alexandre.Cassen@firewall.loc
   smtp_server {{ default "192.168.200.1" .SmtpServer }}
   smtp_connect_timeout 30
   router_id LVS_DEVEL
   vrrp_skip_check_adv_addr
   vrrp_garp_interval 0
   vrrp_gna_interval 0
}
# 定义脚本
vrrp_script check_apiserver {
    script "/etc/keepalived/check_apiserver.sh"
    interval 2
    weight -5
    fall 3
    rise 2
}

vrrp_instance VI_1 {
    state MASTER
    interface eth0
    virtual_router_id 51
    priority 100
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass 1111
    }
    virtual_ipaddress {
      {{ .ApiServerVip}}
    }

    # 调用脚本
    #track_script {
    #    check_apiserver
    #}
}
`

	checkApiserver = `
#!/bin/bash

function check_apiserver(){
 for ((i=0;i<5;i++))
 do
  apiserver_job_id=${pgrep kube-apiserver}
  if [[ ! -z ${apiserver_job_id} ]];then
   return
  else
   sleep 2
  fi
  apiserver_job_id=0
 done
}

# 1->running    0->stopped
check_apiserver
if [[ $apiserver_job_id -eq 0 ]];then
 /usr/bin/systemctl stop keepalived
 exit 1
else
 exit 0
fi
`
)
