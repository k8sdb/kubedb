package validator

import (
	"fmt"

	"github.com/ghodss/yaml"
	tapi "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/cli/pkg/encoder"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

func ValidateDeletion(info *resource.Info) error {
	objByte, err := encoder.Encode(info.Object)
	if err != nil {
		return err
	}

	kind := info.Object.GetObjectKind().GroupVersionKind().Kind
	switch kind {
	case tapi.ResourceKindElasticsearch:
		var elasticsearch *tapi.Elasticsearch
		if err := yaml.Unmarshal(objByte, &elasticsearch); err != nil {
			return err
		}
		if elasticsearch.Spec.DoNotPause {
			return fmt.Errorf(`elasticsearch "%v" can't be paused. To continue delete, unset spec.doNotPause and retry`, elasticsearch.Name)
		}
	case tapi.ResourceKindPostgres:
		var postgres *tapi.Postgres
		if err := yaml.Unmarshal(objByte, &postgres); err != nil {
			return err
		}
		if postgres.Spec.DoNotPause {
			return fmt.Errorf(`postgres "%v" can't be paused. To continue delete, unset spec.doNotPause and retry`, postgres.Name)
		}
	case tapi.ResourceKindMySQL:
		var mysql *tapi.MySQL
		if err := yaml.Unmarshal(objByte, &mysql); err != nil {
			return err
		}
		if mysql.Spec.DoNotPause {
			return fmt.Errorf(`mysql "%v" can't be paused. To continue delete, unset spec.doNotPause and retry`, mysql.Name)
		}
	case tapi.ResourceKindMongoDB:
		var mongodb *tapi.MongoDB
		if err := yaml.Unmarshal(objByte, &mongodb); err != nil {
			return err
		}
		if mongodb.Spec.DoNotPause {
			return fmt.Errorf(`mongodb "%v" can't be paused. To continue delete, unset spec.doNotPause and retry`, mongodb.Name)
		}
	case tapi.ResourceKindRedis:
		var redis *tapi.Redis
		if err := yaml.Unmarshal(objByte, &redis); err != nil {
			return err
		}
		if redis.Spec.DoNotPause {
			return fmt.Errorf(`redis "%v" can't be paused. To continue delete, unset spec.doNotPause and retry`, redis.Name)
		}
	case tapi.ResourceKindMemcached:
		var memcached *tapi.Memcached
		if err := yaml.Unmarshal(objByte, &memcached); err != nil {
			return err
		}
		if memcached.Spec.DoNotPause {
			return fmt.Errorf(`memcached "%v" can't be paused. To continue delete, unset spec.doNotPause and retry`, memcached.Name)
		}
	case tapi.ResourceKindEtcd:
		var etcd *tapi.Etcd
		if err := yaml.Unmarshal(objByte, &etcd); err != nil {
			return err
		}
		if etcd.Spec.DoNotPause {
			return fmt.Errorf(`etcd "%v" can't be paused. To continue delete, unset spec.doNotPause and retry`, etcd.Name)
		}

	}
	return nil
}
