ARG BASE_VERSION=stable-20220822

ARG K8S_ETCD_VERSION2=3.5.3-0

FROM registry.aliyuncs.com/google_containers/etcd:${K8S_ETCD_VERSION2} as etcd-v2
FROM docker.io/wtxue/debian-base:${BASE_VERSION} as default

ENV K9S_SERVER_VERSION v0.26.3
RUN curl -fsSL https://github.com/derailed/k9s/releases/download/$K9S_SERVER_VERSION/k9s_Linux_x86_64.tar.gz -o k9s.tar.gz  \
    && tar -C /tmp/ -xzf k9s.tar.gz \
    && mkdir -p /k8s/bin/ &&  mv /tmp/k9s /k8s/bin/

ENV CNI_PLUGINS_VERSION v1.1.1
RUN curl -fsSL https://github.com/containernetworking/plugins/releases/download/$CNI_PLUGINS_VERSION/cni-plugins-linux-amd64-$CNI_PLUGINS_VERSION.tgz -o cni.tgz \
    && mkdir -p /k8s/bin/ && mv cni.tgz /k8s/bin/


ENV CRICTL_VERSION="v1.25.0"
RUN curl -fsSL https://github.com/kubernetes-sigs/cri-tools/releases/download/$CRICTL_VERSION/crictl-${CRICTL_VERSION}-linux-amd64.tar.gz -o crictl-linux-amd64.tar.gz \
    && tar zxvf crictl-linux-amd64.tar.gz -C /k8s/bin/

ENV RUNC_VERSION="v1.1.4"
RUN curl -fsSL https://github.com/opencontainers/runc/releases/download/$RUNC_VERSION/runc.amd64 -o runc \
    && mv runc /k8s/bin/

ENV Containerd_VERSION 1.6.8
RUN curl -fsSL https://github.com/containerd/containerd/releases/download/v${Containerd_VERSION}/cri-containerd-cni-${Containerd_VERSION}-linux-amd64.tar.gz -o containerd.tar.gz \
    && mkdir -p /k8s/bin/ && mv containerd.tar.gz /k8s/bin/

ENV HELM_VERSION v3.9.4
RUN curl -fsSL https://get.helm.sh/helm-$HELM_VERSION-linux-amd64.tar.gz -o helm.tar.gz \
    && tar -C /tmp/ -xzf helm.tar.gz \
    && mkdir -p /k8s/bin/ && mv /tmp/linux-amd64/helm /k8s/bin/

ENV K8S_V2 v1.24.4
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

ENV K8S_V2 v1.24.4
COPY --from=default /k8s-${K8S_V2}/bin /k8s-${K8S_V2}/bin

COPY --from=default /k8s/bin /k8s/bin

