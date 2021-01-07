# kube-prometheus

~~~ shell
# add helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add stable https://charts.helm.sh/stable
helm repo update

# custom helm install override values
cat > ./overridevalues-kube-prometheus.yaml <<EOF
kubeEtcd:
  enabled: true
  endpoints:
  - 10.227.81.203

grafana:
  adminPassword: "123456"
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: contour
    hosts:
    - grafana.k8s.io
  persistence:
    enabled: true
    type: statefulset
    size: 1Gi
    

prometheus:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: contour
    hosts:
    - prometheus.k8s.io
  prometheusSpec:
    storageSpec: 
      volumeClaimTemplate:
        spec:
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 50Gi

alertmanager:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: contour
    hosts:
    - alertmanager.k8s.io
  alertmanagerSpec:
    storage: 
      volumeClaimTemplate:
        spec:
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 1Gi
EOF

# helm3 
helm upgrade --install --debug --namespace observe-system --create-namespace  \
  prometheus  prometheus-community/kube-prometheus-stack -f ./values-k3s-kube-prometheus.yaml
~~~

# loki

~~~ shell

helm repo add grafana https://grafana.github.io/helm-charts
helm repo update


# custom helm install override values
cat > ./overridevalues-loki.yaml <<EOF
fluent-bit:
  enabled: false

loki:
  persistence:
    enabled: true
    size: 50Gi
  serviceMonitor:
    enabled: true
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: contour
      projectcontour.io/websocket-routes: "/*"
    hosts:
      - host: loki.k8s.io
        paths:
          - "/*"

promtail:
  enabled: true          
  extraScrapeConfigs:
  - job_name: journal
    journal:
      path: /var/log/journal
      max_age: 12h
      labels:
        job: systemd-journal
    relabel_configs:
      - source_labels: ['__journal__systemd_unit']
        target_label: 'unit'
      - source_labels: ['__journal__hostname']
        target_label: 'hostname'
  extraVolumes:
  - name: journal
    hostPath:
      path: /var/log/journal
  extraVolumeMounts:
  - name: journal
    mountPath: /var/log/journal
    readOnly: true
EOF

helm upgrade --install --debug --namespace observe-system --create-namespace \
    loki grafana/loki-stack -f ./overridevalues-loki.yaml

~~~
