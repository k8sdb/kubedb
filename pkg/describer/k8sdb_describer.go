package describer

import (
	"fmt"
	"io"
	"strings"

	"github.com/appscode/go/types"
	mona "github.com/appscode/kube-mon/api"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/cli/pkg/printer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"
	kapi "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/printers"
)

func (d *humanReadableDescriber) describeElasticsearch(item *api.Elasticsearch, describerSettings *printer.DescriberSettings) (string, error) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return "", err
	}

	labelSelector := labels.SelectorFromSet(
		map[string]string{
			api.LabelDatabaseKind: item.ResourceKind(),
			api.LabelDatabaseName: item.Name,
		},
	)

	snapshots, err := d.extensionsClient.Snapshots(item.Namespace).List(
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
	if err != nil {
		return "", err
	}

	var events *kapi.EventList
	if describerSettings.ShowEvents {
		item.Kind = api.ResourceKindElasticsearch
		events, err = clientSet.Core().Events(item.Namespace).Search(scheme.Scheme, item)
		if err != nil {
			return "", err
		}
	}

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "Name:\t%s\n", item.Name)
		fmt.Fprintf(out, "Namespace:\t%s\n", item.Namespace)
		fmt.Fprintf(out, "CreationTimestamp:\t%s\n", timeToString(&item.CreationTimestamp))
		if item.Labels != nil {
			printLabelsMultiline(out, "Labels", item.Labels)
		}
		fmt.Fprintf(out, "Status:\t%s\n", string(item.Status.Phase))
		if len(item.Status.Reason) > 0 {
			fmt.Fprintf(out, "Reason:\t%s\n", item.Status.Reason)
		}

		if item.Spec.Replicas != nil {
			fmt.Fprintf(out, "Replicas:\t%d  total\n", types.Int32(item.Spec.Replicas))
		}

		if item.Annotations != nil {
			printLabelsMultiline(out, "Annotations", item.Annotations)
		}

		describeInitialization(item.Spec.Init, out)

		describeStorage(item.Spec.Storage, out)

		d.showWorkload(item.Namespace, labelSelector, describerSettings.ShowWorkload, out)

		secretVolumes := make(map[string]*core.SecretVolumeSource)
		if item.Spec.DatabaseSecret != nil {
			secretVolumes["Database"] = item.Spec.DatabaseSecret
		}
		if item.Spec.CertificateSecret != nil {
			secretVolumes["Certificate"] = item.Spec.CertificateSecret
		}
		d.showSecret(item.Namespace, secretVolumes, describerSettings.ShowSecret, out)

		specific := map[string]labels.Selector{
			"master": labels.SelectorFromSet(map[string]string{"node.role.master": "set"}),
			"client": labels.SelectorFromSet(map[string]string{"node.role.client": "set"}),
			"data":   labels.SelectorFromSet(map[string]string{"node.role.data": "set"}),
		}
		d.showTopology(item.Namespace, labelSelector, specific, out)

		if item.Spec.Monitor != nil {
			describeMonitor(item.Spec.Monitor, out)
		}

		listSnapshots(snapshots, out)

		if events != nil {
			describeEvents(events, out)
		}

		return nil
	})
}

