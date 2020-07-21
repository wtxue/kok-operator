package etcd

const (
	etcdTemplate = `
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: etcd
  namespace: {{ .Namespace }}
  labels:
    app: etcd
spec:
  serviceName: etcd
  replicas: 3
  selector:
    matchLabels:
      app: etcd
  template:
    metadata:
      name: etcd
      labels:
        app: etcd
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app: etcd
              namespaces:
              - {{ .Namespace }}
              topologyKey: kubernetes.io/hostname
            weight: 1
      containers:
        - name: etcd
          image: {{ default "registry.aliyuncs.com/google_containers/etcd:3.4.3-0" .ImageName }}
          ports:
            - containerPort: 2379
              name: client
            - containerPort: 2380
              name: peer
          env:
            - name: INITIAL_CLUSTER_SIZE
              value: "3"
            - name: CLUSTER_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          resources:
            requests:
              cpu: 100m
              memory: 512Mi
          volumeMounts:
            - name: datadir
              mountPath: /var/run/etcd
          command:
            - /bin/sh
            - -c
            - |
              PEERS="etcd-0=http://etcd-0.etcd:2380,etcd-1=http://etcd-1.etcd:2380,etcd-2=http://etcd-2.etcd:2380"
              exec etcd --name ${HOSTNAME} \
                --listen-peer-urls http://0.0.0.0:2380 \
                --listen-client-urls http://0.0.0.0:2379 \
                --advertise-client-urls http://${HOSTNAME}.etcd:2379 \
                --initial-advertise-peer-urls http://${HOSTNAME}:2380 \
                --initial-cluster-token etcd-cluster-1 \
                --initial-cluster ${PEERS} \
                --initial-cluster-state new \
                --data-dir /var/run/etcd/default.etcd
  volumeClaimTemplates:
    - metadata:
        name: datadir
      spec:
        storageClassName: local-storage
        accessModes:
          - "ReadWriteOnce"
        resources:
          requests:
            storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: etcd
  namespace: {{ .Namespace }}
  labels:
    app: etcd
spec:
  ports:
    - port: 2380
      name: etcd-server
    - port: 2379
      name: etcd-client
  clusterIP: None
  selector:
    app: etcd
  publishNotReadyAddresses: true
---
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: etcd-pdb
  namespace: {{ .Namespace }}
  labels:
    pdb: etcd
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: etcd
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  labels:
    createdBy: controller
  name: local-storage
provisioner: kubernetes.io/no-provisioner
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: {{ .ClusterName }}-etcd0-lpv
spec:
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: 10Gi
  local:
    path: {{ default "/web/default-cluster" .EtcdHostPath }}
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - {{ .EtcdHostName0 }}
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  volumeMode: Filesystem
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: {{ .ClusterName }}-etcd1-lpv
spec:
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: 10Gi
  local:
    path: {{ default "/web/default-cluster" .EtcdHostPath }}
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - {{ .EtcdHostName1 }}
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  volumeMode: Filesystem
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: {{ .ClusterName }}-etcd2-lpv
spec:
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: 10Gi
  local:
    path: {{ default "/web/default-cluster" .EtcdHostPath }}
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - {{ .EtcdHostName2 }}
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  volumeMode: Filesystem`
)
