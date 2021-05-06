FROM debian:stable-20210408

RUN apt-get update && \
    apt-get install -y ca-certificates curl git openssl default-mysql-client unzip && \
    apt-get install -y net-tools tcpdump ipvsadm telnet iotop wget iptables && \
    apt-get autoremove -y && \
    apt-get clean -y
