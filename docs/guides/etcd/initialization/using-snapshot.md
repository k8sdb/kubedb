---
title: Initialize Etcd from Snapshot
menu:
  docs_0.8.0:
    identifier: etcd-using-snapshot-initialization
    name: From Snapshot
    parent: etcd-initialization
    weight: 15
menu_name: docs_0.8.0
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Initialize Etcd with Snapshot

This tutorial will show you how to use KubeDB to initialize a Etcd database with an existing snapshot.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

This tutorial assumes that you have created a namespace `demo` and a snapshot `snapshot-infant`. Follow the steps [here](/docs/guides/etcd/snapshot/backup-and-restore.md) to create a database and take [instant snapshot](/docs/guides/etcd/snapshot/backup-and-restore.md#instant-backups), if you have not done so already. If you have changed the name of either namespace or snapshot object, please modify the YAMLs used in this tutorial accordingly.

Note that the yaml files that are used in this tutorial, stored in [docs/examples](https://github.com/kubedb/cli/tree/master/docs/examples) folder in GitHub repository [kubedb/cli](https://github.com/kubedb/cli).

## Create Etcd with Init-Snapshot

Below is the `Etcd` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Etcd
metadata:
  name: etcd-init-snapshot
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
  init:
    snapshotSource:
      name: snapshot-infant
      namespace: demo

```

```console
$ kubedb create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/etcd/Initialization/demo-2.yaml
etcd "etcd-init-snapshot" created
```

Here,

- `spec.init.snapshotSource.name` refers to a Snapshot object for a Etcd database in the same namespaces as this new `etcd-init-snapshot` Etcd object.

Now, wait several seconds. KubeDB operator will create new `Pod`. Then KubeDB operator launches a Kubernetes Job to initialize the new database using the data from `snapshot-infant` Snapshot.

```console
$ kubedb describe etcd  etcd-init-snapshot
Name:		etcd-init-snapshot
Namespace:	default
StartTimestamp:	Wed, 01 Aug 2018 16:55:39 +0600
Replicas:	3  total
Status:		Running
Volume:
  StorageClass:	standard
  Capacity:	1Gi
  Access Modes:	RWO

Service:
  Name:		etcd-init-snapshot
  Type:		ClusterIP
  IP:		None
  Port:		client	2379/TCP
  Port:		peer	2380/TCP

Service:
  Name:		etcd-init-snapshot-client
  Type:		ClusterIP
  IP:		10.96.76.132
  Port:		client	2379/TCP

No Snapshots.

Events:
  FirstSeen   LastSeen   Count     From            Type       Reason             Message
  ---------   --------   -----     ----            --------   ------             -------
  1m          1m         1                         Normal     New Member Added   New member etcd-init-snapshot-8wj94xljxl added to cluster
  2m          2m         1                         Normal     New Member Added   New member etcd-init-snapshot-hgtxqfc4z8 added to cluster
  3m          3m         1                         Normal     New Member Added   New member etcd-init-snapshot-rb9b5clhs5 added to cluster
  3m          3m         1         Etcd operator   Normal     Successful         Successfully created Etcd
  3m          3m         1         Etcd operator   Normal     Initializing       Initializing from Snapshot: "snapshot"
```

To show existing data `kubectl exec` into a etcd pod and run following command:

```
ETCDCTL_API=3 etcdctl get "" --prefix=true
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete snapshot.kubedb.com/snapshot etcd.kubedb.com/etcd-init-snapshot

$ dormantdatabase.kubedb.com/etcd-init-snapshot

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Monitor your Etcd database with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/etcd/monitoring/using-coreos-prometheus-operator.md).
- Monitor your Etcd database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/etcd/private-registry/using-private-registry.md) to deploy Etcd with KubeDB.
- Detail concepts of [Etcd object](/docs/concepts/databases/etcd.md).
- Detail concepts of [Snapshot object](/docs/concepts/snapshot.md).
- Wondering what features are coming next? Please visit [here](/docs/roadmap.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