func (d *humanReadableDescriber) describePostgres(item *api.Postgres, describerSettings *printer.DescriberSettings) (string, error) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return "", err
	}

	labelSelector := labels.SelectorFromSet(
		map[string]string{
			api.LabelDatabaseKind: item.ResourceKind(),
			api.LabelDatabaseName: item.Name,
		},
	)

	snapshots, err := d.extensionsClient.Snapshots(item.Namespace).List(
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
	if err != nil {
		return "", err
	}

	var events *kapi.EventList
	if describerSettings.ShowEvents {
		item.Kind = api.ResourceKindPostgres
		events, err = clientSet.Core().Events(item.Namespace).Search(scheme.Scheme, item)
		if err != nil {
			return "", err
		}
	}

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "Name:\t%s\n", item.Name)
		fmt.Fprintf(out, "Namespace:\t%s\n", item.Namespace)
		fmt.Fprintf(out, "StartTimestamp:\t%s\n", timeToString(&item.CreationTimestamp))
		if item.Labels != nil {
			printLabelsMultiline(out, "Labels", item.Labels)
		}
		if item.Annotations != nil {
			printLabelsMultiline(out, "Annotations", item.Annotations)
		}
		if item.Spec.Replicas != nil {
			fmt.Fprintf(out, "Replicas:\t%d  total\n", types.Int32(item.Spec.Replicas))
		}
		fmt.Fprintf(out, "Status:\t%s\n", string(item.Status.Phase))
		if len(item.Status.Reason) > 0 {
			fmt.Fprintf(out, "Reason:\t%s\n", item.Status.Reason)
		}

		describeArchiver(item.Spec.Archiver, out)

		describeInitialization(item.Spec.Init, out)

		describeStorage(&item.Spec.Storage, out)

		d.showWorkload(item.Namespace, labelSelector, describerSettings.ShowWorkload, out)

		secretVolumes := make(map[string]*core.SecretVolumeSource)
		if item.Spec.DatabaseSecret != nil {
			secretVolumes["Database"] = item.Spec.DatabaseSecret
		}
		d.showSecret(item.Namespace, secretVolumes, describerSettings.ShowSecret, out)

		specific := map[string]labels.Selector{
			"primary": labels.SelectorFromSet(map[string]string{"kubedb.com/role": "primary"}),
			"replica": labels.SelectorFromSet(map[string]string{"kubedb.com/role": "replica"}),
		}
		d.showTopology(item.Namespace, labelSelector, specific, out)

		if item.Spec.Monitor != nil {
			describeMonitor(item.Spec.Monitor, out)
		}

		listSnapshots(snapshots, out)

		if events != nil {
			describeEvents(events, out)
		}

		return nil
	})
}

func (d *humanReadableDescriber) describeMySQL(item *api.MySQL, describerSettings *printer.DescriberSettings) (string, error) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return "", err
	}

	labelSelector := labels.SelectorFromSet(
		map[string]string{
			api.LabelDatabaseKind: item.ResourceKind(),
			api.LabelDatabaseName: item.Name,
		},
	)

	snapshots, err := d.extensionsClient.Snapshots(item.Namespace).List(
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
	if err != nil {
		return "", err
	}

	var events *kapi.EventList
	if describerSettings.ShowEvents {
		item.Kind = api.ResourceKindMySQL
		events, err = clientSet.Core().Events(item.Namespace).Search(scheme.Scheme, item)
		if err != nil {
			return "", err
		}
	}

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "Name:\t%s\n", item.Name)
		fmt.Fprintf(out, "Namespace:\t%s\n", item.Namespace)
		fmt.Fprintf(out, "StartTimestamp:\t%s\n", timeToString(&item.CreationTimestamp))
		if item.Labels != nil {
			printLabelsMultiline(out, "Labels", item.Labels)
		}
		if item.Annotations != nil {
			printLabelsMultiline(out, "Annotations", item.Annotations)
		}
		if item.Spec.Replicas != nil {
			fmt.Fprintf(out, "Replicas:\t%d  total\n", types.Int32(item.Spec.Replicas))
		}
		fmt.Fprintf(out, "Status:\t%s\n", string(item.Status.Phase))
		if len(item.Status.Reason) > 0 {
			fmt.Fprintf(out, "Reason:\t%s\n", item.Status.Reason)
		}

		describeStorage(&item.Spec.Storage, out)

		d.showWorkload(item.Namespace, labelSelector, describerSettings.ShowWorkload, out)

		secretVolumes := make(map[string]*core.SecretVolumeSource)
		if item.Spec.DatabaseSecret != nil {
			secretVolumes["Database"] = item.Spec.DatabaseSecret
		}
		d.showSecret(item.Namespace, secretVolumes, describerSettings.ShowSecret, out)

		if item.Spec.Monitor != nil {
			describeMonitor(item.Spec.Monitor, out)
		}

		listSnapshots(snapshots, out)

		if events != nil {
			describeEvents(events, out)
		}

		return nil
	})
}

