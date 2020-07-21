#!/bin/bash

set -e  # exit immediately on error
set -x  # display all commands

PACKAGE_DIR="../k8s"

sysOS=`uname -s`
TargetOS="darwin"
Version="2.3.1"
if [ $sysOS == "Darwin" ];then
	TargetOS="darwin"
else
	TargetOS="linux"
fi

if [ ! -f ${PACKAGE_DIR}/bin/kube-apiserver ]; then
	if [ ! -f kubebuilder_${Version}_${TargetOS}_amd64.tar.gz ]; then
		wget https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${Version}/kubebuilder_${Version}_${TargetOS}_amd64.tar.gz
	fi

	tar -xf kubebuilder_${Version}_${TargetOS}_amd64.tar.gz
	mkdir -p PACKAGE_BIN
    cp -rf kubebuilder_${Version}_${TargetOS}_amd64/bin ${PACKAGE_DIR}/
    rm -rf kubebuilder_${Version}_${TargetOS}_amd64.tar.gz kubebuilder_${Version}_${TargetOS}_amd64
fi


echo "all done."