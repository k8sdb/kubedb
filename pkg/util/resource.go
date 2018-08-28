package util

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	tapi "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/cli/pkg/decoder"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func GetSupportedResource(resource string) (string, error) {
	switch strings.ToLower(resource) {
	case strings.ToLower(tapi.ResourceKindElasticsearch),
		strings.ToLower(tapi.ResourcePluralElasticsearch),
		strings.ToLower(tapi.ResourceCodeElasticsearch),
		strings.ToLower(tapi.ResourceSingularElasticsearch):
		return tapi.ResourcePluralElasticsearch + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindPostgres),
		strings.ToLower(tapi.ResourcePluralPostgres),
		strings.ToLower(tapi.ResourceCodePostgres),
		strings.ToLower(tapi.ResourceSingularPostgres):
		return tapi.ResourcePluralPostgres + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindMySQL),
		strings.ToLower(tapi.ResourcePluralMySQL),
		strings.ToLower(tapi.ResourceCodeMySQL),
		strings.ToLower(tapi.ResourceSingularMySQL):
		return tapi.ResourcePluralMySQL + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindMongoDB),
		strings.ToLower(tapi.ResourcePluralMongoDB),
		strings.ToLower(tapi.ResourceCodeMongoDB),
		strings.ToLower(tapi.ResourceSingularMongoDB):
		return tapi.ResourcePluralMongoDB + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindRedis),
		strings.ToLower(tapi.ResourcePluralRedis),
		strings.ToLower(tapi.ResourceCodeRedis),
		strings.ToLower(tapi.ResourceSingularRedis):
		return tapi.ResourcePluralRedis + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindMemcached),
		strings.ToLower(tapi.ResourcePluralMemcached),
		strings.ToLower(tapi.ResourceCodeMemcached),
		strings.ToLower(tapi.ResourceSingularMemcached):
		return tapi.ResourcePluralMemcached + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindSnapshot),
		strings.ToLower(tapi.ResourcePluralSnapshot),
		strings.ToLower(tapi.ResourceCodeSnapshot),
		strings.ToLower(tapi.ResourceSingularSnapshot):
		return tapi.ResourcePluralSnapshot + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindDormantDatabase),
		strings.ToLower(tapi.ResourcePluralDormantDatabase),
		strings.ToLower(tapi.ResourceCodeDormantDatabase),
		strings.ToLower(tapi.ResourceSingularDormantDatabase):
		return tapi.ResourcePluralDormantDatabase + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindEtcd),
		strings.ToLower(tapi.ResourcePluralEtcd),
		strings.ToLower(tapi.ResourceCodeEtcd),
		strings.ToLower(tapi.ResourceSingularEtcd):
		return tapi.ResourcePluralEtcd + "." + tapi.SchemeGroupVersion.Group, nil
	default:
		return "", fmt.Errorf(`kubedb doesn't support a resource type "%v"`, resource)
	}
}

func GetResourceType(resource string) (string, error) {
	switch strings.ToLower(resource) {
	case strings.ToLower(tapi.ResourceKindElasticsearch),
		strings.ToLower(tapi.ResourcePluralElasticsearch),
		strings.ToLower(tapi.ResourceCodeElasticsearch),
		strings.ToLower(tapi.ResourceSingularElasticsearch):
		return tapi.ResourcePluralElasticsearch, nil
	case strings.ToLower(tapi.ResourceKindPostgres),
		strings.ToLower(tapi.ResourcePluralPostgres),
		strings.ToLower(tapi.ResourceCodePostgres),
		strings.ToLower(tapi.ResourceSingularPostgres):
		return tapi.ResourcePluralPostgres, nil
	case strings.ToLower(tapi.ResourceKindMySQL),
		strings.ToLower(tapi.ResourcePluralMySQL),
		strings.ToLower(tapi.ResourceCodeMySQL),
		strings.ToLower(tapi.ResourceSingularMySQL):
		return tapi.ResourcePluralMySQL, nil
	case strings.ToLower(tapi.ResourceKindMongoDB),
		strings.ToLower(tapi.ResourcePluralMongoDB),
		strings.ToLower(tapi.ResourceCodeMongoDB),
		strings.ToLower(tapi.ResourceSingularMongoDB):
		return tapi.ResourcePluralMongoDB, nil
	case strings.ToLower(tapi.ResourceKindRedis),
		strings.ToLower(tapi.ResourcePluralRedis),
		strings.ToLower(tapi.ResourceCodeRedis),
		strings.ToLower(tapi.ResourceSingularRedis):
		return tapi.ResourcePluralRedis, nil
	case strings.ToLower(tapi.ResourceKindMemcached),
		strings.ToLower(tapi.ResourcePluralMemcached),
		strings.ToLower(tapi.ResourceCodeMemcached),
		strings.ToLower(tapi.ResourceSingularMemcached):
		return tapi.ResourcePluralMemcached, nil
	case strings.ToLower(tapi.ResourceKindSnapshot),
		strings.ToLower(tapi.ResourcePluralSnapshot),
		strings.ToLower(tapi.ResourceCodeSnapshot),
		strings.ToLower(tapi.ResourceSingularSnapshot):
		return tapi.ResourcePluralSnapshot, nil
	case strings.ToLower(tapi.ResourceKindDormantDatabase),
		strings.ToLower(tapi.ResourcePluralDormantDatabase),
		strings.ToLower(tapi.ResourceCodeDormantDatabase),
		strings.ToLower(tapi.ResourceSingularDormantDatabase):
		return tapi.ResourcePluralDormantDatabase, nil
	case strings.ToLower(tapi.ResourceKindEtcd),
		strings.ToLower(tapi.ResourcePluralEtcd),
		strings.ToLower(tapi.ResourceCodeEtcd),
		strings.ToLower(tapi.ResourceSingularEtcd):
		return tapi.ResourcePluralEtcd, nil
	default:
		return "", fmt.Errorf(`kubedb doesn't support a resource type "%v"`, resource)
	}
}

