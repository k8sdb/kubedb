---
title: Scheduled Backup of Etcd
menu:
  docs_0.8.0:
    identifier: etcd-scheduled-backup-snapshot
    name: Scheduled Backup
    parent: etcd-snapshot
    weight: 10
menu_name: docs_0.8.0
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Database Scheduled Snapshots

This tutorial will show you how to use KubeDB to take scheduled snapshot of a Etcd database.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```console
$ kubectl create ns demo
namespace "demo" created

$ kubectl get ns
NAME          STATUS    AGE
default       Active    1h
demo          Active    1m
kube-public   Active    1h
kube-system   Active    1h
```

Note that the yaml files that are used in this tutorial, stored in [docs/examples](https://github.com/kubedb/cli/tree/master/docs/examples) folder in GitHub repository [kubedb/cli](https://github.com/kubedb/cli).

## Scheduled Backups

KubeDB supports taking periodic backups for a database using a [cron expression](https://github.com/robfig/cron/blob/v2/doc.go#L26).  KubeDB operator will launch a Job periodically that runs the `etcddump` command and uploads the output bson file to various cloud providers S3, GCS, Azure, OpenStack Swift and/or locally mounted volumes using [osm](https://github.com/appscode/osm).

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

To learn how to configure other storage destinations for Snapshots, please visit [here](/docs/concepts/snapshot.md).  Now, create the `Etcd` object with scheduled snapshot.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Etcd
metadata:
  name: etcd-scheduled
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
  backupSchedule:
    cronExpression: "@every 1m"
    storageSecretName: etcd-snap-secret
    gcs:
      bucket: restic
```

```console
$ kubedb create -f https://raw.githubusercontent.com/kubedb/cli/0.8.0/docs/examples/etcd/snapshot/demo-4.yaml
etcd "etcd-scheduled" created
```

It is also possible to add  backup scheduler to an existing `Etcd`. You just have to edit the `Etcd` CRD and add below spec:

```yaml
$ kubedb edit etcd {db-name} -n demo
spec:
  backupSchedule:
    cronExpression: '@every 1m'
    gcs:
      bucket: restic
    storageSecretName: etcd-snap-secret
```

Once the `spec.backupSchedule` is added, KubeDB operator will create a new Snapshot object on each tick of the cron expression. This triggers KubeDB operator to create a Job as it would for any regular instant backup process. You can see the snapshots as they are created using `kubedb get snap` command.

```console
 $ kubectl get snapshot --all-namespaces
NAMESPACE   NAME                         AGE
default     etcdb-demo-20180801-091507   3h
default     etcdb-demo-20180801-092009   2h
default     etcdb-demo-20180801-093509   2h
default     etcdb-demo-20180801-100524   2h
default     etcdb-demo-20180801-101324   2h
default     etcdb-demo-20180801-102924   1h

```

you should see the output of the `etcddump` command for each snapshot stored in the GCS bucket.

## Remove Scheduler

To remove scheduler, edit the Etcd object  to remove `spec.backupSchedule` section.

```yaml
$ kubedb edit etcd etcd-scheduled -n demo
apiVersion: kubedb.com/v1alpha1
kind: Etcd
metadata:
  name: etcd-scheduled
  namespace: demo
  ...
spec:
# backupSchedule:
#   cronExpression: '@every 1m'
#   gcs:
#     bucket: restic
#   storageSecretName: etcd-snap-secret
  databaseSecret:
    secretName: etcd-scheduled-auth
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
    storageClassName: standard
  version: 3.4
status:
  creationTime: 2018-02-02T10:46:18Z
  phase: Running
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
$ kubectl patch -n demo etcd/etcd-scheduled -p '{"spec":{"doNotPause":false}}' --type="merge"
$ kubectl delete -n demo etcd/etcd-scheduled

$ kubectl patch -n demo drmn/etcd-scheduled -p '{"spec":{"wipeOut":true}}' --type="merge"
$ kubectl delete -n demo drmn/etcd-scheduled

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- See the list of supported storage providers for snapshots [here](/docs/concepts/snapshot.md).
- Initialize [Etcd with Script](/docs/guides/etcd/initialization/using-script.md).
- Initialize [Etcd with Snapshot](/docs/guides/etcd/initialization/using-snapshot.md).
- Monitor your Etcd database with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/etcd/monitoring/using-coreos-prometheus-operator.md).
- Monitor your Etcd database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/etcd/private-registry/using-private-registry.md) to deploy Etcd with KubeDB.
- Detail concepts of [Etcd object](/docs/concepts/databases/etcd.md).
- Wondering what features are coming next? Please visit [here](/docs/roadmap.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
