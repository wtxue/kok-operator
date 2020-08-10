ARG CENTOS_BASE_VERSION=v7.8.2
ARG ETCD_BASE_VERSION=v3.4.9

FROM quay.io/coreos/etcd:${ETCD_BASE_VERSION} as etcd-default
FROM registry.cn-hangzhou.aliyuncs.com/wtxue/centos-base:${CENTOS_BASE_VERSION} as default
ENV K8S_SERVER_VERSION v1.18.5
RUN curl -fsSL https://dl.k8s.io/$K8S_SERVER_VERSION/kubernetes-server-linux-amd64.tar.gz  -o kubernetes-server-linux-amd64.tar.gz \
    && tar -C /usr/local -xzf kubernetes-server-linux-amd64.tar.gz \
    && rm kubernetes-server-linux-amd64.tar.gz

FROM registry.cn-hangzhou.aliyuncs.com/wtxue/centos-base:${CENTOS_BASE_VERSION}

COPY --from=etcd-default /usr/local/bin/etcd \
                          /usr/local/bin/etcdctl \
                          /usr/local/bin/

COPY --from=default /usr/local/kubernetes/server/bin/kube-apiserver \
                    /usr/local/kubernetes/server/bin/kubeadm  \
                    /usr/local/kubernetes/server/bin/kubectl \
                    /usr/local/kubernetes/server/bin/kubelet \
                    /usr/local/kubernetes/server/bin/kube-scheduler \
                    /usr/local/kubernetes/server/bin/kube-controller-manager \
                    /usr/local/bin/
