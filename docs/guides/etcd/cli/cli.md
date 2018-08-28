---
title: CLI | KubeDB
menu:
  docs_0.8.0:
    identifier: etcd-cli-cli
    name: Quickstart
    parent: etcd-cli-mongodb
    weight: 10
menu_name: docs_0.8.0
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Manage KubeDB objects using CLIs

## KubeDB CLI

KubeDB comes with its own cli. It is called `kubedb` cli. `kubedb` can be used to manage any Etcd object. `kubedb` cli also performs various validations to improve ux. To install KubeDB cli on your workstation, follow the steps [here](/docs/setup/install.md).

### How to Create objects

`kubedb create` creates a database CRD object in `default` namespace by default. Following command will create a Etcd object as specified in `etcd.yaml`.

```console
$ kubedb create -f etcd-demo.yaml
etcd.kubedb.com "etcdb-demo" created
```

You can provide namespace as a flag `--namespace`. Provided namespace should match with namespace specified in input file.

```console
$ kubedb create -f etcd-demo.yaml --namespace=kube-system
etcd.kubedb.com "etcdb-demo" created
```

`kubedb create` command also considers `stdin` as input.

```console
cat etcd-demo.yaml | kubedb create -f -
```

To learn about various options of `create` command, please visit [here](/docs/reference/kubedb_create.md).

### How to List Objects

`kubedb get` command allows users to list or find any KubeDB object. To list all MongoDB objects in `default` namespace, run the following command:

```console
$ kubedb get etcd
NAME         STATUS    AGE
etcdb-demo   Running   43s
```

To get YAML of an object, use `--output=yaml` flag.

```yaml
$ kubedb get etcd etcdb-demo --output=yaml
apiVersion: kubedb.com/v1alpha1
kind: Etcd
metadata:
  clusterName: ""
  creationTimestamp: 2018-08-01T09:01:15Z
  finalizers:
  - kubedb.com
  generation: 1
  name: etcdb-demo
  namespace: default
  resourceVersion: "15011"
  selfLink: /apis/kubedb.com/v1alpha1/namespaces/default/etcds/etcdb-demo
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

To get JSON of an object, use `--output=json` flag.

```console
$ kubedb get etcd etcdb-demo --output=json
```

To list all KubeDB objects, use following command:

```console
$ kubedb get all -o wide
NAME               VERSION   STATUS    AGE
etcds/etcdb-demo   3.2.13    Running   9m

```

Flag `--output=wide` is used to print additional information.

List command supports short names for each object types. You can use it like `kubedb get <short-name>`. Below are the short name for KubeDB objects:

- Etcd: `etcd`
- Snapshot: `snap`
- DormantDatabase: `drmn`

You can print labels with objects. The following command will list all Snapshots with their corresponding labels.

```console
$ kubedb get snap --show-labels
NAME                            DATABASE                STATUS      AGE       LABELS
mongodb-demo-20170605-073557    etcd/mongodb-demo         Succeeded   11m       kubedb.com/kind=MongoDB,kubedb.com/name=mongodb-demo
snapshot-20171212-114700        etcd/mongodb-demo         Succeeded   1h        kubedb.com/kind=MongoDB,kubedb.com/name=mongodb-demo
```

You can also filter list using `--selector` flag.

```console
$ kubedb get snap --selector='kubedb.com/kind=MongoDB' --show-labels
NAME                            DATABASE           STATUS      AGE       LABELS
mongodb-demo-20171212-073557    etcd/mongodb-demo    Succeeded   14m       kubedb.com/kind=MongoDB,kubedb.com/name=mongodb-demo
snapshot-20171212-114700        etcd/mongodb-demo    Succeeded   2h        kubedb.com/kind=MongoDB,kubedb.com/name=mongodb-demo
```

To print only object name, run the following command:

```console
$ kubedb get all -o name
etcd.kubedb.com/etcdb-demo
snapshot/etcdb-demo-20170605-073557
snapshot/snapshot-20170505-114700
```

To learn about various options of `get` command, please visit [here](/docs/reference/kubedb_get.md).

### How to Describe Objects

`kubedb describe` command allows users to describe any KubeDB object. The following command will describe MongoDB database `etcdb-demo` with relevant information.

```console
$ kubedb describe etcd etcdb-demo
Name:		etcdb-demo
Namespace:	default
StartTimestamp:	Wed, 01 Aug 2018 15:01:15 +0600
Replicas:	3  total
Status:		Running
Volume:
  StorageClass:	standard
  Capacity:	1Gi
  Access Modes:	RWO