func (d *humanReadableDescriber) describeMongoDB(item *api.MongoDB, describerSettings *printer.DescriberSettings) (string, error) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return "", err
	}

	labelSelector := labels.SelectorFromSet(
		map[string]string{
			api.LabelDatabaseKind: item.ResourceKind(),
			api.LabelDatabaseName: item.Name,
		},
	)

	snapshots, err := d.extensionsClient.Snapshots(item.Namespace).List(
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
	if err != nil {
		return "", err
	}

	var events *kapi.EventList
	if describerSettings.ShowEvents {
		item.Kind = api.ResourceKindMongoDB
		events, err = clientSet.Core().Events(item.Namespace).Search(scheme.Scheme, item)
		if err != nil {
			return "", err
		}
	}

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "Name:\t%s\n", item.Name)
		fmt.Fprintf(out, "Namespace:\t%s\n", item.Namespace)
		fmt.Fprintf(out, "StartTimestamp:\t%s\n", timeToString(&item.CreationTimestamp))
		if item.Labels != nil {
			printLabelsMultiline(out, "Labels", item.Labels)
		}
		if item.Annotations != nil {
			printLabelsMultiline(out, "Annotations", item.Annotations)
		}
		if item.Spec.Replicas != nil {
			fmt.Fprintf(out, "Replicas:\t%d  total\n", types.Int32(item.Spec.Replicas))
		}
		fmt.Fprintf(out, "Status:\t%s\n", string(item.Status.Phase))
		if len(item.Status.Reason) > 0 {
			fmt.Fprintf(out, "Reason:\t%s\n", item.Status.Reason)
		}

		describeStorage(&item.Spec.Storage, out)

		d.showWorkload(item.Namespace, labelSelector, describerSettings.ShowWorkload, out)

		secretVolumes := make(map[string]*core.SecretVolumeSource)
		if item.Spec.DatabaseSecret != nil {
			secretVolumes["Database"] = item.Spec.DatabaseSecret
		}
		d.showSecret(item.Namespace, secretVolumes, describerSettings.ShowSecret, out)

		if item.Spec.Monitor != nil {
			describeMonitor(item.Spec.Monitor, out)
		}

		listSnapshots(snapshots, out)

		if events != nil {
			describeEvents(events, out)
		}

		return nil
	})
}

func (d *humanReadableDescriber) describeRedis(item *api.Redis, describerSettings *printer.DescriberSettings) (string, error) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return "", err
	}

	labelSelector := labels.SelectorFromSet(
		map[string]string{
			api.LabelDatabaseKind: item.ResourceKind(),
			api.LabelDatabaseName: item.Name,
		},
	)

	var events *kapi.EventList
	if describerSettings.ShowEvents {
		item.Kind = api.ResourceKindRedis
		events, err = clientSet.Core().Events(item.Namespace).Search(scheme.Scheme, item)
		if err != nil {
			return "", err
		}
	}

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "Name:\t%s\n", item.Name)
		fmt.Fprintf(out, "Namespace:\t%s\n", item.Namespace)
		fmt.Fprintf(out, "StartTimestamp:\t%s\n", timeToString(&item.CreationTimestamp))
		if item.Labels != nil {
			printLabelsMultiline(out, "Labels", item.Labels)
		}
		if item.Annotations != nil {
			printLabelsMultiline(out, "Annotations", item.Annotations)
		}
		if item.Spec.Replicas != nil {
			fmt.Fprintf(out, "Replicas:\t%d  total\n", types.Int32(item.Spec.Replicas))
		}
		fmt.Fprintf(out, "Status:\t%s\n", string(item.Status.Phase))
		if len(item.Status.Reason) > 0 {
			fmt.Fprintf(out, "Reason:\t%s\n", item.Status.Reason)
		}

		describeStorage(&item.Spec.Storage, out)

		d.showWorkload(item.Namespace, labelSelector, describerSettings.ShowWorkload, out)

		if item.Spec.Monitor != nil {
			describeMonitor(item.Spec.Monitor, out)
		}

		if events != nil {
			describeEvents(events, out)
		}

		return nil
	})
}

