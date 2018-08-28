---
title: Kubedb Describe
menu:
  docs_0.8.0:
    identifier: kubedb-describe
    name: Kubedb Describe
    parent: reference
menu_name: docs_0.8.0
section_menu_id: reference
---
## kubedb describe

Show details of a specific resource or group of resources

### Synopsis

Show details of a specific resource or group of resources. This command joins many API calls together to form a detailed description of a given resource or group of resources.Valid resource types include: 

  * all  
  * elasticsearches  
  * postgreses  
  * mysqls  
  * mongodbs  
  * redises  
  * memcacheds  
  * snapshots  
  * dormantdatabases
  * etcds

```
kubedb describe (TYPE [NAME_PREFIX] | TYPE/NAME) [flags]
```

### Examples

```
  # Describe a elasticsearch
  kubedb describe elasticsearches elasticsearch-demo
  
  # Describe a postgres
  kubedb describe pg/postgres-demo
  
  # Describe all dormantdatabases
  kubedb describe drmn
```

### Options

```
      --all-namespaces     If present, describe the requested object(s) across all namespaces. Namespace specified with --namespace will be ignored.
  -h, --help               help for describe
  -n, --namespace string   Describe object(s) from this namespace. (default "default")
  -l, --selector string    Selector (label query) to filter on, supports '=', '==', and '!='.
  -E, --show-event         If true, display events related to the described object. (default true)
  -S, --show-secret        If true, display secrets. (default true)
  -W, --show-workload      If true, describe statefulSet, service and secrets. (default true)
```

### Options inherited from parent commands

```
      --analytics             Send analytical events to Google Analytics (default true)
      --kube-context string   name of the kubeconfig context to use
```

### SEE ALSO

* [kubedb](/docs/reference/kubedb.md)	 - Command line interface for KubeDB


