# kok-operator

kok-operator 是一个自动化部署高可用kubernetes的operator

# 特性

- 云原生架构，crd+controller，采用声明式api描述一个集群的生命周期(创建，升级，删除)
- 支持裸金属和master托管模式两种方式部署集群
- 可以启用fake-cluster或者k3s，解决裸金属第一次部署集群没有元集群问题
- 无坑版100年集群证书，kubelet自动生成证书
- 除kubelet外集群组件全部容器化部署，采用static pod方式部署高可用etcd集群
- 支持coredns, flannel，metrics-server，kube-proxy, metallb等 addons 模板化部署
- 支持 centos 和 debian 系统

# 安装部署

## 准备

下载fake-cluster需要二进制文件，启动fake-cluster

```bash
# 下载二进制文件, 进入tools目录
$ cd tools
$ ./init.sh

# 进入项目根目录  运行 fake apiserver
$ cd ..
$ go run cmd/admin-controller/main.go fake --baseBinDir k8s/bin --rootDir k8s -v 4 

# 运行正常后
$ cat k8s/cfg/fake-kubeconfig.yaml
apiVersion: v1
clusters:
- cluster:
    server: 127.0.0.1:18080
  name: fake-cluster
contexts:
- context:
    cluster: fake-cluster
    user: devops
  name: devops@fake-cluster
current-context: devops@fake-cluster
kind: Config
preferences: {}
users:
- name: devops
  user: {}
```

## 运行

本地运行
```bash
# apply crd
$ export KUBECONFIG=k8s/cfg/fake-kubeconfig.yaml && kubectl apply -f manifests/crds/
customresourcedefinition.apiextensions.k8s.io/clustercredentials.devops.k8s.io created
customresourcedefinition.apiextensions.k8s.io/clusters.devops.k8s.io created
customresourcedefinition.apiextensions.k8s.io/machines.devops.k8s.io created

# 运行
$ go run cmd/admin-controller/main.go ctrl -v 4 --kubeconfig=k8s/cfg/fake-kubeconfig.yaml
```
docker 运行
```bash
$ docker run --name fake-cluster -d --restart=always \
   --net="host" \
   --pid="host" \
   -v /root/wtxue/k8s:/k8s \
   registry.cn-hangzhou.aliyuncs.com/wtxue/onkube-controller:v0.1.1 \
   onkube-controller fake -v 4

$ docker run --name onkube-controller -d --restart=always \
   --net="host" \
   --pid="host" \
   -v /root/wtxue/k8s:/k8s \
   registry.cn-hangzhou.aliyuncs.com/wtxue/onkube-controller:v0.1.1  \
   onkube-controller ctrl -v 4 --kubeconfig=/k8s/cfg/fake-kubeconfig.yaml


```

## 创建集群
### 创建裸金属集群
```bash
# 设置 fake-cluster kubeconfig
$ export KUBECONFIG=/root/wtxue/k8s/cfg/fake-kubeconfig.yaml

# 创建集群cr
$ kubectl apply -f ./manifests/example-cluster.yaml

# 创建集群结点
$ kubectl apply -f ./manifests/example-cluster-node.yaml
```

### 创建托管集群
创建托管集群时，onkube-controller需要运行在真实集群上，这里使用上面创建裸金属集群 example-cluster, 注意一个namespace一个集群
```bash
# 设置 example-cluster kubeconfig
$ export KUBECONFIG=/root/wtxue/k8s/cfg/example-cluster-kubeconfig.yaml

# 这里演示直接本地运行，也可以deployment跑到集群上
$ go run cmd/admin-controller/main.go ctrl -v 4 --kubeconfig=k8s/cfg/fake-kubeconfig.yaml

# 创建 etcd 集群
$ kubectl apply -f ./manifests/etcd-statefulset.yaml

# 创建托管集群cr
kubectl apply -f ./manifests/hosted-cluster.yaml

# 创建托管集群结点
kubectl apply -f ./manifests/hosted-cluster-node.yaml
```

# 计划

- [x]  打通元集群及托管集群service网络，以支持聚合apiserver
- [x]  支持 helm v3 部署 addons
- [x]  用 k3s 替换fake-cluster