func (d *humanReadableDescriber) describeMemcached(item *api.Memcached, describerSettings *printer.DescriberSettings) (string, error) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return "", err
	}

	labelSelector := labels.SelectorFromSet(
		map[string]string{
			api.LabelDatabaseKind: item.ResourceKind(),
			api.LabelDatabaseName: item.Name,
		},
	)

	var events *kapi.EventList
	if describerSettings.ShowEvents {
		item.Kind = api.ResourceKindMemcached
		events, err = clientSet.Core().Events(item.Namespace).Search(scheme.Scheme, item)
		if err != nil {
			return "", err
		}
	}

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "Name:\t%s\n", item.Name)
		fmt.Fprintf(out, "Namespace:\t%s\n", item.Namespace)
		fmt.Fprintf(out, "StartTimestamp:\t%s\n", timeToString(&item.CreationTimestamp))
		if item.Labels != nil {
			printLabelsMultiline(out, "Labels", item.Labels)
		}
		if item.Annotations != nil {
			printLabelsMultiline(out, "Annotations", item.Annotations)
		}
		if item.Spec.Replicas != nil {
			fmt.Fprintf(out, "Replicas:\t%d  total\n", types.Int32(item.Spec.Replicas))
		}
		fmt.Fprintf(out, "Status:\t%s\n", string(item.Status.Phase))
		if len(item.Status.Reason) > 0 {
			fmt.Fprintf(out, "Reason:\t%s\n", item.Status.Reason)
		}

		d.showWorkload(item.Namespace, labelSelector, describerSettings.ShowWorkload, out)

		if item.Spec.Monitor != nil {
			describeMonitor(item.Spec.Monitor, out)
		}

		if events != nil {
			describeEvents(events, out)
		}

		return nil
	})
}

func (d *humanReadableDescriber) describeEtcd(item *api.Etcd, describerSettings *printer.DescriberSettings) (string, error) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return "", err
	}

	labelSelector := labels.SelectorFromSet(
		map[string]string{
			api.LabelDatabaseKind: item.ResourceKind(),
			api.LabelDatabaseName: item.Name,
		},
	)

	snapshots, err := d.extensionsClient.Snapshots(item.Namespace).List(
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
	if err != nil {
		return "", err
	}

	var events *kapi.EventList
	if describerSettings.ShowEvents {
		item.Kind = api.ResourceKindEtcd
		events, err = clientSet.Core().Events(item.Namespace).Search(scheme.Scheme, item)
		if err != nil {
			return "", err
		}
	}

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "Name:\t%s\n", item.Name)
		fmt.Fprintf(out, "Namespace:\t%s\n", item.Namespace)
		fmt.Fprintf(out, "StartTimestamp:\t%s\n", timeToString(&item.CreationTimestamp))
		if item.Labels != nil {
			printLabelsMultiline(out, "Labels", item.Labels)
		}
		if item.Annotations != nil {
			printLabelsMultiline(out, "Annotations", item.Annotations)
		}
		if item.Spec.Replicas != nil {
			fmt.Fprintf(out, "Replicas:\t%d  total\n", types.Int32(item.Spec.Replicas))
		}
		fmt.Fprintf(out, "Status:\t%s\n", string(item.Status.Phase))
		if len(item.Status.Reason) > 0 {
			fmt.Fprintf(out, "Reason:\t%s\n", item.Status.Reason)
		}

		describeStorage(item.Spec.Storage, out)

		d.showWorkload(item.Namespace, labelSelector, describerSettings.ShowWorkload, out)

		secretVolumes := make(map[string]*core.SecretVolumeSource)
		if item.Spec.DatabaseSecret != nil {
			secretVolumes["Database"] = item.Spec.DatabaseSecret
		}
		d.showSecret(item.Namespace, secretVolumes, describerSettings.ShowSecret, out)

		if item.Spec.Monitor != nil {
			describeMonitor(item.Spec.Monitor, out)
		}

		listSnapshots(snapshots, out)

		if events != nil {
			describeEvents(events, out)
		}

		return nil
	})
}

