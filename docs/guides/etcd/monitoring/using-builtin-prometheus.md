---
title: Monitor Etcd using Builtin Prometheus Discovery
menu:
  docs_0.8.0:
    identifier: etcd-using-builtin-prometheus-monitoring
    name: Builtin Prometheus Discovery
    parent: etcd-monitoring
    weight: 10
menu_name: docs_0.8.0
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Using Prometheus with KubeDB

This tutorial will show you how to monitor KubeDB databases using [Prometheus](https://prometheus.io/).

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```console
$ kubectl create ns demo
namespace "demo" created

$ kubectl get ns
NAME          STATUS    AGE
default       Active    45m
demo          Active    10s
kube-public   Active    45m
kube-system   Active    45m
```

Note that the yaml files that are used in this tutorial, stored in [docs/examples](https://github.com/kubedb/cli/tree/master/docs/examples) folder in GitHub repository [kubedb/cli](https://github.com/kubedb/cli).

## Create a Etcd database

KubeDB implements a `Etcd` CRD to define the specification of a Etcd database. Below is the `Etcd` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Etcd
metadata:
  name: etcd-mon-prometheus
  namespace: demo
spec:
  replicas: 3
  version: "3.2.13"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/builtin
```

```console
$ kubedb create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/etcd/monitoring/builtin-prometheus/demo-1.yaml
etcd "etcd-mon-prometheus" created
```

Here,

- `spec.version` is the version of Etcd database. In this tutorial, a Etcd 3.2 database is going to be created.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. Since release 0.8.0, a storage spec is required for Etcd.
- `spec.monitor` specifies that built-in [Prometheus](https://github.com/prometheus/prometheus) is used to monitor this database instance. KubeDB operator will configure the service of this database in a way that the Prometheus server will automatically find out the service endpoint aka `Etcd Exporter` and will receive metrics from exporter.

KubeDB operator watches for `Etcd` objects using Kubernetes api. When a `Etcd` object is created, KubeDB operator will create  new pod and a ClusterIP Service with the matching crd name. if one is not already present.

```console
$ kubedb get etcd -n demo
NAME                 STATUS     AGE
etcd-mon-prometheus   Creating   30s


$ kubedb get etcd -n demo
NAME                 STATUS    AGE
etcd-mon-prometheus   Running   10m

$ kubedb describe etcd -n demo etcd-mon-prometheus
Name:		etcd-mon-prometheus
Namespace:	demo
StartTimestamp:	Wed, 01 Aug 2018 17:07:48 +0600
Replicas:	3  total
Status:		Running
Volume:
  StorageClass:	standard
  Capacity:	1Gi
  Access Modes:	RWO

Service:
  Name:		etcd-mon-prometheus
  Type:		ClusterIP
  IP:		None
  Port:		client	2379/TCP
  Port:		peer	2380/TCP

Service:
  Name:		etcd-mon-prometheus-client
  Type:		ClusterIP
  IP:		10.101.234.72
  Port:		client	2379/TCP

Monitoring System:
  Agent:	prometheus.io/builtin
  Prometheus:
    Port:	2379

No Snapshots.

Events:
  FirstSeen   LastSeen   Count     From            Type       Reason             Message
  ---------   --------   -----     ----            --------   ------             -------
  10m         10m        1                         Normal     New Member Added   New member etcd-mon-prometheus-8slp4xxxl8 added to cluster
  11m         11m        1                         Normal     New Member Added   New member etcd-mon-prometheus-7pvzjcd7dx added to cluster
  12m         12m        1                         Normal     New Member Added   New member etcd-mon-prometheus-ld7n576tv5 added to cluster
  12m         12m        1         Etcd operator   Normal     Successful         Successfully created Etcd

```

Since `spec.monitoring` was configured, the database service object is configured accordingly. You can verify it running the following commands:

```console
$ kubectl get services -n demo
NAME                         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
etcd-mon-prometheus          ClusterIP   None            <none>        2379/TCP,2380/TCP   12m
etcd-mon-prometheus-client   ClusterIP   10.101.234.72   <none>        2379/TCP            12
```

```yaml
$ kubectl get services etcd-mon-prometheus -n demo -o yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    monitoring.appscode.com/agent: prometheus.io/builtin
    prometheus.io/path: /metrics
    prometheus.io/port: "2379"
    prometheus.io/scrape: "true"
    service.alpha.kubernetes.io/tolerate-unready-endpoints: "true"
  creationTimestamp: 2018-08-01T11:07:51Z
  labels:
    kubedb.com/kind: Etcd
    kubedb.com/name: etcd-mon-prometheus
  name: etcd-mon-prometheus
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha1
    blockOwnerDeletion: false
    kind: Etcd
    name: etcd-mon-prometheus
    uid: 2006fa2a-957b-11e8-95a5-080027c002b2
  resourceVersion: "24445"
  selfLink: /api/v1/namespaces/demo/services/etcd-mon-prometheus
  uid: 21cd19d9-957b-11e8-95a5-080027c002b2
spec:
  clusterIP: None
  ports:
  - name: client
    port: 2379
    protocol: TCP
    targetPort: 2379
  - name: peer
    port: 2380
    protocol: TCP
    targetPort: 2380
  selector:
    kubedb.com/kind: Etcd
    kubedb.com/name: etcd-mon-prometheus
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
```

We can see that the service contains these specific annotations. The Prometheus server will discover the exporter using these specifications.

```yaml
prometheus.io/path: ...
prometheus.io/port: ...
prometheus.io/scrape: ...
```

## Deploy and configure Prometheus Server

The Prometheus server is needed to configure so that it can discover endpoints of services. If a Prometheus server is already running in cluster and if it is configured in a way that it can discover service endpoints, no extra configuration will be needed. If there is no existing Prometheus server running, rest of this tutorial will create a Prometheus server with appropriate configuration.

The configuration file to `Prometheus-Server` will be provided by `ConfigMap`. The below config map will be created:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-server-conf
  labels:
    name: prometheus-server-conf
  namespace: demo
data:
  prometheus.yml: |-
    global:
      scrape_interval: 5s
      evaluation_interval: 5s
    scrape_configs:
    - job_name: 'kubernetes-service-endpoints'

      kubernetes_sd_configs:
      - role: endpoints

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: replace
        target_label: __scheme__
        regex: (https?)
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_service_name]
        action: replace
        target_label: kubernetes_name
```

```console
$ kubectl create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/monitoring/builtin-prometheus/demo-1.yaml
configmap "prometheus-server-conf" created
```

Now, the below yaml is used to deploy Prometheus in kubernetes :

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus-server
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus-server
  template:
    metadata:
      labels:
        app: prometheus-server
    spec:
      containers:
        - name: prometheus
          image: prom/prometheus:v2.1.0
          args:
            - "--config.file=/etc/prometheus/prometheus.yml"
            - "--storage.tsdb.path=/prometheus/"
          ports:
            - containerPort: 9090
          volumeMounts:
            - name: prometheus-config-volume
              mountPath: /etc/prometheus/
            - name: prometheus-storage-volume
              mountPath: /prometheus/
      volumes:
        - name: prometheus-config-volume
          configMap:
            defaultMode: 420
            name: prometheus-server-conf
        - name: prometheus-storage-volume
          emptyDir: {}
```


Run the following command to deploy prometheus in kubernetes:

```console
$ kubectl create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/monitoring/builtin-prometheus/demo-2.yaml
clusterrole "prometheus-server" created
serviceaccount "prometheus-server" created
clusterrolebinding "prometheus-server" created
deployment "prometheus-server" created
service "prometheus-service" created

# Verify RBAC stuffs
$ kubectl get clusterroles
NAME                AGE
prometheus-server   57s

$ kubectl get clusterrolebindings
NAME                AGE
prometheus-server   1m


$ kubectl get serviceaccounts -n demo
NAME                SECRETS   AGE
default             1         48m
prometheus-server   1         1m
```

### Prometheus Dashboard

Now to open prometheus dashboard on Browser:

```console
$ kubectl get svc -n demo
NAME                         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE
etcd-mon-prometheus          ClusterIP   None             <none>        2379/TCP,2380/TCP   21m
etcd-mon-prometheus-client   ClusterIP   10.101.234.72    <none>        2379/TCP            21m
prometheus-service           NodePort    10.103.137.136   <none>        9090:30901/TCP      6
prometheus-service   LoadBalancer   10.103.201.246   <pending>     9090:30901/TCP        8s


$ minikube ip
192.168.99.100

$ minikube service prometheus-service -n demo --url
http://192.168.99.100:30901
```

Now, open your browser and go to the following URL: _http://{minikube-ip}:{prometheus-svc-nodeport}_ to visit Prometheus Dashboard. According to the above example, this URL will be [http://192.168.99.100:30901](http://192.168.99.100:30901).

Now, if you go the Prometheus Dashboard, you should see that this database endpoint as one of the targets.
![prometheus-builtin](/docs/images/etcd/builtin-prometheus.png)

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete etcd.kubedb.com/etcd-mon-prometheus -n demo

$ kubectl delete  dormantdatabase.kubedb.com/etcd-mon-prometheus -n demo

$ kubectl delete clusterrole prometheus-server
$ kubectl delete clusterrolebindings  prometheus-server
$ kubectl delete serviceaccounts -n demo  prometheus-server

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Monitor your Etcd database with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/etcd/monitoring/using-coreos-prometheus-operator.md).
- Detail concepts of [Etcd object](/docs/concepts/databases/etcd.md).
- [Snapshot and Restore](/docs/guides/etcd/snapshot/backup-and-restore.md) process of Etcd databases using KubeDB.
- Take [Scheduled Snapshot](/docs/guides/etcd/snapshot/scheduled-backup.md) of Etcd databases using KubeDB.
- Initialize [Etcd with Script](/docs/guides/etcd/initialization/using-script.md).
- Initialize [Etcd with Snapshot](/docs/guides/etcd/initialization/using-snapshot.md).
- Use [private Docker registry](/docs/guides/etcd/private-registry/using-private-registry.md) to deploy Etcd with KubeDB.
- Wondering what features are coming next? Please visit [here](/docs/roadmap.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
