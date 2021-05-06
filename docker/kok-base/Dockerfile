ARG BASE_VERSION=stable-20210408

ARG K8S_ETCD_VERSION2=v3.4.15

FROM quay.io/coreos/etcd:${K8S_ETCD_VERSION2} as etcd-v2
FROM docker.io/wtxue/debian-base:${BASE_VERSION} as default

ENV K9S_SERVER_VERSION v0.24.9
RUN curl -fsSL https://github.com/derailed/k9s/releases/download/$K9S_SERVER_VERSION/k9s_Linux_x86_64.tar.gz -o k9s.tar.gz  \
    && tar -C /tmp/ -xzf k9s.tar.gz \
    && mkdir -p /k8s/bin/ &&  mv /tmp/k9s /k8s/bin/

ENV CNI_PLUGINS_VERSION v0.9.1
RUN curl -fsSL https://github.com/containernetworking/plugins/releases/download/$CNI_PLUGINS_VERSION/cni-plugins-linux-amd64-$CNI_PLUGINS_VERSION.tgz -o cni.tgz \
    && mkdir -p /k8s/bin/ && mv cni.tgz /k8s/bin/

ENV Containerd_VERSION 1.5.0
RUN curl -fsSL https://github.com/containerd/containerd/releases/download/v${Containerd_VERSION}/cri-containerd-cni-${Containerd_VERSION}-linux-amd64.tar.gz -o containerd.tar.gz \
    && mkdir -p /k8s/bin/ && mv containerd.tar.gz /k8s/bin/

ENV HELM_VERSION v3.5.4
RUN curl -fsSL https://get.helm.sh/helm-$HELM_VERSION-linux-amd64.tar.gz -o helm.tar.gz \
    && tar -C /tmp/ -xzf helm.tar.gz \
    && mkdir -p /k8s/bin/ && mv /tmp/linux-amd64/helm /k8s/bin/

ENV K8S_V2 v1.20.4
RUN curl -fsSL https://dl.k8s.io/${K8S_V2}/kubernetes-server-linux-amd64.tar.gz -o k8s-${K8S_V2}.tar.gz
RUN mkdir -p /k8s-${K8S_V2}/bin/ && tar -C /k8s-${K8S_V2} -xf k8s-${K8S_V2}.tar.gz \
    && mv /k8s-${K8S_V2}/kubernetes/server/bin/kube-apiserver /k8s-${K8S_V2}/bin/ \
    && mv /k8s-${K8S_V2}/kubernetes/server/bin/kubeadm /k8s-${K8S_V2}/bin/ \
    && mv /k8s-${K8S_V2}/kubernetes/server/bin/kubectl /k8s-${K8S_V2}/bin/ \
    && mv /k8s-${K8S_V2}/kubernetes/server/bin/kubelet /k8s-${K8S_V2}/bin/ \
    && mv /k8s-${K8S_V2}/kubernetes/server/bin/kube-scheduler /k8s-${K8S_V2}/bin/ \
    && mv /k8s-${K8S_V2}/kubernetes/server/bin/kube-controller-manager /k8s-${K8S_V2}/bin/


COPY --from=etcd-v2 /usr/local/bin/etcd \
                          /usr/local/bin/etcdctl \
                          /k8s-$K8S_V2/bin/

FROM docker.io/wtxue/debian-base:${BASE_VERSION}

ENV K8S_V2 v1.20.4
COPY --from=default /k8s-${K8S_V2}/bin /k8s-${K8S_V2}/bin

COPY --from=default /k8s/bin /k8s/bin