func (d *humanReadableDescriber) describeSnapshot(item *api.Snapshot, describerSettings *printer.DescriberSettings) (string, error) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return "", err
	}

	var events *kapi.EventList
	if describerSettings.ShowEvents {
		item.Kind = api.ResourceKindSnapshot
		events, err = clientSet.Core().Events(item.Namespace).Search(scheme.Scheme, item)
		if err != nil {
			return "", err
		}
	}

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "Name:\t%s\n", item.Name)
		fmt.Fprintf(out, "Namespace:\t%s\n", item.Namespace)
		fmt.Fprintf(out, "CreationTimestamp:\t%s\n", timeToString(&item.CreationTimestamp))
		if item.Status.CompletionTime != nil {
			fmt.Fprintf(out, "CompletionTimestamp:\t%s\n", timeToString(item.Status.CompletionTime))
		}
		if item.Labels != nil {
			printLabelsMultiline(out, "Labels", item.Labels)
		}
		if item.Annotations != nil {
			printLabelsMultiline(out, "Annotations", item.Annotations)
		}
		fmt.Fprintf(out, "Status:\t%s\n", string(item.Status.Phase))
		if len(item.Status.Reason) > 0 {
			fmt.Fprintf(out, "Reason:\t%s\n", item.Status.Reason)
		}

		fmt.Fprintln(out, "Storage:")
		describeSnapshotStorage(item.Spec.SnapshotStorageSpec, out, 2)

		secretVolumes := make(map[string]*core.SecretVolumeSource)
		if item.Spec.StorageSecretName != "" {
			secretVolumes["Database"] = &core.SecretVolumeSource{SecretName: item.Spec.StorageSecretName}
		}
		d.showSecret(item.Namespace, secretVolumes, describerSettings.ShowSecret, out)

		if events != nil {
			describeEvents(events, out)
		}

		return nil
	})
}

func (d *humanReadableDescriber) describeDormantDatabase(item *api.DormantDatabase, describerSettings *printer.DescriberSettings) (string, error) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return "", err
	}

	labelSelector := labels.SelectorFromSet(
		map[string]string{
			api.LabelDatabaseKind: item.ResourceKind(),
			api.LabelDatabaseName: item.Name,
		},
	)

	snapshots, err := d.extensionsClient.Snapshots(item.Namespace).List(
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
	if err != nil {
		return "", err
	}

	var events *kapi.EventList
	if describerSettings.ShowEvents {
		item.Kind = api.ResourceKindDormantDatabase
		events, err = clientSet.Core().Events(item.Namespace).Search(scheme.Scheme, item)
		if err != nil {
			return "", err
		}
	}

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "Name:\t%s\n", item.Name)
		fmt.Fprintf(out, "Namespace:\t%s\n", item.Namespace)
		fmt.Fprintf(out, "CreationTimestamp:\t%s\n", timeToString(&item.CreationTimestamp))
		if item.Status.PausingTime != nil {
			fmt.Fprintf(out, "PausedTimestamp:\t%s\n", timeToString(item.Status.PausingTime))
		}
		if item.Status.WipeOutTime != nil {
			fmt.Fprintf(out, "WipeOutTimestamp:\t%s\n", timeToString(item.Status.WipeOutTime))
		}
		if item.Labels != nil {
			printLabelsMultiline(out, "Labels", item.Labels)
		}
		if item.Annotations != nil {
			printLabelsMultiline(out, "Annotations", item.Annotations)
		}
		fmt.Fprintf(out, "Status:\t%s\n", string(item.Status.Phase))
		if len(item.Status.Reason) > 0 {
			fmt.Fprintf(out, "Reason:\t%s\n", item.Status.Reason)
		}

		describeOrigin(item.Spec.Origin, out)

		if item.Status.Phase != api.DormantDatabasePhaseWipedOut {
			listSnapshots(snapshots, out)
		}

		if events != nil {
			describeEvents(events, out)
		}

		return nil
	})
}

