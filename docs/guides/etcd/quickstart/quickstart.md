---
title: Etcd Quickstart
menu:
  docs_0.8.0:
    identifier: etcd-quickstart-quickstart
    name: Overview
    parent: etcd-quickstart
    weight: 10
menu_name: docs_0.8.0
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Etcd QuickStart

This tutorial will show you how to use KubeDB to run a Etcd database.

<p align="center">
  <ietcd alt="lifecycle"  src="docs/images/etcd/etcd-lifecycle.png" width="600" height="660">
</p>

The yaml files that are used in this tutorial, stored in [docs/examples](https://github.com/kubedb/cli/tree/master/docs/examples) folder in GitHub repository [kubedb/cli](https://github.com/kubedb/cli).

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
  name: etcd-quickstart
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
```

```console
$ kubedb create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/etcd/quickstart/demo-1.yaml
etcd "etcd-quickstart" created
```

Here,

- `spec.version` is the version of Etcd database. In this tutorial, a Etcd 3.2.13 database is going to be created.
- `spec.doNotPause` tells KubeDB operator that if this object is deleted, it should be automatically reverted. This should be set to true for production databases to avoid accidental deletion.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. Since release 0.8.0, a storage spec is required for Etcd.

KubeDB operator watches for `Etcd` objects using Kubernetes api. When a `Etcd` object is created, KubeDB operator will create a new StatefulSet and a ClusterIP Service with the matching Etcd object name. KubeDB operator will also create a governing service for StatefulSets with the name `kubedb`, if one is not already present. No Etcd specific RBAC permission is required in [RBAC enabled clusters](/docs/setup/install.md#using-yaml).

```console
$ kubedb describe etcd -n demo etcd-quickstart
Name:		etcd-quickstart
Namespace:	demo
StartTimestamp:	Fri, 02 Feb 2018 15:11:58 +0600
Status:		Running
Volume:
  StorageClass:	standard
  Capacity:	50Mi
  Access Modes:	RWO

StatefulSet:
  Name:			etcd-quickstart
  Replicas:		1 current / 1 desired
  CreationTimestamp:	Fri, 02 Feb 2018 15:11:24 +0600
  Pods Status:		1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:
  Name:		etcd-quickstart
  Type:		ClusterIP
  IP:		10.103.114.139
  Port:		db	27017/TCP

Database Secret:
  Name:	etcd-quickstart-auth
  Type:	Opaque
  Data
  ====
  password:	16 bytes
  user:		4 bytes

No Snapshots.


Events:
  FirstSeen   LastSeen   Count     From            Type       Reason             Message
  ---------   --------   -----     ----            --------   ------             -------
  8m          8m         1                         Normal     New Member Added   New member etcdb-quickstart-696p58sv64 added to cluster
  9m          9m         1                         Normal     New Member Added   New member etcdb-quickstart-hfh6pv66td added to cluster
  10m         10m        1                         Normal     New Member Added   New member etcdb-quickstart-s889grj9xr added to cluster
  10m         10m        1         Etcd operator   Normal     Successful         Successfully created Etcd


$ kubectl get statefulset -n demo
NAME             DESIRED   CURRENT   AGE
etcd-quickstart   1         1         4m

$ kubectl get pvc -n demo
NAME                              STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-etcd-quickstart-696p58sv64   Bound     pvc-16158aae-07fa-11e8-946f-080027c05a6e   50Mi       RWO            standard       2m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM                        STORAGECLASS   REASON    AGE
pvc-16158aae-07fa-11e8-946f-080027c05a6e   50Mi       RWO            Delete           Bound     demo/data-etcd-quickstart-696p58sv64   standard                 3m

$ kubectl get service -n demo
NAME                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
etcd-quickstart           ClusterIP   None             <none>        <none>      3m
etcd-quickstart-client    ClusterIP   10.107.133.189   <none>        27017/TCP   3m
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified Etcd object:

```yaml
$ kubedb get etcd -n demo etcd-quickstart -o yaml
apiVersion: kubedb.com/v1alpha1
kind: Etcd
metadata:
  clusterName: ""
  creationTimestamp: 2018-08-01T09:01:15Z
  finalizers:
  - kubedb.com
  generation: 1
  name: etcd-quickstart
  namespace: default
  resourceVersion: "15011"
  selfLink: /apis/kubedb.com/v1alpha1/namespaces/default/etcds/etcd-quickstart
  uid: 726f5576-9569-11e8-95a5-080027c002b2
spec:
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  version: 3.2.13
status:
  phase: Running
```


## Pause Database

KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `doNotPause` feature. If admission webhook is enabled, It prevents user from deleting the database as long as the `spec.doNotPause` is set to true. Since the Etcd object created in this tutorial has `spec.doNotPause` set to true, if you delete the Etcd object, KubeDB operator will nullify the delete operation. You can see this below:

```console
$ kubedb delete etcd etcd-quickstart -n demo
error: Etcd "etcd-quickstart" can't be paused. To continue delete, unset spec.doNotPause and retry.
```

Now, run `kubedb edit etcd etcd-quickstart -n demo` to set `spec.doNotPause` to false or remove this field (which default to false). Then if you delete the Etcd object, KubeDB operator will delete the StatefulSet and its pods, but leaves the PVCs unchanged. In KubeDB parlance, we say that `etcd-quickstart` Etcd database has entered into dormant state. This is represented by KubeDB operator by creating a matching DormantDatabase object.

```console
$ kubedb delete etcd etcd-quickstart -n demo
etcd "etcd-quickstart" deleted

$ kubedb get drmn -n demo etcd-quickstart
NAME             STATUS    AGE
etcd-quickstart   Pausing   39s

$ kubedb get drmn -n demo etcd-quickstart
NAME             STATUS    AGE
etcd-quickstart   Paused    1m
```

```yaml
$ kubedb get drmn -n demo etcd-quickstart -o yaml
apiVersion: kubedb.com/v1alpha1
kind: DormantDatabase
metadata:
  clusterName: ""
  creationTimestamp: 2018-08-01T10:29:32Z
  generation: 1
  labels:
    kubedb.com/kind: Etcd
  name: etcd-quickstart
  namespace: default
  resourceVersion: "21291"
  selfLink: /apis/kubedb.com/v1alpha1/namespaces/default/dormantdatabases/etcd-quickstart
  uid: c75f9034-9575-11e8-95a5-080027c002b2
spec:
  origin:
    metadata:
      creationTimestamp: 2018-08-01T09:01:15Z
      name: etcd-quickstart
      namespace: default
    spec:
      etcd:
        backupSchedule:
          cronExpression: '@every 8m'
          gcs:
            bucket: kubedbetcd
          resources: {}
          storageSecretName: gcs-secret
        replicas: 3
        resources: {}
        storage:
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
          storageClassName: standard
        version: 3.2.13
status:
  creationTime: 2018-08-01T10:29:32Z
  pausingTime: 2018-08-01T10:30:38Z
  phase: Paused
```

Here,

- `spec.origin` is the spec of the original spec of the original Etcd object.
- `status.phase` points to the current database state `Paused`.

## Resume Dormant Database

To resume the database from the dormant state, create same `Etcd` object with same Spec.

In this tutorial, the dormant database can be resumed by creating original Etcd object.

The below command will resume the DormantDatabase `etcd-quickstart`.

```console
$ kubedb create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/etcd/quickstart/demo-1.yaml
etcd "etcd-quickstart" created
```

## WipeOut DormantDatabase

You can wipe out a DormantDatabase while deleting the objet by setting `spec.wipeOut` to true. KubeDB operator will delete any relevant resources of this `Etcd` database (i.e, PVCs, Secrets, Snapshots). It will also delete snapshot data stored in the Cloud Storage buckets.

```yaml
$ kubedb edit drmn -n demo etcd-quickstart
apiVersion: kubedb.com/v1alpha1
kind: DormantDatabase
metadata:
  name: etcd-quickstart
  namespace: demo
  ...
spec:
  wipeOut: true
  ...
status:
  phase: Paused
  ...
```

If `spec.wipeOut` is not set to true while deleting the `dormantdatabase` object, then only this object will be deleted and `kubedb-operator` won't delete related Secrets, PVCs and Snapshots. So, user still can access the stored data in the cloud storage buckets as well as PVCs.

## Delete DormantDatabase

As it is already discussed above, `DormantDatabase` can be deleted with or without wiping out the resources. To delete the `dormantdatabase`,

```console
$ kubectl delete drmn etcd-quickstart -n demo
dormantdatabase "etcd-quickstart" deleted
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl patch -n demo etcd/etcd-quickstart -p '{"spec":{"doNotPause":false}}' --type="merge"
$ kubectl delete -n demo etcd/etcd-quickstart

$ kubectl patch -n demo drmn/etcd-quickstart -p '{"spec":{"wipeOut":true}}' --type="merge"
$ kubectl delete -n demo drmn/etcd-quickstart

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- [Snapshot and Restore](/docs/guides/etcd/snapshot/backup-and-restore.md) process of Etcd databases using KubeDB.
- Take [Scheduled Snapshot](/docs/guides/etcd/snapshot/scheduled-backup.md) of Etcd databases using KubeDB.
- Initialize [Etcd with Script](/docs/guides/etcd/initialization/using-script.md).
- Initialize [Etcd with Snapshot](/docs/guides/etcd/initialization/using-snapshot.md).
- Monitor your Etcd database with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/etcd/monitoring/using-coreos-prometheus-operator.md).
- Monitor your Etcd database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/etcd/private-registry/using-private-registry.md) to deploy Etcd with KubeDB.
- Detail concepts of [Etcd object](/docs/concepts/databases/etcd.md).
- Wondering what features are coming next? Please visit [here](/docs/roadmap.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
