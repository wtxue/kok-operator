ARG CENTOS_BASE_VERSION=v7.8.1
ARG COREDNS_BASE_VERSION=1.6.7

FROM coredns/coredns:${COREDNS_BASE_VERSION} as default

FROM registry.cn-hangzhou.aliyuncs.com/wtxue/centos-base:${CENTOS_BASE_VERSION}

COPY --from=default /etc/ssl/certs /etc/ssl/certs
COPY --from=default /coredns /coredns

EXPOSE 53 53/udp
ENTRYPOINT ["/coredns"]