func CheckSupportedResource(kind string) error {
	switch kind {
	case tapi.ResourceKindElasticsearch,
		tapi.ResourceKindPostgres,
		tapi.ResourceKindMySQL,
		tapi.ResourceKindMongoDB,
		tapi.ResourceKindRedis,
		tapi.ResourceKindMemcached,
		tapi.ResourceKindSnapshot,
		tapi.ResourceKindDormantDatabase,
		tapi.ResourceKindEtcd:
		return nil
	default:
		return fmt.Errorf(`kubedb doesn't support a resource type "%v"`, kind)
	}
}

func GetAllSupportedResources(f cmdutil.Factory) ([]string, error) {

	resources := []string{
		tapi.ResourcePluralElasticsearch + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourcePluralPostgres + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourcePluralMySQL + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourcePluralMongoDB + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourcePluralRedis + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourcePluralMemcached + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourcePluralSnapshot + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourcePluralDormantDatabase + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourcePluralEtcd + "." + tapi.SchemeGroupVersion.Group,
	}

	restConfig, err := f.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := crd_cs.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	availableResources := make([]string, 0)
	for _, val := range resources {
		_, err := clientset.CustomResourceDefinitions().Get(val, metav1.GetOptions{})
		if err != nil {
			if kerr.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		availableResources = append(availableResources, val)
	}

	return availableResources, nil
}

var ShortForms = map[string]string{
	tapi.ResourceCodeElasticsearch:   tapi.ResourcePluralElasticsearch,
	tapi.ResourceCodePostgres:        tapi.ResourcePluralPostgres,
	tapi.ResourceCodeMySQL:           tapi.ResourcePluralMySQL,
	tapi.ResourceCodeMongoDB:         tapi.ResourcePluralMongoDB,
	tapi.ResourceCodeRedis:           tapi.ResourcePluralRedis,
	tapi.ResourceCodeMemcached:       tapi.ResourcePluralMemcached,
	tapi.ResourceCodeSnapshot:        tapi.ResourcePluralSnapshot,
	tapi.ResourceCodeDormantDatabase: tapi.ResourcePluralDormantDatabase,
	tapi.ResourceCodeEtcd:            tapi.ResourcePluralDormantDatabase,
}

func ResourceShortFormFor(resource string) (string, bool) {
	resourceType, err := GetResourceType(resource)
	if err != nil {
		return "", false
	}

	var alias string
	exists := false
	for k, val := range ShortForms {
		if val == resourceType {
			alias = k
			exists = true
			break
		}
	}
	return alias, exists
}

func GetObjectData(obj runtime.Object) ([]byte, error) {
	return yaml.Marshal(obj)
}

func GetStructuredObject(obj runtime.Object) (runtime.Object, error) {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	data, err := GetObjectData(obj)
	if err != nil {
		return obj, err
	}
	return decoder.Decode(kind, data)
}

func checkChainKeyUnchanged(key string, mapData map[string]interface{}) bool {
	keys := strings.Split(key, ".")

	newKey := strings.Join(keys[1:], ".")
	if keys[0] == "*" {
		if len(keys) == 1 {
			return true
		}
		for _, val := range mapData {
			if !checkChainKeyUnchanged(newKey, val.(map[string]interface{})) {
				return false
			}
		}
	} else {
		val, ok := mapData[keys[0]]
		if !ok || len(keys) == 1 {
			return !ok
		}
		return checkChainKeyUnchanged(newKey, val.(map[string]interface{}))
	}

	return true
}

func RequireChainKeyUnchanged(key string) mergepatch.PreconditionFunc {
	return func(patch interface{}) bool {
		patchMap, ok := patch.(map[string]interface{})
		if !ok {
			fmt.Println("Invalid data")
			return true
		}
		return checkChainKeyUnchanged(key, patchMap)
	}
}

func GetPreconditionFunc(kind string) []mergepatch.PreconditionFunc {
	preconditions := []mergepatch.PreconditionFunc{
		mergepatch.RequireKeyUnchanged("apiVersion"),
		mergepatch.RequireKeyUnchanged("kind"),
		mergepatch.RequireMetadataKeyUnchanged("name"),
		mergepatch.RequireMetadataKeyUnchanged("namespace"),
		mergepatch.RequireKeyUnchanged("status"),
	}
	return preconditions
}

var PreconditionSpecField = map[string][]string{
	tapi.ResourceKindElasticsearch: {
		"spec.version",
		"spec.topology.*.prefix",
		"spec.enableSSL",
		"spec.certificateSecret",
		"spec.databaseSecret",
		"spec.storage",
		"spec.nodeSelector",
		"spec.init",
	},
	tapi.ResourceKindPostgres: {
		"spec.version",
		"spec.standby",
		"spec.streaming",
		"spec.archiver",
		"spec.databaseSecret",
		"spec.storage",
		"spec.nodeSelector",
		"spec.init",
	},
	tapi.ResourceKindMySQL: {
		"spec.version",
		"spec.storage",
		"spec.databaseSecret",
		"spec.nodeSelector",
		"spec.init",
	},
	tapi.ResourceKindMongoDB: {
		"spec.version",
		"spec.storage",
		"spec.databaseSecret",
		"spec.nodeSelector",
		"spec.init",
	},
	tapi.ResourceKindRedis: {
		"spec.version",
		"spec.storage",
		"spec.nodeSelector",
	},
	tapi.ResourceKindMemcached: {
		"spec.version",
		"spec.nodeSelector",
	},
	tapi.ResourceKindDormantDatabase: {
		"spec.origin",
	},
	tapi.ResourceKindEtcd: {
		"spec.version",
		"spec.storage",
		"spec.nodeSelector",
		"spec.init",
	},
}

func GetConditionalPreconditionFunc(kind string) []mergepatch.PreconditionFunc {
	preconditions := make([]mergepatch.PreconditionFunc, 0)

	if fields, found := PreconditionSpecField[kind]; found {
		for _, field := range fields {
			preconditions = append(preconditions,
				RequireChainKeyUnchanged(field),
			)
		}
	}

	return preconditions
}

func CheckResourceExists(client internalclientset.Interface, kind, name, namespace string) (bool, error) {
	var err error
	objectMata := metav1.ObjectMeta{
		Name: name,
	}
	var offshootLabels map[string]string
	switch kind {
	case tapi.ResourceKindElasticsearch:
		offshootLabels = tapi.Elasticsearch{ObjectMeta: objectMata}.OffshootLabels()
	case tapi.ResourceKindPostgres:
		offshootLabels = tapi.Postgres{ObjectMeta: objectMata}.OffshootLabels()
	case tapi.ResourceKindMySQL:
		offshootLabels = tapi.MySQL{ObjectMeta: objectMata}.OffshootLabels()
	case tapi.ResourceKindMongoDB:
		offshootLabels = tapi.MongoDB{ObjectMeta: objectMata}.OffshootLabels()
	case tapi.ResourceKindRedis:
		offshootLabels = tapi.Redis{ObjectMeta: objectMata}.OffshootLabels()
	case tapi.ResourceKindMemcached:
		offshootLabels = tapi.Memcached{ObjectMeta: objectMata}.OffshootLabels()
	case tapi.ResourceKindEtcd:
		offshootLabels = tapi.Etcd{ObjectMeta: objectMata}.OffshootLabels()
	}

	labelSelector := labels.SelectorFromSet(offshootLabels)

	statefulSets, err := client.Apps().StatefulSets(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return false, err
	}
	if len(statefulSets.Items) > 0 {
		return true, nil
	}

	deployments, err := client.Extensions().Deployments(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return false, err
	}
	if len(deployments.Items) > 0 {
		return true, nil
	}

	return true, nil
}

func IsPreconditionFailed(err error) bool {
	_, ok := err.(errPreconditionFailed)
	return ok
}

type errPreconditionFailed struct {
	message string
}

func newErrPreconditionFailed(target map[string]interface{}) errPreconditionFailed {
	s := fmt.Sprintf("precondition failed for: %v", target)
	return errPreconditionFailed{s}
}

func (err errPreconditionFailed) Error() string {
	return err.message
}

func CheckConditionalPrecondition(patchData []byte, fns ...mergepatch.PreconditionFunc) error {
	patch := make(map[string]interface{})
	if err := json.Unmarshal(patchData, &patch); err != nil {
		return err
	}
	for _, fn := range fns {
		if !fn(patch) {
			return newErrPreconditionFailed(patch)
		}
	}
	return nil
}
