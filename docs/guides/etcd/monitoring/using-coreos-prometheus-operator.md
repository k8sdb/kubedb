---
title: Monitor Etcd using Coreos Prometheus Operator
menu:
  docs_0.8.0:
    identifier: etcd-using-coreos-prometheus-operator-monitoring
    name: Coreos Prometheus Operator
    parent: etcd-monitoring
    weight: 15
menu_name: docs_0.8.0
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Using Prometheus (CoreOS operator) with KubeDB

This tutorial will show you how to monitor KubeDB databases using Prometheus via [CoreOS Prometheus Operator](https://github.com/coreos/prometheus-operator).

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

Note that the yaml files that are used in this tutorial, stored in [docs/examples](https://github.com/kubedb/cli/tree/master/docs/examples) folder in GitHub repository [kubedb/cli](https://github.com/kubedb/cli).

## Deploy CoreOS-Prometheus Operator

Run the following command to prepare your cluster for this tutorial:

```console
$ kubectl create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/monitoring/coreos-operator/demo-0.yaml
namespace "demo" created
clusterrole "prometheus-operator" created
serviceaccount "prometheus-operator" created
clusterrolebinding "prometheus-operator" created
deployment "prometheus-operator" created

$ kubectl get pods -n demo --watch
NAME                                   READY     STATUS    RESTARTS   AGE
prometheus-operator-79cb9dcd4b-2njgq   1/1       Running   0          2m


$ kubectl get crd
NAME                                    AGE
alertmanagers.monitoring.coreos.com     11m
prometheuses.monitoring.coreos.com      11m
servicemonitors.monitoring.coreos.com   11m
```

Once the Prometheus operator CRDs are registered, run the following command to create a Prometheus.

```console
$ kubectl create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/monitoring/coreos-operator/rbac/demo-1.yaml
clusterrole "prometheus" created
serviceaccount "prometheus" created
clusterrolebinding "prometheus" created
prometheus "prometheus" created
service "prometheus" created

# Verify RBAC stuffs
$ kubectl get clusterroles
NAME                  AGE
prometheus            48s
prometheus-operator   1m

$ kubectl get clusterrolebindings
NAME                  AGE
prometheus            7s
prometheus-operator   25s

$ kubectl get serviceaccounts -n demo
NAME                  SECRETS   AGE
default               1         5m
prometheus            1         4m
prometheus-operator   1         5m
```

### Prometheus Dashboard

Now to open prometheus dashboard on Browser:

```console
$ kubectl get svc -n demo
NAME                  TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
prometheus            LoadBalancer   10.99.201.154   <pending>     9090:30900/TCP   5m
prometheus-operated   ClusterIP      None            <none>        9090/TCP         5m

$ minikube ip
192.168.99.100

$ minikube service prometheus -n demo --url
http://192.168.99.100:30900
```

Now, open your browser and go to the following URL: _http://{minikube-ip}:{prometheus-svc-nodeport}_ to visit Prometheus Dashboard. According to the above example, this URL will be [http://192.168.99.100:30900](http://192.168.99.100:30900).

## Create a Etcd database

KubeDB implements a `Etcd` CRD to define the specification of a Etcd database. Below is the `Etcd` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Etcd
metadata:
  name: etcd-mon-coreos
  namespace: demo
spec:
  version: "3.4"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  monitor:
    agent: prometheus.io/coreos-operator
    prometheus:
      namespace: demo
      labels:
        app: kubedb
      interval: 10s
```

The `Etcd` CRD object contains `monitor` field in it's `spec`.  It is also possible to add CoreOS-Prometheus monitor to an existing `Etcd` database by adding the below part in it's `spec` field.

```yaml
spec:
  monitor:
    agent: prometheus.io/coreos-operator
    prometheus:
      namespace: demo
      labels:
        app: kubedb
      interval: 10s
```

|  Keys  |  Value |  Description |
|--|--|--|
| `spec.monitor.agent` | string | `Required`. Indicates the monitoring agent used. Only valid value currently is `coreos-prometheus-operator` |
| `spec.monitor.prometheus.namespace` | string | `Required`. Indicates namespace where service monitors are created. This must be the same namespace of the Prometheus instance. |
| `spec.monitor.prometheus.labels` | map | `Required`. Indicates labels applied to service monitor.                                                    |
| `spec.monitor.prometheus.interval` | string | `Optional`. Indicates the scrape interval for database exporter endpoint (eg, '10s')                        |
| `spec.monitor.prometheus.port` | int |`Optional`. Indicates the port for database exporter endpoint (default is `56790`)|

__Known Limitations:__ If the database password is updated, exporter must be restarted to use the new credentials. This issue is tracked [here](https://github.com/kubedb/project/issues/53).

Run the following command to deploy the above `Etcd` CRD object.

```console
$ kubedb create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/etcd/monitoring/coreos-operator/demo-1.yaml
etcd "etcd-mon-coreos" created
```

Here,

- `spec.version` is the version of Etcd database. In this tutorial, a Etcd 3.4 database is going to be created.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. Since release 0.8.0, a storage spec is required for Etcd.
- `spec.monitor` specifies that CoreOS Prometheus operator is used to monitor this database instance. A ServiceMonitor should be created in the `demo` namespace with label `app=kubedb`. The exporter endpoint should be scrapped every 10 seconds.

KubeDB operator watches for `Etcd` objects using Kubernetes api. When a `Etcd` object is created, KubeDB operator will create a new StatefulSet and a ClusterIP Service with the matching crd name. KubeDB operator will also create a governing service for StatefulSets with the name `kubedb`, if one is not already present.

```console
$ kubedb get etcd -n demo
NAME             STATUS    AGE
etcd-mon-coreos   Creating  36s

$ kubedb get etcd -n demo
NAME             STATUS    AGE
etcd-mon-coreos   Running   1m

$ kubedb describe etcd -n demo etcd-mon-coreos
Name:		etcd-mon-coreos
Namespace:	demo
StartTimestamp:	Mon, 05 Feb 2018 11:20:20 +0600
Status:		Running
Volume:
  StorageClass:	standard
  Capacity:	50Mi
  Access Modes:	RWO

StatefulSet:
  Name:			etcd-mon-coreos
  Replicas:		1 current / 1 desired
  CreationTimestamp:	Mon, 05 Feb 2018 11:20:27 +0600
  Pods Status:		1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:
  Name:		etcd-mon-coreos
  Type:		ClusterIP
  IP:		10.107.145.36
  Port:		db		27017/TCP
  Port:		prom-http	56790/TCP

Database Secret:
  Name:	etcd-mon-coreos-auth
  Type:	Opaque
  Data
  ====
  password:	16 bytes
  user:		4 bytes

Monitoring System:
  Agent:	prometheus.io/coreos-operator
  Prometheus:
    Namespace:	demo
    Labels:	app=kubedb
    Interval:	10s

No Snapshots.

Events:
  FirstSeen   LastSeen   Count     From            Type       Reason             Message
  ---------   --------   -----     ----            --------   ------             -------
  10m         10m        1                         Normal     New Member Added   New member etcd-mon-prometheus-8slp4xxxl8 added to cluster
  11m         11m        1                         Normal     New Member Added   New member etcd-mon-prometheus-7pvzjcd7dx added to cluster
  12m         12m        1                         Normal     New Member Added   New member etcd-mon-prometheus-ld7n576tv5 added to cluster
  12m         12m        1         Etcd operator   Normal     Successful         Successfully created Etcd
```

Since `spec.monitoring` was configured, a ServiceMonitor object is created accordingly. You can verify it running the following commands:

```yaml
$ kubectl get servicemonitor -n demo
NAME                         AGE
kubedb-demo-etcd-mon-coreos   11m

$ kubectl get servicemonitor -n demo kubedb-demo-etcd-mon-coreos -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  clusterName: ""
  creationTimestamp: 2018-02-05T05:20:46Z
  labels:
    app: kubedb
    monitoring.appscode.com/service: etcd-mon-coreos.demo
  name: kubedb-demo-etcd-mon-coreos
  namespace: demo
  resourceVersion: "57754"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/demo/servicemonitors/kubedb-demo-etcd-mon-coreos
  uid: 5215258a-0a34-11e8-8d7f-080027c05a6e
spec:
  endpoints:
  - interval: 10s
    path: /kubedb.com/v1alpha1/namespaces/demo/etcds/etcd-mon-coreos/metrics
    port: prom-http
    targetPort: 0
  namespaceSelector:
    matchNames:
    - demo
  selector:
    matchLabels:
      kubedb.com/kind: Etcd
      kubedb.com/name: etcd-mon-coreos
```

Now, if you go the Prometheus Dashboard, you should see that this database endpoint as one of the targets.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl patch -n demo etcd/etcd-mon-coreos -p '{"spec":{"doNotPause":false}}' --type="merge"
$ kubectl delete -n demo etcd/etcd-mon-coreos

$ kubectl patch -n demo drmn/etcd-mon-coreos -p '{"spec":{"wipeOut":true}}' --type="merge"
$ kubectl delete -n demo drmn/etcd-mon-coreos

$ kubectl delete clusterrolebindings prometheus-operator  prometheus
$ kubectl delete clusterrole prometheus-operator prometheus

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Monitor your Etcd database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Etcd object](/docs/concepts/databases/etcd.md).
- [Snapshot and Restore](/docs/guides/etcd/snapshot/backup-and-restore.md) process of Etcd databases using KubeDB.
- Take [Scheduled Snapshot](/docs/guides/etcd/snapshot/scheduled-backup.md) of Etcd databases using KubeDB.
- Initialize [Etcd with Script](/docs/guides/etcd/initialization/using-script.md).
- Initialize [Etcd with Snapshot](/docs/guides/etcd/initialization/using-snapshot.md).
- Use [private Docker registry](/docs/guides/etcd/private-registry/using-private-registry.md) to deploy Etcd with KubeDB.
- Wondering what features are coming next? Please visit [here](/docs/roadmap.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
