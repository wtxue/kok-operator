http://mirrors.163.com/debian-cd/10.7.0/amd64/iso-dvd/debian-10.7.0-amd64-DVD-1.iso

# kubeadm config images list

k8s.gcr.io/kube-apiserver:v1.19.6 k8s.gcr.io/kube-controller-manager:v1.19.6 k8s.gcr.io/kube-scheduler:v1.19.6 k8s.gcr.io/kube-proxy:v1.19.6
k8s.gcr.io/pause:3.2 k8s.gcr.io/etcd:3.4.13-0 k8s.gcr.io/coredns:1.7.0


docker run --rm -it registry.cn-hangzhou.aliyuncs.com/wtxue/kok-base:v0.2.0-20031816 bash


tar -czvf k8s.tar.gz k8s k8s-v1.18.16 k8s-v1.20.3

docker cp 45efd2f08683:/k8s.tar.gz bin/linux/