func describeStorage(pvcSpec *core.PersistentVolumeClaimSpec, out io.Writer) {
	if pvcSpec == nil {
		fmt.Fprint(out, "No volumes.\n")
		return
	}

	accessModes := getAccessModesAsString(pvcSpec.AccessModes)
	val, _ := pvcSpec.Resources.Requests[core.ResourceStorage]
	capacity := val.String()
	fmt.Fprint(out, "Volume:\n")
	if pvcSpec.StorageClassName != nil {
		fmt.Fprintf(out, "  StorageClass:\t%s\n", *pvcSpec.StorageClassName)
	}
	fmt.Fprintf(out, "  Capacity:\t%s\n", capacity)
	if accessModes != "" {
		fmt.Fprintf(out, "  Access Modes:\t%s\n", accessModes)
	}
}

func describeArchiver(archiver *api.PostgresArchiverSpec, out io.Writer) {
	if archiver == nil {
		return
	}
	fmt.Fprintln(out, "Archiver:")
	if archiver.Storage != nil {
		describeSnapshotStorage(*archiver.Storage, out, 1)
	}
}

func describeInitialization(init *api.InitSpec, out io.Writer) {
	if init == nil {
		return
	}

	fmt.Fprintln(out, "Init:")
	if init.ScriptSource != nil {
		fmt.Fprintln(out, "  scriptSource:")
		describeVolumes(init.ScriptSource.VolumeSource, out)
	}
	if init.SnapshotSource != nil {
		fmt.Fprintln(out, "  snapshotSource:")
		fmt.Fprintf(out, "    namespace:\t%s\n", init.SnapshotSource.Namespace)
		fmt.Fprintf(out, "    name:\t%s\n", init.SnapshotSource.Name)
	}
	if init.PostgresWAL != nil {
		fmt.Fprintln(out, "  postgresWAL:")
		describeSnapshotStorage(init.PostgresWAL.SnapshotStorageSpec, out, 2)
	}
}

func describeMonitor(monitor *mona.AgentSpec, out io.Writer) {
	if monitor == nil {
		return
	}

	fmt.Fprint(out, "\n")
	fmt.Fprint(out, "Monitoring System:\n")
	fmt.Fprintf(out, "  Agent:\t%s\n", monitor.Agent)
	if monitor.Prometheus != nil {
		prom := monitor.Prometheus
		fmt.Fprint(out, "  Prometheus:\n")
		if prom.Port != 0 {
			fmt.Fprintf(out, "    Port:\t%v\n", prom.Port)
		}
		if prom.Namespace != "" {
			fmt.Fprintf(out, "    Namespace:\t%s\n", prom.Namespace)
		}
		if prom.Labels != nil {
			printLabelsMultiline(out, "    Labels", prom.Labels)
		}
		if prom.Interval != "" {
			fmt.Fprintf(out, "    Interval:\t%s\n", prom.Interval)
		}

	}
}

func listSnapshots(snapshotList *api.SnapshotList, out io.Writer) {
	fmt.Fprint(out, "\n")

	if len(snapshotList.Items) == 0 {
		fmt.Fprint(out, "No Snapshots.\n")
		return
	}

	fmt.Fprint(out, "Snapshots:\n")
	w := printers.GetNewTabWriter(out)

	fmt.Fprint(w, "  Name\tBucket\tStartTime\tCompletionTime\tPhase\n")
	fmt.Fprint(w, "  ----\t------\t---------\t--------------\t-----\n")
	for _, e := range snapshotList.Items {
		location, err := e.Spec.SnapshotStorageSpec.Location()
		if err != nil {
			location = "<invalid>"
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n",
			e.Name,
			location,
			timeToString(e.Status.StartTime),
			timeToString(e.Status.CompletionTime),
			e.Status.Phase,
		)
	}
	w.Flush()
}

