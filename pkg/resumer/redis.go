/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resumer

import (
	"context"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2"
	dbutil "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	kmapi "kmodules.xyz/client-go/api/v1"
)

type RedisResumer struct {
	dbClient cs.KubedbV1alpha2Interface
}

func NewRedisResumer(clientConfig *rest.Config) (*RedisResumer, error) {
	k, err := cs.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return &RedisResumer{
		dbClient: k,
	}, nil
}

func (e *RedisResumer) Resume(name, namespace string) error {
	db, err := e.dbClient.Redises(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, err = dbutil.UpdateRedisStatus(context.TODO(), e.dbClient, db.ObjectMeta, func(status *api.RedisStatus) (types.UID, *api.RedisStatus) {
		status.Conditions = kmapi.RemoveCondition(status.Conditions, api.DatabasePaused)
		return db.UID, status
	}, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
