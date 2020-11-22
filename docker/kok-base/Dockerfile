ARG CENTOS_BASE_VERSION=v7.8

ARG K8S_ETCD_VERSION1=v3.4.9
ARG K8S_ETCD_VERSION2=v3.4.13

FROM quay.io/coreos/etcd:${K8S_ETCD_VERSION1} as etcd-v1
FROM quay.io/coreos/etcd:${K8S_ETCD_VERSION2} as etcd-v2
FROM wtxue/centos-base:${CENTOS_BASE_VERSION} as default

ENV K9S_SERVER_VERSION v0.23.5
RUN curl -fsSL https://github.com/derailed/k9s/releases/download/$K9S_SERVER_VERSION/k9s_Linux_x86_64.tar.gz -o k9s.tar.gz \
    && tar -C /tmp/ -xzf k9s.tar.gz \
    && mkdir -p /k8s/bin/ &&  mv /tmp/k9s /k8s/bin/  \
    && rm k9s.tar.gz

ENV CNI_PLUGINS_VERSION v0.8.7
RUN curl -fsSL https://github.com/containernetworking/plugins/releases/download/$CNI_PLUGINS_VERSION/cni-plugins-linux-amd64-$CNI_PLUGINS_VERSION.tgz -o cni.tgz \
    && mkdir -p /k8s/bin/ && mv cni.tgz /k8s/bin/

ENV HELM_VERSION v3.4.0
RUN curl -fsSL https://get.helm.sh/helm-$HELM_VERSION-linux-amd64.tar.gz -o helm.tar.gz \
    && tar -C /tmp/ -xzf helm.tar.gz \
    && mkdir -p /k8s/bin/ && mv /tmp/linux-amd64/helm /k8s/bin/

ENV K8S_V1 v1.18.10
RUN curl -fsSL https://dl.k8s.io/$K8S_V1/kubernetes-server-linux-amd64.tar.gz -o k8s-$K8S_V1.tar.gz \
    && mkdir -p /k8s-$K8S_V1/bin/ && tar -C /k8s-$K8S_V1 -xzf k8s-$K8S_V1.tar.gz \
    && mv /k8s-$K8S_V1/kubernetes/server/bin/{kube-apiserver,kubeadm,kubectl,kubelet,kube-scheduler,kube-controller-manager} /k8s-$K8S_V1/bin/

COPY --from=etcd-v1 /usr/local/bin/etcd \
                          /usr/local/bin/etcdctl \
                          /k8s-$K8S_V1/bin/

ENV K8S_V2 v1.19.3
RUN curl -fsSL https://dl.k8s.io/$K8S_V2/kubernetes-server-linux-amd64.tar.gz -o k8s-$K8S_V2.tar.gz \
    && mkdir -p /k8s-$K8S_V2/bin/ && tar -C /k8s-$K8S_V2 -xzf k8s-$K8S_V2.tar.gz \
    && mv /k8s-$K8S_V2/kubernetes/server/bin/{kube-apiserver,kubeadm,kubectl,kubelet,kube-scheduler,kube-controller-manager} /k8s-$K8S_V2/bin/

COPY --from=etcd-v2 /usr/local/bin/etcd \
                          /usr/local/bin/etcdctl \
                          /k8s-$K8S_V2/bin/

FROM wtxue/centos-base:${CENTOS_BASE_VERSION}

ENV K8S_V1 v1.18.10
COPY --from=default /k8s-$K8S_V1/bin /k8s-$K8S_V1/bin

ENV K8S_V2 v1.19.3
COPY --from=default /k8s-$K8S_V2/bin /k8s-$K8S_V2/bin

COPY --from=default /k8s/bin /k8s/bin
