ARG CENTOS_BASE_VERSION=v7.8.10
ARG METRICS_SERVER_BASE_VERSION=v0.3.6

FROM k8s.gcr.io/metrics-server:${METRICS_SERVER_BASE_VERSION} as default

FROM registry.cn-hangzhou.aliyuncs.com/wtxue/centos-base:${CENTOS_BASE_VERSION}

COPY --from=default /metrics-server /

USER 65534

ENTRYPOINT ["/metrics-server"]
