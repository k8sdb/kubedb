package decoder

import (
	"fmt"

	"github.com/ghodss/yaml"
	tapi "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Decode(kind string, data []byte) (runtime.Object, error) {
	switch kind {
	case tapi.ResourceKindElasticsearch:
		var elasticsearch *tapi.Elasticsearch
		if err := yaml.Unmarshal(data, &elasticsearch); err != nil {
			return nil, err
		}
		return elasticsearch, nil
	case tapi.ResourceKindPostgres:
		var postgres *tapi.Postgres
		if err := yaml.Unmarshal(data, &postgres); err != nil {
			return nil, err
		}
		return postgres, nil
	case tapi.ResourceKindMySQL:
		var mysql *tapi.MySQL
		if err := yaml.Unmarshal(data, &mysql); err != nil {
			return nil, err
		}
		return mysql, nil
	case tapi.ResourceKindMongoDB:
		var mongodb *tapi.MongoDB
		if err := yaml.Unmarshal(data, &mongodb); err != nil {
			return nil, err
		}
		return mongodb, nil
	case tapi.ResourceKindRedis:
		var redis *tapi.Redis
		if err := yaml.Unmarshal(data, &redis); err != nil {
			return nil, err
		}
		return redis, nil
	case tapi.ResourceKindMemcached:
		var memcached *tapi.Memcached
		if err := yaml.Unmarshal(data, &memcached); err != nil {
			return nil, err
		}
		return memcached, nil
	case tapi.ResourceKindSnapshot:
		var snapshot *tapi.Snapshot
		if err := yaml.Unmarshal(data, &snapshot); err != nil {
			return nil, err
		}
		return snapshot, nil
	case tapi.ResourceKindDormantDatabase:
		var deletedDb *tapi.DormantDatabase
		if err := yaml.Unmarshal(data, &deletedDb); err != nil {
			return nil, err
		}
		return deletedDb, nil
	case tapi.ResourceKindEtcd:
		var etcd *tapi.Etcd
		if err := yaml.Unmarshal(data, &etcd); err != nil {
			return nil, err
		}
		return etcd, nil
	}

	return nil, fmt.Errorf(`Invalid kind: "%v"`, kind)
}
