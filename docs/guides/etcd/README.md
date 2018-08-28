---
title: Etcd
menu:
  docs_0.8.0:
    identifier: etcd-readme
    name: Etcd
    parent: etcd-guides
    weight: 10
menu_name: docs_0.8.0
section_menu_id: guides
url: /docs/0.8.0/guides/etcd/
aliases:
  - /docs/0.8.0/guides/etcd/README/
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

## Supported Etcd Features

|Features                                     | Availability |
|---------------------------------------------|:------------:|
|Clustering                                   | &#10003;     |
|Persistent Volume                            | &#10003;     |
|Instant Backup                               | &#10003;     |
|Scheduled Backup                             | &#10003;     |
|Builtin Prometheus Discovery                 | &#10003;     |
|Using CoreOS Prometheus Operator             | &#10003;     |

<br/>

## Life Cycle of a Etcd Object

<p align="center">
  <ietcd alt="lifecycle"  src="docs/images/etcd/etcd-lifecycle.png" width="600" height="660">
</p>

<br/>

## Supported Etcd Versions

| KubeDB Version | Etcd:3.2 | Etcd:3.3 |
|:--------------:|:---------:|:---------:|
| 0.8.0          | &#10003;  | &#10003;  |

## User Guide

- [Quickstart Etcd](/docs/guides/etcd/quickstart/quickstart.md) with KubeDB Operator.
- [Snapshot and Restore](/docs/guides/etcd/snapshot/backup-and-restore.md) process of Etcd databases using KubeDB.
- Take [Scheduled Snapshot](/docs/guides/etcd/snapshot/scheduled-backup.md) of Etcd databases using KubeDB.
- Initialize [Etcd with Script](/docs/guides/etcd/initialization/using-script.md).
- Initialize [Etcd with Snapshot](/docs/guides/etcd/initialization/using-snapshot.md).
- Monitor your Etcd database with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/etcd/monitoring/using-coreos-prometheus-operator.md).
- Monitor your Etcd database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/etcd/private-registry/using-private-registry.md) to deploy Etcd with KubeDB.
- Use [kubedb cli](/docs/guides/etcd/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Etcd object](/docs/concepts/databases/etcd.md).
- Detail concepts of [Snapshot object](/docs/concepts/snapshot.md).
- Wondering what features are coming next? Please visit [here](/docs/roadmap.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
