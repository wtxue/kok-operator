# kube-on-kube-operator

kube-on-kube-operator 是一个自动化部署高可用kubernetes的operator

# 特性

- 云原生架构，crd+controller，采用声明式api描述一个集群的最终状态
- 可以启用一个fake-cluster，解决裸金属第一次部署集群没有元集群问题
- 无坑版100年集群证书
- 除kubelet外集群组件全部容器化部署，componentstatuses可以发现三个etcd
- 支持flannel，metrics-server等直接一键部署

# 安装部署

## 准备

下载fake-cluster需要二进制文件，启动fake-cluster

```bash
# 下载二进制文件, 进入tools目录
$ cd tools
$ ./init.sh

# 进入项目根目录  运行 fake apiserver
$ cd ..
$ go run cmd/admin-controller/main.go fake -v 4 

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

```bash
# apply crd
$ export KUBECONFIG=k8s/cfg/fake-kubeconfig.yaml && kubectl apply -f manifests/crds/
customresourcedefinition.apiextensions.k8s.io/clustercredentials.devops.k8s.io created
customresourcedefinition.apiextensions.k8s.io/clusters.devops.k8s.io created
customresourcedefinition.apiextensions.k8s.io/machines.devops.k8s.io created

# 运行
$ go run cmd/admin-controller/main.go ctrl -v 4 --kubeconfig=k8s/cfg/fake-kubeconfig.yaml
```

# 计划

- [x]  master组件托管