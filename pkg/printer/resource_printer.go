package printer

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/golang/glog"
	tapi "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	"github.com/kubedb/cli/pkg/decoder"
	"github.com/kubedb/cli/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// ref: k8s.io/kubernetes/pkg/kubectl/resource_printer.go

const (
	tabwriterMinWidth = 10
	tabwriterWidth    = 4
	tabwriterPadding  = 3
	tabwriterPadChar  = ' '
	tabwriterFlags    = 0
	statusUnknown     = "Unknown"
)

type handlerEntry struct {
	printFunc reflect.Value
	args      []reflect.Value
}

type PrintOptions struct {
	WithNamespace bool
	WithKind      bool
	Wide          bool
	ShowAll       bool
	ShowLabels    bool
	Kind          string
}

type HumanReadablePrinter struct {
	handlerMap   map[reflect.Type]*handlerEntry
	options      PrintOptions
	lastType     reflect.Type
	hiddenObjNum int
}

func NewHumanReadablePrinter(options PrintOptions) *HumanReadablePrinter {
	printer := &HumanReadablePrinter{
		handlerMap: make(map[reflect.Type]*handlerEntry),
		options:    options,
	}
	printer.addDefaultHandlers()
	return printer
}

func shortHumanDuration(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds <= 0 {
		return fmt.Sprintf("0s")
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*364 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}

func (h *HumanReadablePrinter) addDefaultHandlers() {
	h.Handler(h.printElasticList)
	h.Handler(h.printElastic)
	h.Handler(h.printPostgresList)
	h.Handler(h.printPostgres)
	h.Handler(h.printMySQLList)
	h.Handler(h.printMySQL)
	h.Handler(h.printMongoDBList)
	h.Handler(h.printMongoDB)
	h.Handler(h.printRedisList)
	h.Handler(h.printRedis)
	h.Handler(h.printMemcachedList)
	h.Handler(h.printMemcached)
	h.Handler(h.printSnapshotList)
	h.Handler(h.printSnapshot)
	h.Handler(h.printDormantDatabaseList)
	h.Handler(h.printDormantDatabase)
	h.Handler(h.printEtcd)
	h.Handler(h.printEtcdList)
}

func (h *HumanReadablePrinter) Handler(printFunc interface{}) error {
	printFuncValue := reflect.ValueOf(printFunc)
	if err := h.validatePrintHandlerFunc(printFuncValue); err != nil {
		glog.Errorf("Unable to add print handler: %v", err)
		return err
	}

	objType := printFuncValue.Type().In(0)

	h.handlerMap[objType] = &handlerEntry{
		printFunc: printFuncValue,
	}
	return nil
}

func (h *HumanReadablePrinter) validatePrintHandlerFunc(printFunc reflect.Value) error {
	if printFunc.Kind() != reflect.Func {
		return fmt.Errorf("invalid print handler. %#v is not a function", printFunc)
	}
	funcType := printFunc.Type()
	if funcType.NumIn() != 3 || funcType.NumOut() != 1 {
		return fmt.Errorf("invalid print handler." +
			"Must accept 3 parameters and return 1 value.")
	}
	if funcType.In(1) != reflect.TypeOf((*io.Writer)(nil)).Elem() ||
		funcType.In(2) != reflect.TypeOf((*PrintOptions)(nil)).Elem() ||
		funcType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("invalid print handler. The expected signature is: "+
			"func handler(obj %v, w io.Writer, options PrintOptions) error", funcType.In(0))
	}
	return nil
}

func getColumns(options PrintOptions, t reflect.Type) []string {

	columns := make([]string, 0)

	if options.WithNamespace {
		columns = append(columns, "NAMESPACE")
	}

	columns = append(columns, "NAME")

	// Must be kept in sync with apiversion
	switch t.String() {
	case "*v1alpha1.Elasticsearch", "*v1alpha1.ElasticList":
		if options.Wide {
			columns = append(columns, "VERSION")
		}
	case "*v1alpha1.Postgres", "*v1alpha1.PostgresList":
		if options.Wide {
			columns = append(columns, "VERSION")
		}
	case "*v1alpha1.MySQL", "*v1alpha1.MySQLList":
		if options.Wide {
			columns = append(columns, "VERSION")
		}
	case "*v1alpha1.MongoDB", "*v1alpha1.MongoDBList":
		if options.Wide {
			columns = append(columns, "VERSION")
		}
	case "*v1alpha1.Redis", "*v1alpha1.RedisList":
		if options.Wide {
			columns = append(columns, "VERSION")
		}
	case "*v1alpha1.Snapshot", "*v1alpha1.SnapshotList":
		columns = append(columns, "DATABASE")
		if options.Wide {
			columns = append(columns, "BUCKET")
		}
	case "*v1alpha1.Etcd", "*v1alpha1.EtcdList":
		if options.Wide {
			columns = append(columns, "VERSION")
		}
	}

	columns = append(columns, "STATUS")
	columns = append(columns, "AGE")
	if options.ShowLabels {
		columns = append(columns, "LABELS")
	}
	return columns
}

