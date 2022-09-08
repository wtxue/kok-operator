#!/usr/bin/env bash

set -xeuo pipefail

K8s_Version=${1:-"v1.24.4"}
Etcd_Version=${2:-"3.5.3-0"}
CoreDns_Version=${3:-"v1.8.6"}
DstImagePrefix=${4:-"docker.io/wtxue"}
Pause_Version=${5:-"3.7"}

docker pull k8s.gcr.io/kube-apiserver:${K8s_Version} && \
    docker tag k8s.gcr.io/kube-apiserver:${K8s_Version} ${DstImagePrefix}/kube-apiserver:${K8s_Version} && \
    docker push ${DstImagePrefix}/kube-apiserver:${K8s_Version}

docker pull k8s.gcr.io/kube-controller-manager:${K8s_Version} && \
    docker tag k8s.gcr.io/kube-controller-manager:${K8s_Version} ${DstImagePrefix}/kube-controller-manager:${K8s_Version} && \
    docker push ${DstImagePrefix}/kube-controller-manager:${K8s_Version}

docker pull k8s.gcr.io/kube-scheduler:${K8s_Version} && \
    docker tag k8s.gcr.io/kube-scheduler:${K8s_Version} ${DstImagePrefix}/kube-scheduler:${K8s_Version} && \
    docker push ${DstImagePrefix}/kube-scheduler:${K8s_Version}

docker pull k8s.gcr.io/kube-proxy:${K8s_Version} && \
    docker tag k8s.gcr.io/kube-proxy:${K8s_Version} ${DstImagePrefix}/kube-proxy:${K8s_Version} && \
    docker push ${DstImagePrefix}/kube-proxy:${K8s_Version}

docker pull k8s.gcr.io/pause:${Pause_Version} && \
    docker tag k8s.gcr.io/pause:${Pause_Version} ${DstImagePrefix}/pause:${Pause_Version} && \
    docker push ${DstImagePrefix}/pause:${Pause_Version}

docker pull k8s.gcr.io/etcd:${Etcd_Version} && \
    docker tag k8s.gcr.io/etcd:${Etcd_Version} ${DstImagePrefix}/etcd:${Etcd_Version} && \
    docker push ${DstImagePrefix}/etcd:${Etcd_Version}

docker pull k8s.gcr.io/coredns/coredns:${CoreDns_Version} && \
    docker tag k8s.gcr.io/coredns/coredns:${CoreDns_Version} ${DstImagePrefix}/coredns:${CoreDns_Version} && \
    docker push ${DstImagePrefix}/coredns:${CoreDns_Version}


