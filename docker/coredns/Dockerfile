ARG CENTOS_BASE_VERSION=8.3
ARG COREDNS_BASE_VERSION=1.7.0

FROM coredns/coredns:${COREDNS_BASE_VERSION} as default

FROM wtxue/centos-base:${CENTOS_BASE_VERSION}

COPY --from=default /etc/ssl/certs /etc/ssl/certs
COPY --from=default /coredns /coredns

EXPOSE 53 53/udp
ENTRYPOINT ["/coredns"]


