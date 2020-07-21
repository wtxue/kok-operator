#!/usr/bin/env bash
etcd-v3.4.9-darwin-amd64.zip

set -e  # exit immediately on error
set -x  # display all commands

PACKAGE_DIR="../k8s"

sysOS=`uname -s`
TargetOS="darwin"
EtcdVersion="3.4.9"
K8sVersion="1.18.3"
if [ $sysOS == "Darwin" ];then
	TargetOS="darwin"
else
	TargetOS="linux"
fi

if [ ! -f ${PACKAGE_DIR}/bin/etcd ]; then
	if [ ! -f etcd-v${EtcdVersion}-${TargetOS}-amd64.zip ]; then
		wget https://github.com/etcd-io/etcd/releases/download/v${EtcdVersion}/etcd-v${EtcdVersion}-${TargetOS}-amd64.zip
	fi

	tar -xf etcd-v${EtcdVersion}-${TargetOS}-amd64.zip
	mkdir -p ${PACKAGE_DIR}/bin
    cp -f etcd-v${EtcdVersion}-${TargetOS}-amd64/etcd* ${PACKAGE_DIR}/bin/
    rm -rf etcd-v${EtcdVersion}-${TargetOS}-amd64.zip etcd-v${EtcdVersion}-${TargetOS}-amd64
fi

if [ ! -f ${PACKAGE_DIR}/bin/kube-apiserver ]; then
	if [ ! -f kubernetes-server-linux-amd64.tar.gz ]; then
		wget https://dl.k8s.io/v${K8sVersion}/kubernetes-server-linux-amd64.tar.gz
	fi

	tar -xf kubernetes-server-linux-amd64.tar.gz
#	mkdir -p ${PACKAGE_DIR}/bin
#    cp -f etcd-v${EtcdVersion}-${TargetOS}-amd64/etcd* ${PACKAGE_DIR}/bin/
#    rm -rf etcd-v${EtcdVersion}-${TargetOS}-amd64.zip etcd-v${EtcdVersion}-${TargetOS}-amd64
fi

echo "all done."