Service:
  Name:		etcdb-demo
  Type:		ClusterIP
  IP:		None
  Port:		client	2379/TCP
  Port:		peer	2380/TCP

Service:
  Name:		etcdb-demo-client
  Type:		ClusterIP
  IP:		10.110.93.14
  Port:		client	2379/TCP

No Snapshots.

Events:
  FirstSeen   LastSeen   Count     From            Type       Reason             Message
  ---------   --------   -----     ----            --------   ------             -------
  8m          8m         1                         Normal     New Member Added   New member etcdb-demo-696p58sv64 added to cluster
  9m          9m         1                         Normal     New Member Added   New member etcdb-demo-hfh6pv66td added to cluster
  10m         10m        1                         Normal     New Member Added   New member etcdb-demo-s889grj9xr added to cluster
  10m         10m        1         Etcd operator   Normal     Successful         Successfully created Etcd

```

`kubedb describe` command provides following basic information about a Etcd database.

- Pd
- Storage (Persistent Volume)
- Service
- Secret (If available)
- Snapshots (If any)
- Monitoring system (If available)

To hide details about StatefulSet & Service, use flag `--show-workload=false`
To hide details about Secret, use flag `--show-secret=false`
To hide events on KubeDB object, use flag `--show-events=false`

To describe all MongoDB objects in `default` namespace, use following command

```console
$ kubedb describe etcd
```

To describe all MongoDB objects from every namespace, provide `--all-namespaces` flag.

```console
$ kubedb describe etcd --all-namespaces
```

To describe all KubeDB objects from every namespace, use the following command:

```console
$ kubedb describe all --all-namespaces
```

You can also describe KubeDb objects with matching labels. The following command will describe all MongoDB objects with specified labels from every namespace.

```console
$ kubedb describe etcd --all-namespaces --selector='group=dev'
```

To learn about various options of `describe` command, please visit [here](/docs/reference/kubedb_describe.md).

### How to Edit Objects

`kubedb edit` command allows users to directly edit any KubeDB object. It will open the editor defined by _KUBEDB_EDITOR_, or _EDITOR_ environment variables, or fall back to `nano`.

Lets edit an existing running MongoDB object to setup [Scheduled Backup](/docs/guides/mongodb/snapshot/scheduled-backup.md). The following command will open MongoDB `etcdb-demo` in editor.

```console
$ kubedb edit etcd etcdb-demo

# Add following under Spec to configure periodic backups
# backupSchedule:
#   cronExpression: '@every 1m'
#   storageSecretName: gcs-secret
#   gcs:
#     bucket: bucket-name

etcd "mongodb-demo" edited
```

#### Edit Restrictions

Various fields of a KubeDb object can't be edited using `edit` command. The following fields are restricted from updates for all KubeDB objects:

- apiVersion
- kind
- metadata.name
- metadata.namespace
- status

If StatefulSets exists for a MongoDB database, following fields can't be modified as well.

- spec.version
- spec.databaseSecret
- spec.storage
- spec.nodeSelector
- spec.init

For DormantDatabase, `spec.origin` can't be edited using `kubedb edit`

To learn about various options of `edit` command, please visit [here](/docs/reference/kubedb_edit.md).

### How to Delete Objects

`kubedb delete` command will delete an object in `default` namespace by default unless namespace is provided. The following command will delete a MongoDB `mongodb-dev` in default namespace

```console
$ kubedb delete mongodb mongodb-dev
mongodb "mongodb-dev" deleted
```

You can also use YAML files to delete objects. The following command will delete a mongodb using the type and name specified in `mongodb.yaml`.

```console
$ kubedb delete -f mongodb-demo.yaml
mongodb "mongodb-dev" deleted
```

`kubedb delete` command also takes input from `stdin`.

```console
cat mongodb-demo.yaml | kubedb delete -f -
```

To delete database with matching labels, use `--selector` flag. The following command will delete mongodb with label `mongodb.kubedb.com/name=mongodb-demo`.

```console
$ kubedb delete mongodb -l mongodb.kubedb.com/name=mongodb-demo
```

To learn about various options of `delete` command, please visit [here](/docs/reference/kubedb_delete.md).

## Using Kubectl

You can use Kubectl with KubeDB objects like any other CRDs. Below are some common examples of using Kubectl with KubeDB objects.

```console
# List objects
$ kubectl get mongodb
$ kubectl get mongodb.kubedb.com

# Delete objects
$ kubectl delete mongodb <name>
```

## Next Steps

- Learn how to use KubeDB to run a MongoDB database [here](/docs/guides/mongodb/README.md).
- Wondering what features are coming next? Please visit [here](/docs/roadmap.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
