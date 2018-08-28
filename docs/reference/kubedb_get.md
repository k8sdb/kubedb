---
title: Kubedb Get
menu:
  docs_0.8.0:
    identifier: kubedb-get
    name: Kubedb Get
    parent: reference
menu_name: docs_0.8.0
section_menu_id: reference
---
## kubedb get

Display one or many resources

### Synopsis

Display one or many resources. 

Valid resource types include: 

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
kubedb get [flags]
```

### Examples

```
  # List all elasticsearch in ps output format.
  kubedb get es
  
  # List all elasticsearch in ps output format with more information (such as version).
  kubedb get elasticsearches -o wide
  
  # List a single postgres with specified NAME in ps output format.
  kubedb get postgres database
  
  # List a single snapshot in JSON output format.
  kubedb get -o json snapshot snapshot-xyz
  
  # List all postgreses and elastics together in ps output format.
  kubedb get postgreses,elastics
  
  # List one or more resources by their type and names.
  kubedb get es/es-db postgres/pg-db
```

### Options

```
      --all-namespaces     If present, list the requested object(s) across all namespaces. Namespace specified with --namespace will be ignored.
  -h, --help               help for get
  -n, --namespace string   List the requested object(s) from this namespace. (default "default")
  -o, --output string      Output format. One of: json|yaml|wide|name.
  -l, --selector string    Selector (label query) to filter on, supports '=', '==', and '!='.
  -a, --show-all           When printing, show all resources (default hide terminated pods.)
      --show-kind          If present, list the resource type for the requested object(s).
      --show-labels        When printing, show all labels as the last column (default hide labels column)
```

### Options inherited from parent commands

```
      --analytics             Send analytical events to Google Analytics (default true)
      --kube-context string   name of the kubeconfig context to use
```

### SEE ALSO

* [kubedb](/docs/reference/kubedb.md)	 - Command line interface for KubeDB


