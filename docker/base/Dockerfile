FROM debian:stable-20220822

RUN apt-get update && \
    apt-get install -y procps dnsutils ca-certificates curl git openssl unzip && \
    apt-get install -y net-tools tcpdump ipvsadm telnet iotop wget iptables dnsutils && \
    apt-get autoremove -y && \
    apt-get clean -y

RUN curl -fsSL https://hey-release.s3.us-east-2.amazonaws.com/hey_linux_amd64 -o hey  \
    && chmod a+x hey &&  mv hey /usr/local/bin/
