# kok-operator

kok-operator 是一个自动化部署高可用kubernetes的operator

# 特性

- 启用 k3s，解决裸金属第一次部署集群没有元集群问题
- 云原生架构，crd+controller，采用声明式 api 描述一个集群的生命周期(创建，升级，删除)
- 支持 裸金属模式 和 托管模式 两种方式部署集群
- kubelet 自动生成证书，无坑版100年集群证书
- 除 kubelet 外集群组件全部容器化部署，采用 static pod 方式部署高可用 etcd 集群
- 支持 coredns, flannel，metrics-server，kube-proxy, metall b等 addons 模板化部署
- 支持 centos 和 debian 系统
- 支持结点下线清理
- 支持 helm 部署

# 安装部署

## 准备

下载启动 k3s 集群
```bash
# 下载二进制文件, 进入tools目录
$ cd tools
$ bash https://raw.githubusercontent.com/wtxue/kok-operator/master/tools/centos-k3s-node.sh 

# 等待 k3s 运行正常后，查看 k3s admin kubeconfig
$ cat /etc/rancher/k3s/k3s.yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJXRENCL3FBREFnRUNBZ0VBTUFvR0NDcUdTTTQ5QkFNQ01DTXhJVEFmQmdOVkJBTU1HR3N6Y3kxelpYSjIKWlhJdFkyRkFNVFl3TkRNNU56STVNakFlRncweU1ERXhNRE13T1RVME5USmFGdzB6TURFeE1ERXdPVFUwTlRKYQpNQ014SVRBZkJnTlZCQU1NR0dzemN5MXpaWEoyWlhJdFkyRkFNVFl3TkRNNU56STVNakJaTUJNR0J5cUdTTTQ5CkFnRUdDQ3FHU000OUF3RUhBMElBQkE5WGZEVTRkcmZPTnplSWlKMDV4WUNTWjA4REJYN2ZoMURaZzFNdUQ4VmYKWVQwS2R5SCtyRzZQVi9xdExMbHFocGM2Rkp1MlZiR3VsbFZ6T0hIa2VDaWpJekFoTUE0R0ExVWREd0VCL3dRRQpBd0lDcERBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUFvR0NDcUdTTTQ5QkFNQ0Ewa0FNRVlDSVFEckZvK1pyNWY3CllHTUdQSnVhQ3dQdmZlNURqZGRXNm52R2pWNVRKM1IwclFJaEFQcXVWelJUMUczMEgrYmdDS3NobWhUQXZxMWwKUU9sakVUdjlkYnFGMTNSUAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://127.0.0.1:6443   # 注意外部访问，修改 127.0.0.1 为 k3s 结点 IP
  name: default
contexts:
- context:
    cluster: default
    user: default
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: default
  user:
    password: 6253ebe7e75ce5afe7baaad49f99371c
    username: admin
```

## 运行

本地运行
```bash
# apply crd
$ kubectl apply -f manifests/crds/
customresourcedefinition.apiextensions.k8s.io/clustercredentials.devops.k8s.io created
customresourcedefinition.apiextensions.k8s.io/clusters.devops.k8s.io created
customresourcedefinition.apiextensions.k8s.io/machines.devops.k8s.io created

# 指定 kubeconfig 运行
$ go run cmd/admin-controller/main.go ctrl -v 4 --kubeconfig={}/k3s-kubeconfig.yaml
```
k3s 安装运行
```bash
helm upgrade kok-operator --create-namespace --namespace kok-system --debug ./charts/kok-operator

kubectl get pod -n kok-system      
NAME                            READY   STATUS    RESTARTS   AGE
kok-operator-6ff65bc44b-hg4nh   1/1     Running   0          31m

```

## 创建集群
### 创建裸金属集群
```bash
# 创建集群cr
$ kubectl apply -f ./manifests/example-cluster.yaml

# 创建集群结点
$ kubectl apply -f ./manifests/example-cluster-node.yaml
```

### 创建托管集群
创建托管集群时，kok-operator 需要运行在 meta 高可用集群上，这里使用集群名为 meta-cluster, 注意一个 namespace 一个托管集群
```bash
# 创建 etcd 集群
$ kubectl apply -f ./manifests/etcd-statefulset.yaml

# 创建托管集群cr
kubectl apply -f ./manifests/hosted-cluster.yaml

# 创建托管集群结点
kubectl apply -f ./manifests/hosted-cluster-node.yaml
```

# 计划

- [x]  打通元集群及托管集群 service 网络，以支持聚合 apiserver
- [x]  支持 helm v3 部署 addons
