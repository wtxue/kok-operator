FROM centos:8.3.2011

RUN yum update -y && yum install -y --nogpgcheck \
    iptables \
    lsof \
    curl \
    net-tools \
    tree \
    tcpdump \
    ipvsadm \
    ipset \
    telnet \
    strace \
    file \
    sysstat \
    iotop \
    wget \
    which \
    nc  \
    perf \
    iputils \
    tcptraceroute \
    && yum clean all \
    && rm -rf /var/cache/yum/*