func (h *HumanReadablePrinter) printElastic(item *tapi.Elasticsearch, w io.Writer, options PrintOptions) error {
	name := formatResourceName(options.Kind, item.Name, options.WithKind)

	namespace := item.Namespace

	if options.WithNamespace {
		if _, err := fmt.Fprintf(w, "%s\t", namespace); err != nil {
			return err
		}
	}

	status := item.Status.Phase
	if status == "" {
		status = statusUnknown
	}
	if _, err := fmt.Fprintf(w, "%s\t", name); err != nil {
		return err
	}

	if options.Wide {
		if _, err := fmt.Fprintf(w, "%s\t", item.Spec.Version); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s\t%s", status, TranslateTimestamp(item.CreationTimestamp)); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, appendAllLabels(options.ShowLabels, item.Labels))

	return err
}

func (h *HumanReadablePrinter) printElasticList(itemList *tapi.ElasticsearchList, w io.Writer, options PrintOptions) error {
	for _, item := range itemList.Items {
		if err := h.printElastic(&item, w, options); err != nil {
			return err
		}
	}
	return nil
}

func (h *HumanReadablePrinter) printPostgres(item *tapi.Postgres, w io.Writer, options PrintOptions) error {
	name := formatResourceName(options.Kind, item.Name, options.WithKind)

	namespace := item.Namespace

	if options.WithNamespace {
		if _, err := fmt.Fprintf(w, "%s\t", namespace); err != nil {
			return err
		}
	}

	status := item.Status.Phase
	if status == "" {
		status = statusUnknown
	}
	if _, err := fmt.Fprintf(w, "%s\t", name); err != nil {
		return err
	}

	if options.Wide {
		if _, err := fmt.Fprintf(w, "%s\t", item.Spec.Version); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s\t%s", status, TranslateTimestamp(item.CreationTimestamp)); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, appendAllLabels(options.ShowLabels, item.Labels))

	return err
}

func (h *HumanReadablePrinter) printPostgresList(itemList *tapi.PostgresList, w io.Writer, options PrintOptions) error {
	for _, item := range itemList.Items {
		if err := h.printPostgres(&item, w, options); err != nil {
			return err
		}
	}
	return nil
}

func (h *HumanReadablePrinter) printMySQL(item *tapi.MySQL, w io.Writer, options PrintOptions) error {
	name := formatResourceName(options.Kind, item.Name, options.WithKind)

	namespace := item.Namespace

	if options.WithNamespace {
		if _, err := fmt.Fprintf(w, "%s\t", namespace); err != nil {
			return err
		}
	}

	status := item.Status.Phase
	if status == "" {
		status = statusUnknown
	}
	if _, err := fmt.Fprintf(w, "%s\t", name); err != nil {
		return err
	}

	if options.Wide {
		if _, err := fmt.Fprintf(w, "%s\t", item.Spec.Version); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s\t%s", status, TranslateTimestamp(item.CreationTimestamp)); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, appendAllLabels(options.ShowLabels, item.Labels))

	return err
}

func (h *HumanReadablePrinter) printMySQLList(itemList *tapi.MySQLList, w io.Writer, options PrintOptions) error {
	for _, item := range itemList.Items {
		if err := h.printMySQL(&item, w, options); err != nil {
			return err
		}
	}
	return nil
}

func (h *HumanReadablePrinter) printMongoDB(item *tapi.MongoDB, w io.Writer, options PrintOptions) error {
	name := formatResourceName(options.Kind, item.Name, options.WithKind)

	namespace := item.Namespace

	if options.WithNamespace {
		if _, err := fmt.Fprintf(w, "%s\t", namespace); err != nil {
			return err
		}
	}

	status := item.Status.Phase
	if status == "" {
		status = statusUnknown
	}
	if _, err := fmt.Fprintf(w, "%s\t", name); err != nil {
		return err
	}

	if options.Wide {
		if _, err := fmt.Fprintf(w, "%s\t", item.Spec.Version); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s\t%s", status, TranslateTimestamp(item.CreationTimestamp)); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, appendAllLabels(options.ShowLabels, item.Labels))

	return err
}

func (h *HumanReadablePrinter) printMongoDBList(itemList *tapi.MongoDBList, w io.Writer, options PrintOptions) error {
	for _, item := range itemList.Items {
		if err := h.printMongoDB(&item, w, options); err != nil {
			return err
		}
	}
	return nil
}

func (h *HumanReadablePrinter) printEtcd(item *tapi.Etcd, w io.Writer, options PrintOptions) error {
	name := formatResourceName(options.Kind, item.Name, options.WithKind)

	namespace := item.Namespace

	if options.WithNamespace {
		if _, err := fmt.Fprintf(w, "%s\t", namespace); err != nil {
			return err
		}
	}

	status := item.Status.Phase
	if status == "" {
		status = statusUnknown
	}
	if _, err := fmt.Fprintf(w, "%s\t", name); err != nil {
		return err
	}

	if options.Wide {
		if _, err := fmt.Fprintf(w, "%s\t", item.Spec.Version); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s\t%s", status, TranslateTimestamp(item.CreationTimestamp)); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, appendAllLabels(options.ShowLabels, item.Labels))

	return err
}

func (h *HumanReadablePrinter) printEtcdList(itemList *tapi.EtcdList, w io.Writer, options PrintOptions) error {
	for _, item := range itemList.Items {
		if err := h.printEtcd(&item, w, options); err != nil {
			return err
		}
	}
	return nil
}
func (h *HumanReadablePrinter) printRedis(item *tapi.Redis, w io.Writer, options PrintOptions) error {
	name := formatResourceName(options.Kind, item.Name, options.WithKind)

	namespace := item.Namespace

	if options.WithNamespace {
		if _, err := fmt.Fprintf(w, "%s\t", namespace); err != nil {
			return err
		}
	}

	status := item.Status.Phase
	if status == "" {
		status = statusUnknown
	}
	if _, err := fmt.Fprintf(w, "%s\t", name); err != nil {
		return err
	}

	if options.Wide {
		if _, err := fmt.Fprintf(w, "%s\t", item.Spec.Version); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s\t%s", status, TranslateTimestamp(item.CreationTimestamp)); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, appendAllLabels(options.ShowLabels, item.Labels))

	return err
}

func (h *HumanReadablePrinter) printRedisList(itemList *tapi.RedisList, w io.Writer, options PrintOptions) error {
	for _, item := range itemList.Items {
		if err := h.printRedis(&item, w, options); err != nil {
			return err
		}
	}
	return nil
}

func (h *HumanReadablePrinter) printMemcached(item *tapi.Memcached, w io.Writer, options PrintOptions) error {
	name := formatResourceName(options.Kind, item.Name, options.WithKind)

	namespace := item.Namespace

	if options.WithNamespace {
		if _, err := fmt.Fprintf(w, "%s\t", namespace); err != nil {
			return err
		}
	}

	status := item.Status.Phase
	if status == "" {
		status = statusUnknown
	}
	if _, err := fmt.Fprintf(w, "%s\t", name); err != nil {
		return err
	}

	if options.Wide {
		if _, err := fmt.Fprintf(w, "%s\t", item.Spec.Version); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s\t%s", status, TranslateTimestamp(item.CreationTimestamp)); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, appendAllLabels(options.ShowLabels, item.Labels))

	return err
}

func (h *HumanReadablePrinter) printMemcachedList(itemList *tapi.MemcachedList, w io.Writer, options PrintOptions) error {
	for _, item := range itemList.Items {
		if err := h.printMemcached(&item, w, options); err != nil {
			return err
		}
	}
	return nil
}

func (h *HumanReadablePrinter) printSnapshot(item *tapi.Snapshot, w io.Writer, options PrintOptions) error {
	name := formatResourceName(options.Kind, item.Name, options.WithKind)

	namespace := item.Namespace

	if options.WithNamespace {
		if _, err := fmt.Fprintf(w, "%s\t", namespace); err != nil {
			return err
		}
	}

	status := string(item.Status.Phase)
	if status == "" {
		status = statusUnknown
	}

	short, found := util.ResourceShortFormFor(item.Labels[tapi.LabelDatabaseKind])
	database := fmt.Sprintf(`%v/%v`, short, item.Spec.DatabaseName)
	if !found {
		database = fmt.Sprintf(`%v`, item.Spec.DatabaseName)
	}

	if _, err := fmt.Fprintf(w, "%s\t%s\t", name, database); err != nil {
		return err
	}

	if options.Wide {
		loc, err := item.Spec.SnapshotStorageSpec.Location()
		if err != nil {
			loc = statusUnknown
		}
		if _, err := fmt.Fprintf(w, "%s\t", loc); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s\t%s", status, TranslateTimestamp(item.CreationTimestamp)); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, appendAllLabels(options.ShowLabels, item.Labels))

	return err
}

func (h *HumanReadablePrinter) printSnapshotList(itemList *tapi.SnapshotList, w io.Writer, options PrintOptions) error {
	for _, item := range itemList.Items {
		if err := h.printSnapshot(&item, w, options); err != nil {
			return err
		}
	}
	return nil
}

func (h *HumanReadablePrinter) printDormantDatabase(item *tapi.DormantDatabase, w io.Writer, options PrintOptions) error {
	name := formatResourceName(options.Kind, item.Name, options.WithKind)

	namespace := item.Namespace

	if options.WithNamespace {
		if _, err := fmt.Fprintf(w, "%s\t", namespace); err != nil {
			return err
		}
	}

	phase := string(item.Status.Phase)
	if phase == "" {
		phase = statusUnknown
	}
	if _, err := fmt.Fprintf(w, "%s\t%s\t", name, phase); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "%s", TranslateTimestamp(item.CreationTimestamp)); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, appendAllLabels(options.ShowLabels, item.Labels))

	return err
}

func (h *HumanReadablePrinter) printDormantDatabaseList(itemList *tapi.DormantDatabaseList, w io.Writer, options PrintOptions) error {
	for _, item := range itemList.Items {
		if err := h.printDormantDatabase(&item, w, options); err != nil {
			return err
		}
	}
	return nil
}

func (h *HumanReadablePrinter) PrintObj(obj runtime.Object, output io.Writer) error {
	w, found := output.(*tabwriter.Writer)
	if !found {
		w = GetNewTabWriter(output)
		defer w.Flush()
	}

	kind := obj.GetObjectKind().GroupVersionKind().Kind
	codec := scheme.Codecs.LegacyCodec(tapi.SchemeGroupVersion)

	switch obj.(type) {
	case *unstructured.UnstructuredList, *unstructured.Unstructured, *runtime.Unknown:
		if objBytes, err := runtime.Encode(codec, obj); err == nil {
			if decodedObj, err := decoder.Decode(kind, objBytes); err == nil {
				obj = decodedObj
			}
		}
	}

	t := reflect.TypeOf(obj)
	if handler := h.handlerMap[t]; handler != nil {
		if t != h.lastType {
			headers := getColumns(h.options, t)
			if h.lastType != nil {
				printNewline(w)
			}
			h.printHeader(headers, w)
			h.lastType = t
		}
		args := []reflect.Value{reflect.ValueOf(obj), reflect.ValueOf(w), reflect.ValueOf(h.options)}
		resultValue := handler.printFunc.Call(args)[0]
		if resultValue.IsNil() {
			return nil
		}
		return resultValue.Interface().(error)
	}

	return fmt.Errorf(`kubedb doesn't support: "%v"`, kind)
}

func (h *HumanReadablePrinter) HandledResources() []string {
	return []string{}
}

func (h *HumanReadablePrinter) AfterPrint(io.Writer, string) error {
	return nil
}

func (h *HumanReadablePrinter) printHeader(columnNames []string, w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t")); err != nil {
		return err
	}
	return nil
}

func (h *HumanReadablePrinter) IsGeneric() bool {
	return false
}

func printNewline(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "\n"); err != nil {
		return err
	}
	return nil
}

func formatResourceName(kind, name string, withKind bool) string {
	if !withKind || kind == "" {
		return name
	}

	return kind + "/" + name
}

func (h *HumanReadablePrinter) GetResourceKind() string {
	return h.options.Kind
}

func (h *HumanReadablePrinter) EnsurePrintWithKind(kind string) {
	h.options.WithKind = true
	h.options.Kind = kind
}

func appendAllLabels(showLabels bool, itemLabels map[string]string) string {
	var buffer bytes.Buffer

	if showLabels {
		buffer.WriteString(fmt.Sprint("\t"))
		buffer.WriteString(labels.FormatLabels(itemLabels))
	}
	buffer.WriteString("\n")

	return buffer.String()
}

func TranslateTimestamp(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	return shortHumanDuration(time.Now().Sub(timestamp.Time))
}

func GetNewTabWriter(output io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(
		output,
		tabwriterMinWidth,
		tabwriterWidth,
		tabwriterPadding,
		tabwriterPadChar,
		tabwriterFlags,
	)
}
