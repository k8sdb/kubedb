---
title: Instant Backup of Etcd
menu:
  docs_0.8.0:
    identifier: etcd-backup-and-restore-snapshot
    name: Instant Backup
    parent: etcd-snapshot-etcd
    weight: 10
menu_name: docs_0.8.0
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Database Snapshots

This tutorial will show you how to take snapshots of a KubeDB managed Etcd database.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

A `Etcd` database is needed to take snapshot for this tutorial. To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```console
$ kubectl create ns demo
namespace "demo" created

$ kubectl get ns
NAME          STATUS    AGE
default       Active    1h
demo          Active    1m
kube-public   Active    1h
kube-system   Active    1h

$ kubedb create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/etcd/snapshot/demo-1.yaml
etcd "etcd-infant" created
```

Note that the yaml files that are used in this tutorial, stored in [docs/examples](https://github.com/kubedb/cli/tree/master/docs/examples) folder in GitHub repository [kubedb/cli](https://github.com/kubedb/cli).

## Instant Backups

You can easily take a snapshot of `Etcd` database by creating a `Snapshot` object. When a `Snapshot` object is created, KubeDB operator will launch a Job that runs the `etcddump` command and uploads the output bson file to various cloud providers S3, GCS, Azure, OpenStack Swift and/or locally mounted volumes using [osm](https://github.com/appscode/osm).

In this tutorial, snapshots will be stored in a Google Cloud Storage (GCS) bucket. To do so, a secret is needed that has the following 2 keys:

| Key                               | Description                                                |
|-----------------------------------|------------------------------------------------------------|
| `GOOGLE_PROJECT_ID`               | `Required`. Google Cloud project ID                        |
| `GOOGLE_SERVICE_ACCOUNT_JSON_KEY` | `Required`. Google Cloud service account json key          |

```console
$ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
$ mv downloaded-sa-json.key > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ kubectl create secret generic etcd-snap-secret -n demo \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret "etcd-snap-secret" created
```

```yaml
$ kubectl get secret etcd-snap-secret -n demo -o yaml
apiVersion: v1
data:
  GOOGLE_PROJECT_ID: PHlvdXItcHJvamVjdC1pZD4=
  GOOGLE_SERVICE_ACCOUNT_JSON_KEY: ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3V...9tIgp9Cg==
kind: Secret
metadata:
  creationTimestamp: 2018-02-02T10:02:09Z
  name: etcd-snap-secret
  namespace: demo
  resourceVersion: "48679"
  selfLink: /api/v1/namespaces/demo/secrets/etcd-snap-secret
  uid: 220a7c60-0800-11e8-946f-080027c05a6e
type: Opaque
```

To lean how to configure other storage destinations for Snapshots, please visit [here](/docs/concepts/snapshot.md). Now, create the Snapshot object.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Snapshot
metadata:
  name: snapshot-infant
  namespace: demo
  labels:
    kubedb.com/kind: Etcd
spec:
  databaseName: etcd-infant
  storageSecretName: etcd-snap-secret
  gcs:
    bucket: restic
```

```console
$ kubedb create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/etcd/snapshot/demo-2.yaml
snapshot "snapshot-infant" created

$ kubedb get snap -n demo
NAME              DATABASE        STATUS    AGE
snapshot-infant   etcd/etcd-infant   Running   47s
```

```yaml
$ kubedb get snap -n demo snapshot-infant -o yaml
apiVersion: kubedb.com/v1alpha1
kind: Snapshot
metadata:
  clusterName: ""
  creationTimestamp: 2018-08-01T09:01:15Z
  finalizers:
  - kubedb.com
  generation: 0
  labels:
    kubedb.com/kind: Etcd
    kubedb.com/name: etcd-infant
  name: snapshot-infant
  namespace: demo
  resourceVersion: "48991"
  selfLink: /apis/kubedb.com/v1alpha1/namespaces/demo/snapshots/snapshot-infant
  uid: 9d4f37a0-0800-11e8-946f-080027c05a6e
spec:
  databaseName: etcd-infant
  gcs:
    bucket: restic
  storageSecretName: etcd-snap-secret
status:
  completionTime: 2018-02-02T10:06:43Z
  phase: Succeeded
  startTime: 2018-02-02T10:05:37Z
```

Here,

- `metadata.labels` should include the type of database `kubedb.com/kind: Etcd` whose snapshot will be taken.
- `spec.databaseName` points to the database whose snapshot is taken.
- `spec.storageSecretName` points to the Secret containing the credentials for snapshot storage destination.
- `spec.gcs.bucket` points to the bucket name used to store the snapshot data.

You can also run the `kubedb describe` command to see the recent snapshots taken for a database.

```console
$ kubedb describe etcd -n demo etcd-infant
Name:		etcd-infant
Namespace:	demo
StartTimestamp:	Fri, 02 Feb 2018 16:04:50 +0600
Status:		Running
Volume:
  StorageClass:	standard
  Capacity:	50Mi
  Access Modes:	RWO

StatefulSet:
  Name:			etcd-infant
  Replicas:		1 current / 1 desired
  CreationTimestamp:	Fri, 02 Feb 2018 16:04:56 +0600
  Pods Status:		1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:
  Name:		etcd-infant
  Type:		ClusterIP
  IP:		10.99.34.23
  Port:		db	27017/TCP

Database Secret:
  Name:	etcd-infant-auth
  Type:	Opaque
  Data
  ====
  password:	16 bytes
  user:		4 bytes

Snapshots:
  Name              Bucket      StartTime                         CompletionTime                    Phase
  ----              ------      ---------                         --------------                    -----
  snapshot-infant   gs:restic   Fri, 02 Feb 2018 16:05:37 +0600   Fri, 02 Feb 2018 16:06:43 +0600   Succeeded

Events:
  FirstSeen   LastSeen   Count     From                  Type       Reason               Message
  ---------   --------   -----     ----                  --------   ------               -------
  8m          8m         1         Job Controller        Normal     SuccessfulSnapshot   Successfully completed snapshot
  9m          9m         1         Snapshot Controller   Normal     Starting             Backup running
  9m          9m         1         Etcd operator      Normal     Successful           Successfully patched StatefulSet
  9m          9m         1         Etcd operator      Normal     Successful           Successfully patched Etcd
  9m          9m         1         Etcd operator      Normal     Successful           Successfully created StatefulSet
  9m          9m         1         Etcd operator      Normal     Successful           Successfully created Etcd
  10m         10m        1         Etcd operator      Normal     Successful           Successfully created Service
```

Once the snapshot Job is complete, you should see the output of the `etcddump` command stored in the GCS bucket.

![snapshot-console](/docs/images/etcd/etcd-xyz-snapshot.png)

From the above image, you can see that the snapshot output is stored in a folder called `{bucket}/kubedb/{namespace}/{etcd-object}/{snapshot}/`.

## Restore from Snapshot

You can create a new database from a previously taken Snapshot. Specify the Snapshot name in the `spec.init.snapshotSource` field of a new Etcd object. See the example `etcd-recovered` object below:

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
      name: snapshot
      namespace: demo
```

```console
$ kubedb create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/etcd/snapshot/demo-3.yaml
etcd "etcd-recovered" created
```

Here,

- `spec.init.snapshotSource.name` refers to a Snapshot object for a Etcd database in the same namespaces as this new `etcd-recovered` Etcd object.

Now, wait several seconds. KubeDB operator will create a new StatefulSet. Then KubeDB operator launches a Kubernetes Job to initialize the new database using the data from `snapshot-infant` Snapshot.

```console
$ kubedb get etcd -n demo
NAME            STATUS    AGE
etcd-infant      Running   23m
etcd-recovered   Running   4m

$ kubedb describe etcd -n demo etcd-recovered
Name:		etcd-recovered
Namespace:	demo
StartTimestamp:	Fri, 02 Feb 2018 16:24:23 +0600
Status:		Running
Annotations:	kubedb.com/initialized=
Volume:
  StorageClass:	standard
  Capacity:	50Mi
  Access Modes:	RWO

StatefulSet:
  Name:			etcd-recovered
  Replicas:		1 current / 1 desired
  CreationTimestamp:	Fri, 02 Feb 2018 16:24:36 +0600
  Pods Status:		1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:
  Name:		etcd-recovered
  Type:		ClusterIP
  IP:		10.107.157.253
  Port:		db	27017/TCP

Database Secret:
  Name:	etcd-recovered-auth
  Type:	Opaque
  Data
  ====
  user:		4 bytes
  password:	16 bytes

No Snapshots.


Events:
  FirstSeen   LastSeen   Count     From            Type       Reason             Message
  ---------   --------   -----     ----            --------   ------             -------
  10m         10m        1                         Normal     New Member Added   New member etcd-mon-prometheus-8slp4xxxl8 added to cluster
  11m         11m        1                         Normal     New Member Added   New member etcd-mon-prometheus-7pvzjcd7dx added to cluster
  12m         12m        1                         Normal     New Member Added   New member etcd-mon-prometheus-ld7n576tv5 added to cluster
  12m         12m        1         Etcd operator   Normal     Successful         Successfully created Etcd

```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl patch -n demo etcd/etcd-infant etcd/etcd-recovered -p '{"spec":{"doNotPause":false}}' --type="merge"
$ kubectl delete -n demo etcd/etcd-infant etcd/etcd-recovered

$ kubectl patch -n demo drmn/etcd-infant drmn/etcd-recovered -p '{"spec":{"wipeOut":true}}' --type="merge"
$ kubectl delete -n demo drmn/etcd-infant drmn/etcd-recovered

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- See the list of supported storage providers for snapshots [here](/docs/concepts/snapshot.md).
- Take [Scheduled Snapshot](/docs/guides/etcd/snapshot/scheduled-backup.md) of Etcd databases using KubeDB.
- Initialize [Etcd with Script](/docs/guides/etcd/initialization/using-script.md).
- Initialize [Etcd with Snapshot](/docs/guides/etcd/initialization/using-snapshot.md).
- Monitor your Etcd database with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/etcd/monitoring/using-coreos-prometheus-operator.md).
- Monitor your Etcd database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/etcd/private-registry/using-private-registry.md) to deploy Etcd with KubeDB.
- Detail concepts of [Etcd object](/docs/concepts/databases/etcd.md).
- Wondering what features are coming next? Please visit [here](/docs/roadmap.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
