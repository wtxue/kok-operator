FROM centos:7.8.2003

RUN yum update -y && yum install -y --nogpgcheck \
    deltarpm  \
    device-mapper \
    bind-utils  \
    bridge-utils \
    iptables \
    lsof \
    curl \
    net-tools \
    tree \
    tcpdump \
    ipvsadm \
    vim \
    netstat \
    ipset \
    telnet \
    iproute \
    strace \
    file \
    sysstat \
    iotop \
    wget \
    which \
    tcptraceroute \
    && yum clean all \
    && rm -rf /var/cache/yum/*