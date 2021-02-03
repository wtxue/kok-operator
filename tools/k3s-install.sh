#!/usr/bin/env bash

set -xeuo pipefail

AdvertiseIP=${1:-"10.40.0.28"}
ContainerdVersion=${2:-"1.4.3"}
EtcdVersion=${3:-"v3.4.14"}
K3sVersion=${4:-"v1.20.0"}

function Install_depend_software(){
    echo -e "\033[32;32m 开始安装依赖环境包 \033[0m \n"
    apt-get update && apt-get upgrade
    apt-get install -y sudo
    sudo apt-get install -y curl wget vim telnet ipvsadm tree telnet wget net-tools  \
           bash-completion sysstat chrony jq sysstat socat conntrack lsof libseccomp2 util-linux apt-transport-https
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

function Install_containerd_service() {
  if [ ! -f /usr/local/bin/containerd ]; then
    if [ ! -f cri-containerd-cni-${ContainerdVersion}-linux-amd64.tar.gz ]; then
#      wget https://github.com/containerd/containerd/releases/download/v${ContainerdVersion}/cri-containerd-cni-${ContainerdVersion}-linux-amd64.tar.gz
      wget https://download.fastgit.org/containerd/containerd/releases/download/v${ContainerdVersion}/cri-containerd-cni-${ContainerdVersion}-linux-amd64.tar.gz
	  fi

#    sudo tar -C / -xzf cri-containerd-cni-${ContainerdVersion}-linux-amd64.tar.gz
    tar tf cri-containerd-cni-${ContainerdVersion}-linux-amd64.tar.gz \
      usr/local/bin/containerd   \
      usr/local/bin/containerd-shim  \
      usr/local/bin/containerd-shim-runc-v1  \
      usr/local/bin/containerd-shim-runc-v2  \
      usr/local/bin/crictl     \
      usr/local/bin/ctr   \
      usr/local/sbin/runc  \
      etc/systemd/system/containerd.service  \
      etc/crictl.yaml

    mkdir /etc/containerd
    containerd config default > /etc/containerd/config.toml
    systemctl enable containerd && systemctl daemon-reload && systemctl restart containerd
  fi
}

function Install_etcd_service() {
  if [ ! -f /usr/local/bin/etcd ]; then
    if [ ! -f etcd-${EtcdVersion}-linux-amd64.tar.gz ]; then
		  wget https://github.com/etcd-io/etcd/releases/download/${EtcdVersion}/etcd-${EtcdVersion}-linux-amd64.tar.gz
	  fi

    tar -xf etcd-${EtcdVersion}-linux-amd64.tar.gz
    cp -f etcd-${EtcdVersion}-linux-amd64/etcd* /usr/local/bin/
    rm -rf etcd-${EtcdVersion}-linux-amd64
  fi

  echo -e "\033[32;32m 开始写 /lib/systemd/system/etcd.service \033[0m \n"

  # ExecStart=/usr/local/bin/etcd --listen-client-urls http://0.0.0.0:2379 --listen-peer-urls http://0.0.0.0:2380 --advertise-client-urls http://xxxxx:2379 --data-dir=/var/lib/etcd/ --logger=zap
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
ExecStart=/usr/local/bin/etcd --name etcd1 --listen-client-urls http://0.0.0.0:2379 --listen-peer-urls http://0.0.0.0:2380 --advertise-client-urls http://${AdvertiseIP}:2379 --initial-cluster etcd1=http://${AdvertiseIP}:2380  --initial-cluster-state new --data-dir=/var/lib/etcd/ --logger=zap
Restart=always
LimitNOFILE=65536
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
    systemctl enable etcd && systemctl daemon-reload && systemctl restart etcd
}

function Install_k3s_service() {
  if [ ! -f /usr/local/bin/k3s ]; then
    curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--tls-san=bootstrop.vks.k8s.io --flannel-backend=none --no-flannel --node-name=${AdvertiseIP} --container-runtime-endpoint=unix:///run/containerd/containerd.sock --default-local-storage-path=/mnt/local-storage --disable=traefik,servicelb --datastore-endpoint=http://localhost:2379' sh -
  fi
}


echo -e "\033[32;32m 开始初始化 k3s 结点 ${AdvertiseIP} \033[0m \n"
Install_depend_software && \
Install_ipvs && \
Install_containerd_service && \
Install_etcd_service && \
Install_k3s_service