func describeOrigin(origin api.Origin, out io.Writer) {
	fmt.Fprint(out, "\n")
	fmt.Fprint(out, "Origin:\n")
	fmt.Fprintf(out, "  Name:\t%s\n", origin.Name)
	fmt.Fprintf(out, "  Namespace:\t%s\n", origin.Namespace)
	if origin.Labels != nil {
		printLabelsMultiline(out, "  Labels", origin.Labels)
	}
	if origin.Annotations != nil {
		printLabelsMultiline(out, "  Annotations", origin.Annotations)
	}
}

func (d *humanReadableDescriber) showWorkload(namespace string, labelSelector labels.Selector, show bool, out io.Writer) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return
	}

	statefulSets, _ := clientSet.Apps().StatefulSets(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})

	deployments, _ := clientSet.Extensions().Deployments(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})

	services, _ := clientSet.Core().Services(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})

	if show {
		if len(statefulSets.Items) > 0 {
			for _, s := range statefulSets.Items {
				d.describeStatefulSet(s, out)
			}
		}
		if len(deployments.Items) > 0 {
			for _, s := range deployments.Items {
				d.describeDeployments(s, out)
			}
		}
		if len(services.Items) > 0 {
			for _, s := range services.Items {
				d.describeService(s, out)
			}
		}
	} else {
		if len(statefulSets.Items) > 0 {
			statefulSetNames := make([]string, 0)
			for _, s := range statefulSets.Items {
				statefulSetNames = append(statefulSetNames, s.Name)
			}
			fmt.Fprintf(out, "StatefulSet:\t%s\n", strings.Join(statefulSetNames, ", "))
		}

		if len(deployments.Items) > 0 {
			deploymentNames := make([]string, 0)
			for _, s := range deployments.Items {
				deploymentNames = append(deploymentNames, s.Name)
			}
			fmt.Fprintf(out, "Deployment:\t%s\n", strings.Join(deploymentNames, ", "))
		}

		if len(services.Items) > 0 {
			serviceNames := make([]string, 0)
			for _, s := range services.Items {
				serviceNames = append(serviceNames, s.Name)
			}
			fmt.Fprintf(out, "Service:\t%s\n", strings.Join(serviceNames, ", "))
		}
	}
}

func (d *humanReadableDescriber) showSecret(namespace string, secretVolumes map[string]*core.SecretVolumeSource, show bool, out io.Writer) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return
	}

	secrets := make(map[string]*kapi.Secret)

	c := clientSet.Core().Secrets(namespace)

	for key, sv := range secretVolumes {
		secret, err := c.Get(sv.SecretName, metav1.GetOptions{})
		if err != nil {
			continue
		}
		secrets[key] = secret
	}

	if show {
		for key, s := range secrets {
			d.describeSecret(namespace, s.Name, key, out)
		}
	} else {
		secretNames := make([]string, 0)
		for _, s := range secrets {
			secretNames = append(secretNames, s.Name)
		}
		fmt.Fprintf(out, "Secrets:\t%s\n", strings.Join(secretNames, ", "))
	}
}

func (d *humanReadableDescriber) showTopology(namespace string, labelSelector labels.Selector, specific map[string]labels.Selector, out io.Writer) {
	clientSet, err := d.ClientSet()
	if err != nil {
		return
	}

	fmt.Fprint(out, "\n")

	fmt.Fprint(out, "Topology:\n")
	w := printers.GetNewTabWriter(out)

	fmt.Fprint(w, "  Type\tPod\tStartTime\tPhase\n")
	fmt.Fprint(w, "  ----\t---\t---------\t-----\n")

	pods, _ := clientSet.Core().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})

	for _, pod := range pods.Items {
		types := make([]string, 0)
		for key, val := range specific {
			if val.Matches(labels.Set(pod.Labels)) {
				types = append(types, key)
			}
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n",
			strings.Join(types, "|"),
			pod.Name,
			pod.Status.StartTime,
			pod.Status.Phase,
		)
	}

	w.Flush()
}
