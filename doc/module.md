# containerd

~~~shell
sudo tar -C / -xzf cri-containerd-cni-1.4.3-linux-amd64.tar.gz
systemctl enable containerd && systemctl daemon-reload && systemctl restart containerd
ctr version


mkdir /etc/containerd
containerd config default > /etc/containerd/config.toml


docker.io/rancher/pause:3.2
k8s.gcr.io/pause:3.2


# mirrors and config
[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    sandbox_image = "docker.io/rancher/pause:3.2"
    [plugins."io.containerd.grpc.v1.cri".registry]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
          endpoint = [
          	"https://yqdzw3p0.mirror.aliyuncs.com"
          ]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."k8s.gcr.io"]
          endpoint = [
          	"https://registry.aliyuncs.com/k8sxio"
          ]     
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."quay.io"]
          endpoint = [
          	"https://quay.mirrors.ustc.edu.cn"
          ]        
          
    [plugins."io.containerd.grpc.v1.cri".registry]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
        ...
      [plugins."io.containerd.grpc.v1.cri".registry.configs]
        [plugins."io.containerd.grpc.v1.cri".registry.configs."harbor.example.net".tls]
          insecure_skip_verify = true
        [plugins."io.containerd.grpc.v1.cri".registry.configs."harbor.example.net".auth]
          username = "admin"
          password = "Mydlq123456"       
          
~~~

# Containerd + Docker

事实上，Docker 和 Containerd 是可以同时使用的，只不过 Docker 默认使用的 Containerd 的命名空间不是 default，而是 `moby`。 首先从其他装了 Docker 的机器或者 GitHub 上下载 Docker
相关的二进制文件，然后使用下面的命令启动 Docker：

~~~shell
dockerd --containerd /run/containerd/containerd.sock --cri-containerd
ctr ns ls
ctr -n moby c ls
~~~
